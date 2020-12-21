package dohboy

import (
	"errors"
	"fmt"

	"github.com/miekg/dns"
)

func rfc8482_canRejectForTypeAny(dnsQueryRequest *dns.Msg) bool {
	return dnsQueryRequest != nil &&
		len(dnsQueryRequest.Question) == 1 &&
		dnsQueryRequest.Question[0].Qtype == dns.TypeANY
}

func rfc8482_createResponse(dnsQueryRequest *dns.Msg) (*dns.Msg, error) {
	if !rfc8482_canRejectForTypeAny(dnsQueryRequest) {
		return nil, errors.New("Request is not a QTYPE=ANY request")
	}

	dnsQueryResponse := dns.Msg{}
	dnsQueryResponse.SetReply(dnsQueryRequest)
	hinfo, _ := dns.NewRR(fmt.Sprintf("%v 3600 IN HINFO \"RFC8482\" ", dnsQueryRequest.Question[0].Name))
	dnsQueryResponse.Answer = append(dnsQueryResponse.Answer, hinfo)
	return &dnsQueryResponse, nil
}
