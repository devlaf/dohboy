package doh

import (
	"regexp"

	"github.com/miekg/dns"
)

type Upstream struct {
	regex   *regexp.Regexp
	useDOH  bool
	address string
}

func (upstream *Upstream) ResolveIfMatched(message *dns.Msg) (*dns.Msg, error) {
	return nil, nil
}

func CreateUpstream(config UpstreamConfig) (*Upstream, error) {
	r, err := regexp.Compile(config.NameRegex)
	if err != nil {
		return nil, err
	}
	return &Upstream{
		regex:   r,
		useDOH:  config.UseDOH,
		address: config.Address,
	}, nil
}

func CreateDefaultUpstream() *Upstream {
	defaultUpstreamConfig := UpstreamConfig{
		NameRegex: "*",
		UseDOH:    true,
		Address:   "dns.google/dns-query",
	}
	defaultUpstream, _ := CreateUpstream(defaultUpstreamConfig)
	return defaultUpstream
}
