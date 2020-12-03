package doh

import (
	"fmt"
	"net/http"
)

func createRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "Hello")
	})

	return mux
}
