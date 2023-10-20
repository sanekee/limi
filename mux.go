package limi

import (
	"net/http"
	"reflect"
	"strings"
)

type Handler interface{}

type HandlerType struct {
	// Map of HTTP Handlers by Methods
	Handlers map[string]http.Handler

	// List of middlewares
	Middlewares []func(http.Handler) http.Handler
}

// AddHandler adds handler with optional middlewares for this particular handler
// handler's path is automatically discovered from
//   - a 'path' tag in a field named 'limi'
//   - or assigned based on the package path after a 'handler' directory
//
// handler's methods that fulfill the http.HandlerFunc interface are automatically added
func AddHandler(handler Handler, mws ...func(http.Handler) http.Handler) {
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
			Handlers:    methods,
			Middlewares: mws,
		}
	}
}

// Mux returns the mux for manual manipulation
func Mux() *mux {
	return defaultMux
}

// Serve returns the HTTP handler for http server
func Serve(opts ...MuxOption) http.Handler {
	return defaultMux.Serve(opts...)
}

type MuxOption func(m *mux)

// WithPath sets the base path
func WithPath(path string) MuxOption {
	return func(m *mux) {
		m.router.SetPath(path)
	}
}

// WithPath sets the host
func WithHost(host string) MuxOption {
	return func(m *mux) {
		m.router.SetHost(host)
	}
}

type Middleware func(http.Handler) http.Handler

// WithPathMiddleWares sets the middlewares
func WithMiddleWares(mw ...func(http.Handler) http.Handler) MuxOption {
	return func(m *mux) {
		m.router.Use(mw...)
	}
}

type mux struct {
	handlers map[string]HandlerType
	router   *Router
}

var defaultMux *mux

// AddHandler adds HTTP handler by path manually
func (r *mux) AddHandler(path string, h HandlerType) {
	r.handlers[path] = h
}

// Serve returns the HTTP handler for http server
func (r *mux) Serve(opts ...MuxOption) http.Handler {
	for _, opt := range opts {
		opt(r)
	}

	for p, h := range r.handlers {
		r.router.Insert(p, h)
	}

	return r.router
}

func init() {
	defaultMux = newMux("/")
}

func newMux(path string) *mux {
	return &mux{
		handlers: make(map[string]HandlerType),
		router:   NewRouter(path),
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
