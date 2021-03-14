package logger

import (
	"log"
	"net/http"
)

func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("new request: %s %s", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}
