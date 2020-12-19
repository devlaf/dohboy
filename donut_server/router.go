package donut

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/miekg/dns"
)

type Router struct {
	rateLimiter    RateLimiter
	relay          *Relay
	terseResponses bool
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
	httpError := func(httpStatusCode int, err error) {
		if !router.terseResponses && err != nil {
			http.Error(response, fmt.Sprintf("%v: %v", http.StatusText(httpStatusCode), err), httpStatusCode)
		} else {
			http.Error(response, http.StatusText(httpStatusCode), httpStatusCode)
		}
	}

	if request.URL.Path != "/dns-query" {
		httpError(http.StatusNotFound, nil)
		return
	}

	if !router.rateLimiter.Please(router.rateLimiter.GetIP(request), request.URL.Query().Get("token")) {
		httpError(http.StatusTooManyRequests, nil)
		return
	}

	if request.Method != http.MethodGet && request.Method != http.MethodPost {
		httpError(http.StatusMethodNotAllowed, nil)
		return
	}

	if request.Method == http.MethodPost && request.Header.Get("Content-Type") != "application/dns-message" {
		httpError(http.StatusUnsupportedMediaType, nil)
		return
	}

	requestMsg, err := extractDNSMessage(request)
	if err != nil {
		httpError(http.StatusBadRequest, err)
		return
	}

	responseMsg, err := router.relay.ResolveDNSQuery(requestMsg)
	if err != nil {
		httpError(http.StatusInternalServerError, err)
		return
	}

	responseWireFormat, err := responseMsg.Pack()
	if err != nil {
		httpError(http.StatusInternalServerError, err)
		return
	}

	response.Header().Set("Content-Type", "application/dns-message")
	response.Write(responseWireFormat)
}

func CreateRouter(config Config) *http.ServeMux {
	rateLimiter := NewRateLimiter(config)
	relay := NewRelay(config)

	router := &Router{
		rateLimiter:    rateLimiter,
		relay:          relay,
		terseResponses: config.Development.TerseResponses,
	}

	mux := http.NewServeMux()
	mux.Handle("/", router)
	return mux
}
