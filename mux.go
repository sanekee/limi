package limi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sanekee/limi/internal/limi"
)

var defaultMux mux

type mux struct {
	routers         []*Router
	notFoundHandler http.Handler
}

func AddRouter(path string, opts ...RouterOptions) (*Router, error) {
	return defaultMux.addRouter(path, opts...)
}

func SetNotFoundHandler(h http.Handler) {
	defaultMux.setNotFoundHandler(h)
}

func ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defaultMux.ServeHTTP(w, req)
}

func (m *mux) addRouter(path string, opts ...RouterOptions) (*Router, error) {
	r, err := NewRouter(path, opts...)
	if err != nil {
		return nil, fmt.Errorf("error adding router %w", err)
	}

	m.routers = append(m.routers, r)
	return r, nil
}

func (m *mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if ctx == nil {
		ctx = context.TODO()
	}

	if !limi.IsContextSet(ctx) {
		ctx = limi.NewContext(ctx)
		req = req.WithContext(ctx)
	}

	for _, r := range m.routers {
		if r.IsSupportedHost(ctx, parseHost(req.Host)) {
			h, _, err := r.lookup(ctx, req.Method, req.URL.Path)
			if err == nil && h != nil {
				h.ServeHTTP(w, req)
				return
			}
		}
		limi.Reset(ctx)
	}

	if m.notFoundHandler != nil {
		m.notFoundHandler.ServeHTTP(w, req)
		return
	}
	http.NotFound(w, req)
}

func (m *mux) setNotFoundHandler(h http.Handler) {
	m.notFoundHandler = h
}
