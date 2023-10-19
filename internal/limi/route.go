package limi

import (
	"context"
	"net/http"
)

type Router struct {
	path string
	node *node
}

func NewRouter(path string) *Router {
	return &Router{path: path, node: &node{}}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if ctx == nil {
		ctx = context.TODO()
	}

	ctx = NewContext(ctx)

	path := req.URL.Path
	h := r.node.Lookup(ctx, path)
	if h == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hMap, ok := h.(map[string]http.Handler)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	handler, ok := hMap[req.Method]
	if !ok {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	handler.ServeHTTP(w, req)
}

func (r *Router) SetPath(path string) {
	r.path = path
}

func (r *Router) Insert(path string, handle any) error {
	return r.node.Insert(path, handle)
}
