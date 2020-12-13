package doh

import (
	"fmt"
	"html"
	"net/http"
)

type Router struct {
	rateLimiter RateLimiter
}

func getAnyUserKey(request *http.Request) string {
	userKey, exists := request.URL.Query()["token"]
	if exists {
		return userKey[0]
	} else {
		return ""
	}
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

	if !router.rateLimiter.Please(request.RemoteAddr, getAnyUserKey(request)) {
		http.Error(response, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		return
	}

	fmt.Fprintf(response, "%q, %q", request.Method, html.EscapeString(request.URL.Path))
}

func CreateRouter(config Config) *http.ServeMux {
	rateLimiter := NewRateLimiter(config)
	router := &Router{
		rateLimiter: rateLimiter,
	}

	mux := http.NewServeMux()
	mux.Handle("/", router)
	return mux
}
