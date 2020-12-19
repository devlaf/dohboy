package test

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/miekg/dns"
)

func testSimpleGet() {
	resp, err := http.Get("http://127.0.0.1:8080/dns-query?dns=AAABAAABAAAAAAAAA3d3dwdleGFtcGxlA2NvbQAAAQAB")
	if err != nil {
		log.Fatalf("ERR [testSimpleGet]: %v", err)
		return
	}
	checkResponse("testSimpleGet", resp)
}

func testSimplePost() {
	dnsMsg := []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x77, 0x77, 0x77, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x01, 0x00, 0x01}

	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/dns-query", bytes.NewReader(dnsMsg))
	if err != nil {
		log.Fatalf("ERR [testSimplePost] Can't create req: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("ERR [testSimplePost]: %v", err)
		return
	}
	checkResponse("testSimplePost", resp)
}

func checkResponse(testname string, resp *http.Response) {
	if resp.StatusCode != 200 {
		log.Fatalf("ERR [%v]: http status was not 200. Returned %v", testname, resp.Status)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ERR [%v]: %v", testname, err)
		return
	}

	dnsResp := new(dns.Msg)
	if dnsResp.Unpack(body) != nil {
		log.Fatalf("ERR [%v]: can't unpack dns msg: %v", testname, err)
		return
	}

	log.Printf("Response [%v]: %v", testname, dnsResp.Answer[0])
}

func Run() {
	// these examples are the dummy ones straight outta the rfc
	testSimpleGet()
	testSimplePost()
}
