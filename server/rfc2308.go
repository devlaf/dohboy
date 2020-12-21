package dohboy

import (
	"errors"

	"github.com/miekg/dns"
)

func rfc2308_isNegativeResponse(dnsQueryResult *dns.Msg) bool {
	if dnsQueryResult == nil || !dnsQueryResult.Response {
		return false
	}

	return dnsQueryResult.Rcode == dns.RcodeNameError ||
		rfc2308_isNODATA(dnsQueryResult)
}

func rfc2308_isNODATA(dnsQueryResult *dns.Msg) bool {
	return dnsQueryResult != nil &&
		dnsQueryResult.Response &&
		dnsQueryResult.Rcode == dns.RcodeSuccess &&
		len(dnsQueryResult.Answer) == 0
}

func rfc2308_getSOARecord(dnsQueryResult *dns.Msg) (*dns.SOA, error) {
	for _, rr := range dnsQueryResult.Ns {
		header := rr.Header()
		if header != nil && header.Rrtype == dns.TypeSOA {
			if soaRR, ok := rr.(*dns.SOA); ok {
				return soaRR, nil
			}
		}
	}
	return nil, errors.New("No SOA record found.")
}

func rfc2308_getTTLForNegativeResponse(dnsQueryResult *dns.Msg) uint32 {
	soaRR, err := rfc2308_getSOARecord(dnsQueryResult)
	if err != nil {
		// RFC2308:
		// > Negative responses without SOA records SHOULD NOT be cached as there
		// > is no way to prevent the negative responses looping forever between a
		// > pair of servers even with a short TTL.
		return 0
	}
	return minOf(soaRR.Header().Ttl, soaRR.Minttl)
}

func minOf(vars ...uint32) uint32 {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}
