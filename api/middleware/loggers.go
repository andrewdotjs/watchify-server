package middleware

import (
	"log"
	"net/http"
)

// Middleware to log the activity of the API's endpoints.
func LogEndpoint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("LOG : %-6v => %-30v", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
