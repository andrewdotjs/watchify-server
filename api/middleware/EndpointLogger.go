package middleware

import (
	"log"
	"net/http"
)

func EndpointLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("LOG : %v => %v", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
