package doh

import (
	"regexp"
	"time"

	"github.com/miekg/dns"
)

type Upstream struct {
	regex   *regexp.Regexp
	useDOH  bool
	address string
	timeout time.Duration
}

func (upstream *Upstream) ResolveIfMatched(message *dns.Msg) (*dns.Msg, error) {
	if !upstream.regex.MatchString(message.Question[0].Name) {
		return nil, nil
	}

	if upstream.useDOH {
		return resolveDnsOverHttps(upstream.address, upstream.timeout, message)
	} else {
		return resolveTraditionalDns(upstream.address, upstream.timeout, message)
	}
}

func resolveDnsOverHttps(address string, timeout time.Duration, message *dns.Msg) (*dns.Msg, error) {
	return nil, nil
}

func resolveTraditionalDns(address string, timeout time.Duration, message *dns.Msg) (*dns.Msg, error) {
	client := dns.Client{
		Net:     "tcp",
		Timeout: timeout,
	}

	resp, _, err := client.Exchange(message, address)
	return resp, err
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
		timeout: time.Duration(config.TimeoutMillis) * time.Millisecond,
	}, nil
}

func CreateDefaultUpstream() *Upstream {
	defaultUpstreamConfig := UpstreamConfig{
		NameRegex: ".*",
		UseDOH:    false,
		Address:   "8.8.8.8:53",
		//UseDOH:  true,
		//Address: "dns.google/dns-query",
		TimeoutMillis: 5000,
	}
	defaultUpstream, _ := CreateUpstream(defaultUpstreamConfig)
	return defaultUpstream
}
