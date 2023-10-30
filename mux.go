package limi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sanekee/limi/internal/limi"
)

// mux is the router multiplexer.
type mux struct {
	routers         []*Router
	notFoundHandler http.Handler
}

// NewMux creates a new router multiplexer with a list of routers.
func NewMux(routers ...*Router) *mux {
	return &mux{
		routers: routers,
	}
}

// AddRouters adds a list of routers to the multiplexer
func (m *mux) AddRouters(routers ...*Router) {
	m.routers = append(m.routers, routers...)
}

// AddRouter creates a new router with optional list of RouterOptions and adds the new router to the multiplexer
func (m *mux) AddRouter(path string, opts ...RouterOptions) (*Router, error) {
	r, err := NewRouter(path, opts...)
	if err != nil {
		return nil, fmt.Errorf("error adding router %w", err)
	}

	m.routers = append(m.routers, r)
	return r, nil
}

// SetNotFoundHandler sets the not found handler.
func (m *mux) SetNotFoundHandler(h http.Handler) {
	m.notFoundHandler = h
}

// ServeHTTP handle HTTP request from net/http server.
func (m *mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if ctx == nil {
		ctx = context.TODO()
	}

	if !limi.IsContextSet(ctx) {
		ctx = limi.NewContext(ctx)
		req = req.WithContext(ctx)
	}

	var lastNotAllowedHandle limi.Handle
	for _, r := range m.routers {
		if r.IsSupportedHost(ctx, parseHost(req.Host)) {
			h, _ := r.lookup(ctx, req.URL.Path)
			if h != nil {
				if !h.IsMethodAllowed(req.Method) {
					lastNotAllowedHandle = h
					continue
				}
				h.ServeHTTP(w, req)
				return
			}
		}
		limi.ResetContext(ctx)
	}

	if lastNotAllowedHandle != nil {
		lastNotAllowedHandle.ServeHTTP(w, req)
		return
	}

	if m.notFoundHandler != nil {
		m.notFoundHandler.ServeHTTP(w, req)
		return
	}
	http.NotFound(w, req)
}
