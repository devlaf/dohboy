package test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/miekg/dns"
)

func getFromMsg(address string, msg *dns.Msg) (*dns.Msg, *http.Response, error) {
	wireformat, err := msg.Pack()
	if err != nil {
		return nil, nil, err
	}

	encodedQuery := base64.RawURLEncoding.EncodeToString(wireformat)
	return getFromEncodedWireFormat(address, encodedQuery)
}

func getFromEncodedWireFormat(address string, qEncodedWireFormat string) (*dns.Msg, *http.Response, error) {
	resp, err := http.Get(fmt.Sprintf("%v?dns=%v", address, qEncodedWireFormat))
	if err != nil {
		return nil, nil, err
	}
	return parseResponse(resp)
}

func postFromMsg(address string, msg *dns.Msg) (*dns.Msg, *http.Response, error) {
	wireformat, err := msg.Pack()
	if err != nil {
		return nil, nil, err
	}

	return postFromWireFormat(address, wireformat)
}

func postFromWireFormat(address string, data []byte) (*dns.Msg, *http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("Couldn't create http request: %v", err)
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("Couldn't make http request: %v", err)
	}
	return parseResponse(resp)
}

func parseResponse(resp *http.Response) (*dns.Msg, *http.Response, error) {
	if resp.StatusCode != 200 {
		return nil, resp, fmt.Errorf("http status was not 200. Returned %v", resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, fmt.Errorf("Couldn't read response body: %v", err)
	}

	dnsResp := new(dns.Msg)
	if err := dnsResp.Unpack(body); err != nil {
		return nil, resp, fmt.Errorf("Couldn't unpack dns msg: %v", err)
	}

	return dnsResp, resp, nil
}

func createQueryMsg(host string, dnsType uint16, dnsClass uint16) *dns.Msg {
	// example params: "www.example.com.", dns.TypeMX, dns.ClassINET
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{host, dnsType, dnsClass}
	return m
}
