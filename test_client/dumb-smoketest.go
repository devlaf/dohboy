package test

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/miekg/dns"
)

const testAddress = "http://127.0.0.1:8080/dns-query"
const exampleDotComEncodedGetMsg = "AAABAAABAAAAAAAAA3d3dwdleGFtcGxlA2NvbQAAAQAB"

func testRFC8484GetExample() {
	msg, _, err := getFromEncodedWireFormat(testAddress, "AAABAAABAAAAAAAAA3d3dwdleGFtcGxlA2NvbQAAAQAB")
	if err != nil {
		log.Fatalf("ERR: testRFCGetExample -- %v", err)
	}
	if !strings.Contains(msg.Answer[0].String(), "www.example.com.") {
		log.Fatalf("ERR: testRFCGetExample resp missing expected -- %v", msg)
	}
}

func testRFC8484PostExample() {
	dnsMsg := []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x77, 0x77, 0x77, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x01, 0x00, 0x01}

	msg, _, err := postFromWireFormat(testAddress, dnsMsg)
	if err != nil {
		log.Fatalf("ERR: testRFCPostExample -- %v", err)
	}
	if !strings.Contains(msg.Answer[0].String(), "www.example.com.") {
		log.Fatalf("ERR: testRFCPostExample, resp missing expected -- %v", msg)
	}
}

func testNotFound() {
	msg, _, err := getFromEncodedWireFormat(testAddress, "AAABAAABAAAAAAAAAWE-NjJjaGFyYWN0ZXJsYWJlbC1tYWtlcy1iYXNlNjR1cmwtZGlzdGluY3QtZnJvbS1zdGFuZGFyZC1iYXNlNjQHZXhhbXBsZQNjb20AAAEAAQ")
	if err != nil {
		log.Fatalf("ERR: testNotFound -- %v", err)
	}
	if msg.Rcode != dns.RcodeNameError {
		log.Fatalf("ERR: testNotFound resp not of type NXDOMAIN -- %v", msg)
	}
}

func testHttpCaching() {
	msg, resp, err := getFromEncodedWireFormat(testAddress, exampleDotComEncodedGetMsg)
	if err != nil {
		log.Fatalf("ERR: testHttpCaching -- %v", err)
	}
	cc := strings.ReplaceAll(resp.Header.Get("Cache-Control"), "max-age=", "")
	if !strings.Contains(msg.Answer[0].String(), cc) {
		log.Fatalf("ERR: testHttpCaching cache control != ttl in A answer -- resp=%v, msg=%v", cc, msg.Answer[0].String())
	}
}

func testRateLimiting() {
	limited := false
	getRequest := fmt.Sprintf("%v?dns=%v", testAddress, exampleDotComEncodedGetMsg)
	for i := 0; i < 100; i++ {
		resp, err := http.Get(getRequest)
		if err == nil && resp.StatusCode == 429 {
			limited = true
			break
		}
	}
	if !limited {
		log.Fatalf("ERR: testRateLimiting -- this is jenky, but we should have been rate limited.")
	}
}

func testRateLimitingWithKey() {
	limited := false
	getRequest := fmt.Sprintf("%v?dns=%v&token=blee", testAddress, exampleDotComEncodedGetMsg)
	for i := 0; i < 100; i++ {
		resp, err := http.Get(getRequest)
		if err == nil && resp.StatusCode == 429 {
			limited = true
			break
		}
	}
	if limited {
		log.Fatalf("ERR: testRateLimiting -- this is jenky, but we should not have been rate limited.")
	}
}

func Run() {
	testRFC8484GetExample()
	testRFC8484PostExample()
	testNotFound()
	testHttpCaching()
	testRateLimiting()
	testRateLimitingWithKey()
}
