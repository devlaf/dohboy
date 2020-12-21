package donut

import (
	"math"

	"github.com/miekg/dns"
)

func getOverallTTL(dnsQueryResult *dns.Msg) uint32 {
	if dnsQueryResult == nil || !dnsQueryResult.Response || dnsQueryResult.Truncated {
		return 0
	}

	if rfc2308_isNegativeResponse(dnsQueryResult) {
		return rfc2308_getTTLForNegativeResponse(dnsQueryResult)
	}

	if dnsQueryResult.Rcode == dns.RcodeSuccess {
		answerFound, answerLowest := getSmallestTTL(dnsQueryResult.Answer)
		nsFound, nsLowest := getSmallestTTL(dnsQueryResult.Ns)
		extraFound, extraLowest := getSmallestTTL(dnsQueryResult.Extra)

		if answerFound || nsFound || extraFound {
			return minOf(answerLowest, nsLowest, extraLowest)
		}
	}

	return 0
}

func getSmallestTTL(rrs []dns.RR) (bool, uint32) {
	found := false
	smallest := uint32(math.MaxUint32)

	for _, rr := range rrs {
		header := rr.Header()
		if header != nil && header.Ttl < smallest {
			found = true
			smallest = header.Ttl
		}
	}

	return found, smallest
}

func overrideAnyLargeTTL(dnsQueryResult *dns.Msg, maxTTL uint32) {
	if dnsQueryResult == nil {
		return
	}

	for _, rrs := range [][]dns.RR{dnsQueryResult.Answer, dnsQueryResult.Ns, dnsQueryResult.Extra} {
		for _, rr := range rrs {
			header := rr.Header()
			if header != nil && header.Ttl >= maxTTL {
				header.Ttl = maxTTL
			}
		}
	}
}
