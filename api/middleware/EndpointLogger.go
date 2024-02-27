package middleware

import (
	"log"
	"net/http"
)

func EndpointLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("LOG : %-6v => %-30v", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
