package middleware

import (
	"net/http"
	"time"
)

type Logger interface {
	Println(arg ...any)
}

func Log(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			rw := responseWriter{ResponseWriter: w}
			next.ServeHTTP(&rw, req)

			log.Println(req.RemoteAddr, req.Host, req.Method, req.RequestURI, req.Proto, rw.statusCode, rw.contentLength, time.Since(start))
		})
	}
}

type responseWriter struct {
	statusCode    int
	contentLength int
	http.ResponseWriter
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseWriter) Write(b []byte) (int, error) {
	w, err := r.ResponseWriter.Write(b)
	if err != nil {
		return w, err
	}
	r.contentLength += w
	return w, nil
}
