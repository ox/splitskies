package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		slog.Info(fmt.Sprintf("%s %s", r.Method, r.RequestURI))
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
