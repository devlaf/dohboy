package donut

import (
	"errors"
	"fmt"
	"log"

	"github.com/miekg/dns"
)

type relay struct {
	upstreamMatrix []upstream
}

func (relay *relay) resolveDNSQuery(requestMsg *dns.Msg) (*dns.Msg, error) {
	if len(requestMsg.Question) != 1 {
		responseMsg := dns.Msg{}
		return responseMsg.SetRcodeFormatError(requestMsg), nil
	}

	// RFC8482
	if requestMsg.Question[0].Qtype == dns.TypeANY {
		responseMsg := dns.Msg{}
		responseMsg.SetReply(requestMsg)
		hinfo, _ := dns.NewRR(fmt.Sprintf("%v 3600 IN HINFO \"RFC8482\" ", requestMsg.Question[0].Name))
		responseMsg.Answer = append(responseMsg.Answer, hinfo)
		return &responseMsg, nil
	}

	for _, upstream := range relay.upstreamMatrix {
		matched, resp, err := upstream.resolveIfMatched(requestMsg)
		if matched {
			return resp, err
		}
	}

	return nil, errors.New("No matched upstreams found.")
}

func newRelay(config *Config) *relay {
	upstreamMatrix := make([]upstream, 0, len(config.Upstream.Custom)+1)

	for _, config := range config.Upstream.Custom {
		us, err := createUpstream(config)
		if err != nil {
			log.Printf("ERR: Configured upstream for pattern [%v] is bad. The upstream won't be included.", config.NameRegex)
			log.Printf("ERR: %v", err)
			continue
		}
		upstreamMatrix = append(upstreamMatrix, us)
	}

	upstreamMatrix = append(upstreamMatrix, createDefaultDnsOverHttpsUpstream())

	return &relay{
		upstreamMatrix: upstreamMatrix,
	}
}
