package doh

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/miekg/dns"
)

type Router struct {
	rateLimiter       RateLimiter
	terseResponses    bool
	fetchIPFromHeader bool
}

func getIP(request *http.Request, allowIPFromHeader bool) string {
	if allowIPFromHeader {
		if forwarded := request.Header.Get("X-FORWARDED-FOR"); forwarded != "" {
			return forwarded
		}
		if real := request.Header.Get("X-Real-IP"); real != "" {
			return real
		}
	}
	return request.RemoteAddr
}

func extractDNSWireFormat(request *http.Request) ([]byte, error) {
	if request.Method == http.MethodGet {
		dnsQuery := request.URL.Query().Get("dns")
		if dnsQuery == "" {
			return nil, errors.New("No dns query param supplied.")
		}

		return base64.RawURLEncoding.DecodeString(dnsQuery)
	}

	if request.Method == http.MethodPost {
		defer request.Body.Close()
		return ioutil.ReadAll(request.Body)
	}

	return *new([]byte), nil
}

func extractDNSMessage(request *http.Request) (*dns.Msg, error) {
	wireFormat, err := extractDNSWireFormat(request)
	if err != nil {
		return nil, err
	}

	retval := new(dns.Msg)
	return retval, retval.Unpack(wireFormat)
}

func (router *Router) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/dns-query" {
		http.Error(response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if request.Method != http.MethodGet && request.Method != http.MethodPost {
		http.Error(response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if request.Method == http.MethodPost && request.Header.Get("Content-Type") != "application/dns-message" {
		http.Error(response, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	if !router.rateLimiter.Please(getIP(request, router.fetchIPFromHeader), request.URL.Query().Get("token")) {
		http.Error(response, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		return
	}

	msg, err := extractDNSMessage(request)
	if err != nil {
		if !router.terseResponses {
			http.Error(response, fmt.Sprintf("%v: %v", http.StatusText(http.StatusBadRequest), err), http.StatusBadRequest)
		} else {
			http.Error(response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		return
	}

	fmt.Fprintf(response, "%q", msg.Question[0].Name)

}

func CreateRouter(config Config) *http.ServeMux {
	rateLimiter := NewRateLimiter(config)
	router := &Router{
		rateLimiter:       rateLimiter,
		terseResponses:    config.Development.TerseResponses,
		fetchIPFromHeader: config.IPRateLimit.FetchIPFromHeaders,
	}

	mux := http.NewServeMux()
	mux.Handle("/", router)
	return mux
}
