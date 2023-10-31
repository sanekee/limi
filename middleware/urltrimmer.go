package middleware

import (
	"net/http"
	"strings"
)

// URLTrimmer trim req.URL.Path url prefix.
func URLTrimmer(prefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)

			next.ServeHTTP(w, req)
		})
	}
}
