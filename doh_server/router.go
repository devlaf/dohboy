package doh

import (
	"fmt"
	"html"
	"net/http"
)

func createRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/dns-query" {
			http.NotFound(response, request)
			return
		}

		if request.Method == "GET" {
			fmt.Fprintf(response, "GET, %q", html.EscapeString(request.URL.Path))
		} else if request.Method == "POST" {
			fmt.Fprintf(response, "POST, %q", html.EscapeString(request.URL.Path))
		} else {
			http.Error(response, "Invalid request.", 405)
		}
	})

	return mux
}
