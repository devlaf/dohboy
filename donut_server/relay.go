package donut

import (
	"errors"
	"log"

	"github.com/miekg/dns"
)

type relay struct {
	upstreamMatrix     []upstream
	maximumTTLOverride uint32
}

func (relay *relay) resolveDNSQuery(requestMsg *dns.Msg) (*dns.Msg, error) {
	if len(requestMsg.Question) != 1 {
		// Format technically allows this (RFC1305) but in practice nobody seems to
		// support it, including probably anything upstream of this relay. Specifics
		// required to implement multiple questions are not universally defined.
		responseMsg := dns.Msg{}
		return responseMsg.SetRcodeFormatError(requestMsg), nil
	}

	if rfc8482_canRejectForTypeAny(requestMsg) {
		return rfc8482_createResponse(requestMsg)
	}

	for _, upstream := range relay.upstreamMatrix {
		matched, resp, err := upstream.resolveIfMatched(requestMsg)
		if matched {
			if err != nil && relay.maximumTTLOverride != 0 {
				overrideAnyLargeTTL(resp, relay.maximumTTLOverride)
			}

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
		upstreamMatrix:     upstreamMatrix,
		maximumTTLOverride: config.Upstream.MaximumTTLOverrideSeconds,
	}
}
