package limi

import (
	"context"
	"net/http"
	"strings"

	"github.com/sanekee/limi/internal/limi"
)

// Router is a http router
type Router struct {
	host        string
	path        string
	node        *limi.Node
	middlewares []func(http.Handler) http.Handler

	notFoundHandler         http.Handler
	methodNotAllowedHandler func(...string) http.Handler
}

// NewRouter returns Router with path preset
func NewRouter(path string) *Router {
	return &Router{
		path:                    path,
		node:                    &limi.Node{},
		notFoundHandler:         http.NotFoundHandler(),
		methodNotAllowedHandler: methodNotAllowedHandler,
	}
}

// SetPath set Router's path
func (r *Router) SetPath(path string) {
	r.path = path
}

// SetPath set Router's host
func (r *Router) SetHost(host string) {
	r.host = host
}

// Use set Router's middlewares
func (r *Router) Use(mws ...func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, mws...)
}

// SetNotFoundHandler set the not found handler
func (r *Router) SetNotFoundHandler(h http.Handler) {
	r.notFoundHandler = h
}

// SetMethodNotAllowedHandler set the method not allowed handler with list of allowed methods
func (r *Router) SetMethodNotAllowedHandler(h func(...string) http.Handler) {
	r.methodNotAllowedHandler = h
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if ctx == nil {
		ctx = context.TODO()
	}

	ctx = limi.NewContext(ctx)

	path := req.URL.Path
	h := r.node.Lookup(ctx, path)
	if h == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hMap, ok := h.(handlerMap)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	handler, ok := hMap[req.Method]
	if !ok {
		r.methodNotAllowedHandler(hMap.keys()...).ServeHTTP(w, req)
		return
	}

	handler.ServeHTTP(w, req)
}

// Insert inserts new handler by path
func (r *Router) Insert(p string, h HandlerType) error {
	path := r.buildPath(p)
	mws := append(r.middlewares, h.Middlewares...)
	handlers := buildHandlers(mws, h.Handlers)
	return r.node.Insert(path, handlers)
}

func (r *Router) buildPath(path string) string {
	return buildPath(r.host, r.path, path)
}

func buildHandlers(mws []func(http.Handler) http.Handler, hvs handlerMap) handlerMap {
	handlers := make(handlerMap)

	for method, h := range hvs {
		if len(mws) == 0 {
			handlers[method] = h
			continue
		}

		h = mws[len(mws)-1](h)
		for i := len(mws) - 2; i >= 0; i-- {
			h = mws[i](h)
		}
		handlers[method] = h
	}
	return handlers
}

func buildPath(host string, parent string, path string) string {
	var p string
	if host != "" {
		if strings.HasPrefix(path, host) {
			return path
		}
		p = removeLeadingSlash(removeTraillingSlash(host))
	}

	if parent != "" && parent != "/" {
		p += removeTraillingSlash(ensureLeadingSlash(parent))
	}

	if path != "" {
		p += ensureLeadingSlash(path)
	}

	return p
}

func ensureLeadingSlash(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func removeTraillingSlash(path string) string {
	if strings.HasSuffix(path, "/") {
		return path[:len(path)-1]
	}
	return path
}

func ensureTrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}

func removeLeadingSlash(path string) string {
	if strings.HasPrefix(path, "/") {
		return path[1:]
	}
	return path
}

func findHandlerPath(path string) string {
	arrPath := strings.SplitAfter(path, "/handler")
	if len(arrPath) == 1 {
		return path
	}

	return strings.Join(arrPath[1:], "/")
}

func methodNotAllowedHandler(allowedMethods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, m := range allowedMethods {
			w.Header().Add("Allow", strings.ToUpper(m))
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

type handlerMap map[string]http.Handler

func (h handlerMap) keys() []string {
	var keys []string
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}
