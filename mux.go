package limi

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/sanekee/limi/internal/limi"
)

type Handler interface{}

type HandlerType struct {
	Methods     map[string]http.Handler
	middlewares []Middleware
}

type Mux struct {
	host        string
	path        string
	handlers    map[string]HandlerType
	middlewares []Middleware

	routes *limi.Router
}

var defaultMux *Mux

func init() {
	defaultMux = newMux("/")
}

func newMux(path string) *Mux {
	return &Mux{
		path:     path,
		handlers: make(map[string]HandlerType),
		routes:   limi.NewRouter(path),
	}
}

func AddHandler(handler Handler, mws ...Middleware) {
	rt := reflect.TypeOf(handler)
	baseRT := rt
	if rt.Kind() == reflect.Pointer {
		baseRT = rt.Elem()
	}
	rv := reflect.ValueOf(handler)

	var hPath = ""
	var pathDef bool
	field, ok := baseRT.FieldByName("limi")
	if ok {
		if p, ok := field.Tag.Lookup("path"); ok {
			hPath = p
			pathDef = true
		}
	}

	if !strings.HasPrefix(hPath, "/") {
		pkgPath := baseRT.PkgPath()
		pkgPath = removeTraillingSlash(findHandlerPath(pkgPath))

		if !pathDef {
			hPath = pkgPath + ensureLeadingSlash(strings.ToLower(baseRT.Name()))
		} else {
			hPath = pkgPath + ensureLeadingSlash(hPath)
		}
	}

	methods := make(map[string]http.Handler)
	for i := 0; i < rt.NumMethod(); i++ {
		m := rt.Method(i)
		lName := strings.ToLower(m.Name)
		if isHTTPHandler(m.Func) {
			vs := m.Func.Call([]reflect.Value{rv})
			v := vs[0]
			if v.Kind() == reflect.Func {
				methods[lName] = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					v.Call([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(req)})
				})
			}
		}
	}
	if len(methods) > 0 {
		defaultMux.handlers[hPath] = HandlerType{
			Methods:     methods,
			middlewares: mws,
		}
	}
}

func Serve(opts ...MuxOption) http.Handler {
	return defaultMux.Serve(opts...)
}

func (r *Mux) AddHandler(path string, h HandlerType) {
	r.handlers[path] = h
}

func (r *Mux) Serve(opts ...MuxOption) http.Handler {
	for _, opt := range opts {
		opt(r)
	}

	for p, h := range r.handlers {
		p, h := p, h

		path := r.buildPath(p)
		mws := append(r.middlewares, h.middlewares...)
		handlers := buildHandlers(mws, h.Methods)

		r.routes.Insert(path, handlers)
	}

	return r.routes
}

func (r *Mux) Use(mw ...Middleware) {
	r.middlewares = append(r.middlewares, mw...)
}

func (r *Mux) buildPath(path string) string {
	return buildPath(r.host, r.path, path)
}

type MuxOption func(r *Mux)

func WithPath(path string) MuxOption {
	return func(r *Mux) {
		r.path = path
		r.routes.SetPath(path)
	}
}

func WithHost(host string) MuxOption {
	return func(r *Mux) {
		r.host = host
	}
}

type Middleware func(http.Handler) http.Handler

func WithMiddleWares(mw ...Middleware) MuxOption {
	return func(r *Mux) {
		r.Use(mw...)
	}
}

func isHTTPHandler(v reflect.Value) bool {
	if v.Kind() != reflect.Func {
		return false
	}

	vt := v.Type()
	if vt.NumOut() != 1 {
		return false
	}

	of := vt.Out(0)
	if of.Kind() != reflect.Func {
		return false
	}

	handlerFn := func(http.ResponseWriter, *http.Request) {}
	ht := reflect.TypeOf(handlerFn)

	chk := of.AssignableTo(ht)
	return chk
}

func buildHandlers(mws []Middleware, hvs map[string]http.Handler) map[string]http.Handler {
	handlers := make(map[string]http.Handler)

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
