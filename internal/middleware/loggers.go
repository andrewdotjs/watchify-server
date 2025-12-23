package middleware

import (
	"fmt"
	"net/http"

	"github.com/andrewdotjs/watchify-server/internal/logger"
)

// Middleware to log the activity of the API's endpoints.
func LogEndpoint(next http.Handler, log *logger.Logger, transactionId *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	  var message string = fmt.Sprintf("%-6v %-30v", r.Method, r.URL.Path)
		log.Info(*transactionId, message)
		next.ServeHTTP(w, r)
	})
}
