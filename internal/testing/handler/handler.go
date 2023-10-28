package handler

import "net/http"

func NewHandler(status int, headers map[string]string, body []byte) http.Handler {
	return http.HandlerFunc(NewHandlerFunc(status, headers, body))
}

func NewHandlerFunc(status int, headers map[string]string, body []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}

		w.WriteHeader(status)

		if len(body) > 0 {
			w.Write(body) //nolint:errcheck
		}
	}
}
