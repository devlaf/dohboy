package donut

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/miekg/dns"
)

type upstream interface {
	resolveIfMatched(dnsQuery *dns.Msg) (bool, *dns.Msg, error) // (was_matched, resp_msg_if_matched, err)
}

func createUpstream(config UpstreamConfig) (upstream, error) {
	timeout := time.Duration(config.TimeoutMillis) * time.Millisecond

	regex, err := regexp.Compile(config.NameRegex)
	if err != nil {
		return nil, err
	}

	if config.UseDOH {
		return createDnsOverHttpsUpstream(regex, config.Address, timeout, config.HttpTransportConfig)
	} else {
		return createTraditionalUpstream(regex, config.Address, timeout), nil
	}
}

func createDefaultTraditionalUpstream() upstream {
	defaultUpstreamConfig := UpstreamConfig{
		NameRegex:     ".*",
		UseDOH:        false,
		Address:       "8.8.8.8:53",
		TimeoutMillis: 5000,
	}
	defaultUpstream, _ := createUpstream(defaultUpstreamConfig)
	return defaultUpstream
}

func createDefaultDnsOverHttpsUpstream() upstream {
	defaultUpstreamConfig := UpstreamConfig{
		NameRegex:     ".*",
		UseDOH:        true,
		Address:       "https://dns.google/dns-query",
		TimeoutMillis: 5000,
	}
	defaultUpstream, _ := createUpstream(defaultUpstreamConfig)
	return defaultUpstream
}

type traditionalUpstream struct {
	regex     *regexp.Regexp
	address   string
	tcpClient *dns.Client
	udpClient *dns.Client
}

func createTraditionalUpstream(regex *regexp.Regexp, address string, timeout time.Duration) upstream {
	tcpClient := &dns.Client{
		Net:     "tcp",
		Timeout: timeout,
	}

	udpClient := &dns.Client{
		Net:     "udp",
		Timeout: timeout,
	}

	return &traditionalUpstream{
		regex:     regex,
		address:   address,
		tcpClient: tcpClient,
		udpClient: udpClient,
	}
}

func (upstream *traditionalUpstream) resolveIfMatched(dnsQuery *dns.Msg) (bool, *dns.Msg, error) {
	if !upstream.regex.MatchString(dnsQuery.Question[0].Name) {
		return false, nil, nil
	}

	udpResp, _, err := upstream.udpClient.Exchange(dnsQuery, upstream.address)
	if err != nil {
		return true, nil, err
	}

	if !udpResp.Truncated {
		return true, udpResp, nil
	}

	tcpResp, _, err := upstream.tcpClient.Exchange(dnsQuery, upstream.address)
	return true, tcpResp, err
}

type dnsOverHttpsUpstream struct {
	regex      *regexp.Regexp
	address    string
	httpClient *http.Client
}

func createDnsOverHttpsUpstream(regex *regexp.Regexp, address string, timeout time.Duration, transportConfig HttpTransportConfig) (upstream, error) {
	validUrl, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, err
	}

	if validUrl.Scheme != "https" {
		return nil, fmt.Errorf("Address scheme for DOH upstream must be [https]. Provided: [%v].", validUrl.Scheme)
	}

	transport := &http.Transport{
		MaxConnsPerHost: transportConfig.MaxConnsPerHost,
		IdleConnTimeout: time.Duration(transportConfig.IdleConnTimeoutMillis) * time.Millisecond,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return &dnsOverHttpsUpstream{
		regex:      regex,
		address:    address,
		httpClient: httpClient,
	}, nil
}

func (upstream *dnsOverHttpsUpstream) resolveIfMatched(dnsQuery *dns.Msg) (bool, *dns.Msg, error) {
	if !upstream.regex.MatchString(dnsQuery.Question[0].Name) {
		return false, nil, nil
	}

	wireformat, err := dnsQuery.Pack()
	if err != nil {
		return true, nil, err
	}

	encodedQuery := base64.RawURLEncoding.EncodeToString(wireformat)
	uri := fmt.Sprintf("%v?dns=%v", upstream.address, encodedQuery)

	requestToUpstream, err := http.NewRequestWithContext(context.Background(), http.MethodGet, uri, nil)
	if err != nil {
		return true, nil, err
	}
	requestToUpstream.Header.Set("Accept", "application/dns-message")

	responseFromUpstream, err := upstream.httpClient.Do(requestToUpstream)
	if err != nil {
		return true, nil, err
	}
	defer responseFromUpstream.Body.Close()

	if responseFromUpstream.StatusCode != http.StatusOK {
		err := fmt.Errorf("HTTP status code returned from upstream was [%v: %v]",
			responseFromUpstream.StatusCode, http.StatusText(responseFromUpstream.StatusCode))
		return true, nil, err
	}

	body, err := ioutil.ReadAll(responseFromUpstream.Body)
	if err != nil {
		return true, nil, err
	}

	dnsResultFromUpstream := new(dns.Msg)
	if err := dnsResultFromUpstream.Unpack(body); err != nil {
		return true, nil, err
	}

	if dnsQuery.Id != dnsResultFromUpstream.Id {
		err := fmt.Errorf("DNS query ID mismatch: sent=%v received=%v", dnsQuery.Id, dnsResultFromUpstream.Id)
		return true, nil, err
	}

	return true, dnsResultFromUpstream, nil
}
