package doh

import (
	"regexp"
	"time"

	"github.com/miekg/dns"
)

type Upstream interface {
	ResolveIfMatched(message *dns.Msg) (bool, *dns.Msg, error) // (was_matched, resp_msg_if_matched, err)
}

func CreateUpstream(config UpstreamConfig) (Upstream, error) {
	timeout := time.Duration(config.TimeoutMillis) * time.Millisecond

	regex, err := regexp.Compile(config.NameRegex)
	if err != nil {
		return nil, err
	}

	if config.UseDOH {
		return createDnsOverHttpsUpstream(regex, config.Address, timeout), nil
	} else {
		return createTraditionalUpstream(regex, config.Address, timeout), nil
	}
}

func CreateDefaultTraditionalUpstream() Upstream {
	defaultUpstreamConfig := UpstreamConfig{
		NameRegex:     ".*",
		UseDOH:        false,
		Address:       "8.8.8.8:53",
		TimeoutMillis: 5000,
	}
	defaultUpstream, _ := CreateUpstream(defaultUpstreamConfig)
	return defaultUpstream
}

func CreateDefaultDnsOverHttpsUpstream() Upstream {
	defaultUpstreamConfig := UpstreamConfig{
		NameRegex:     ".*",
		UseDOH:        true,
		Address:       "dns.google/dns-query",
		TimeoutMillis: 5000,
	}
	defaultUpstream, _ := CreateUpstream(defaultUpstreamConfig)
	return defaultUpstream
}

type TraditionalUpstream struct {
	regex     *regexp.Regexp
	address   string
	tcpClient *dns.Client
	udpClient *dns.Client
}

func createTraditionalUpstream(regex *regexp.Regexp, address string, timeout time.Duration) Upstream {
	tcpClient := &dns.Client{
		Net:     "tcp",
		Timeout: timeout,
	}

	udpClient := &dns.Client{
		Net:     "udp",
		Timeout: timeout,
	}

	return &TraditionalUpstream{
		regex:     regex,
		address:   address,
		tcpClient: tcpClient,
		udpClient: udpClient,
	}
}

func (upstream *TraditionalUpstream) ResolveIfMatched(message *dns.Msg) (bool, *dns.Msg, error) {
	if !upstream.regex.MatchString(message.Question[0].Name) {
		return false, nil, nil
	}

	udpResp, _, err := upstream.udpClient.Exchange(message, upstream.address)
	if err != nil {
		return true, nil, err
	}

	if !udpResp.Truncated {
		return true, udpResp, nil
	}

	tcpResp, _, err := upstream.tcpClient.Exchange(message, upstream.address)
	return true, tcpResp, err
}

type DnsOverHttpsUpstream struct {
	regex   *regexp.Regexp
	address string
	timeout time.Duration
}

func createDnsOverHttpsUpstream(regex *regexp.Regexp, address string, timeout time.Duration) Upstream {
	return &DnsOverHttpsUpstream{
		regex:   regex,
		address: address,
		timeout: timeout,
	}
}

func (upstream *DnsOverHttpsUpstream) ResolveIfMatched(message *dns.Msg) (bool, *dns.Msg, error) {
	if !upstream.regex.MatchString(message.Question[0].Name) {
		return false, nil, nil
	}

	return true, nil, nil
}
