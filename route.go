package limi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"reflect"
	"strings"

	"github.com/sanekee/limi/internal/limi"
)

type Handler any

type httpMethodHandler struct {
	// Map of HTTP Handlers by Methods
	handlers map[string]http.Handler

	// List of middlewares
	middlewares []func(http.Handler) http.Handler
}

// Router is a http router
type Router struct {
	path        string
	host        *limi.Node
	node        *limi.Node
	middlewares []func(http.Handler) http.Handler

	notFoundHandler         http.Handler
	methodNotAllowedHandler func(...string) http.Handler

	isSubRoute bool
}

// NewRouter returns Router with path preset
func NewRouter(path string, opts ...RouterOptions) (*Router, error) {
	r := &Router{
		path:                    path,
		node:                    &limi.Node{},
		notFoundHandler:         http.NotFoundHandler(),
		methodNotAllowedHandler: methodNotAllowedHandler,
	}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("error creating router %w", err)
		}
	}
	return r, nil
}

type RouterOptions func(r *Router) error

// WithHost set Router's host
func WithHost(host string) RouterOptions {
	return func(r *Router) error {
		if r.isSubRoute {
			return errors.New(limi.ErrUnsupportedOperation)
		}
		if r.host == nil {
			n := &limi.Node{}
			r.host = n
		}
		return r.host.Insert(host, hostHandler{})
	}
}

// WithMiddlewares set Router's middlewares
func WithMiddlewares(mws ...func(http.Handler) http.Handler) RouterOptions {
	return func(r *Router) error {
		r.middlewares = append(r.middlewares, mws...)
		return nil
	}
}

// WithNotFoundHandler set the not found handler
func WithNotFoundHandler(h http.Handler) RouterOptions {
	return func(r *Router) error {
		r.notFoundHandler = h
		return nil
	}
}

// WithMethodNotAllowedHandler set the method not allowed handler with list of allowed methods
func WithMethodNotAllowedHandler(h func(...string) http.Handler) RouterOptions {
	return func(r *Router) error {
		r.methodNotAllowedHandler = h
		return nil
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if ctx == nil {
		ctx = context.TODO()
	}

	var path string
	if !r.isSubRoute {
		ctx = limi.NewContext(ctx)
		req = req.WithContext(ctx)

		host := parseHost(req.URL.Host)
		if !r.IsSupportedHost(ctx, host) {
			r.notFoundHandler.ServeHTTP(w, req)
		}
		path = req.URL.Path
	} else {
		path = limi.GetRoutingPath(ctx)
	}

	h, trail := r.node.Lookup(ctx, path)
	if h == nil {
		r.notFoundHandler.ServeHTTP(w, req)
		return
	}

	var handler http.Handler
	switch hh := h.(type) {
	case handlerMap:
		hdl, ok := hh[req.Method]
		if !ok {
			r.methodNotAllowedHandler(hh.keys()...).ServeHTTP(w, req)
			return
		}
		handler = hdl
	case *Router:
		limi.SetRoutingPath(ctx, trail)
		handler = hh
	case limi.HTTPHandler:
		handler = http.HandlerFunc(hh)
	}

	handler.ServeHTTP(w, req)
}

// AddHandler adds handler with optional middleware
// handler's path is automatically discovered from
//   - a 'path' tag in a field named 'limi'
//   - or assigned based on the package path after a 'handler' directory
//
// handler's methods fulfill the http.HandlerFunc interface are automatically added
func (r *Router) AddHandler(handler Handler, mws ...func(http.Handler) http.Handler) error {
	rt := reflect.TypeOf(handler)
	baseRT := rt
	if rt.Kind() == reflect.Pointer {
		baseRT = rt.Elem()
	}
	rv := reflect.ValueOf(handler)

	var hPath = ""
	var pathDef bool
	if baseRT.Kind() == reflect.Struct {
		field, ok := baseRT.FieldByName("limi")
		if ok {
			if p, ok := field.Tag.Lookup("path"); ok {
				hPath = p
				pathDef = true
			}
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
		lName := strings.ToUpper(m.Name)
		if isHTTPHandlerProducer(m.Func) {
			vs := m.Func.Call([]reflect.Value{rv})
			v := vs[0]
			if v.Kind() == reflect.Func {
				methods[lName] = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					v.Call([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(req)})
				})
			}
		} else if isHTTPHandlerMethod(m.Func) {
			methods[lName] = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				m.Func.Call([]reflect.Value{rv, reflect.ValueOf(w), reflect.ValueOf(req)})
			})
		}
	}

	if len(methods) > 0 {
		return r.insertMethodHandler(hPath, httpMethodHandler{
			handlers:    methods,
			middlewares: mws,
		})
	}
	return nil
}

// AddHandlers adds multiple handlers with the same middlewares
func (r *Router) AddHandlers(handlers []Handler, mws ...func(http.Handler) http.Handler) error {
	for _, h := range handlers {
		if err := r.AddHandler(h, mws...); err != nil {
			return err
		}
	}
	return nil
}

// AddHandlerFunc adds http handler with path and method
func (r *Router) AddHandlerFunc(path string, method string, fn http.HandlerFunc, mws ...func(http.Handler) http.Handler) error {
	return r.insertMethodHandler(path, httpMethodHandler{
		handlers: map[string]http.Handler{
			method: fn,
		},
		middlewares: mws,
	})
}

// AddHTTPHandler adds a catch all http handler with path
func (r *Router) AddHTTPHandler(path string, h http.Handler, mws ...func(http.Handler) http.Handler) error {
	path = r.buildPath(path)
	h = chainMiddlewares(h, r.middlewares...)
	return r.node.Insert(path, limi.HTTPHandler(h.ServeHTTP))
}

// AddRouter adds a sub router
func (r *Router) AddRouter(path string, opts ...RouterOptions) (*Router, error) {
	nr, err := NewRouter(r.buildPath(path))
	if err != nil {
		return nil, err
	}

	nr.isSubRoute = true

	fn := WithMiddlewares(r.middlewares...)
	if err := fn(nr); err != nil {
		return nil, fmt.Errorf("error applying middleware option to sub route %w", err)
	}

	for _, opt := range opts {
		if err := opt(nr); err != nil {
			return nil, fmt.Errorf("error applying router option to sub route %w", err)
		}
	}

	if err := r.insertRouter(nr); err != nil {
		return nil, fmt.Errorf("error inserting router %w", err)
	}

	return nr, nil
}

// IsSupportedHost match host with supported host
func (r *Router) IsSupportedHost(ctx context.Context, host string) bool {
	if r.host == nil {
		return true
	}

	h, _ := r.host.Lookup(ctx, host)
	return h != nil
}

func WithProfiler() RouterOptions {
	return func(r *Router) error {
		if err := r.AddHTTPHandler("/debug/pprof", pprofHandler{}); err != nil {
			return err
		}
		return nil
	}
}

// IsPartial implemens Node interface, returning true indicate partial match is return for futher matching
func (r *Router) IsPartial() bool {
	return true
}

// insertMethodHandler inserts new handler
func (r *Router) insertMethodHandler(path string, h httpMethodHandler) error {
	if !r.isSubRoute {
		path = r.buildPath(path)
	}
	mws := append(r.middlewares, h.middlewares...)
	handlers := buildHandlers(mws, h.handlers)
	return r.node.Insert(path, handlers)
}

// insertRouter inserts new router
func (r *Router) insertRouter(r1 *Router) error {
	return r.node.Insert(r1.path, r1)
}

func (r *Router) buildPath(path string) string {
	return buildPath(r.path, path)
}

func buildHandlers(mws []func(http.Handler) http.Handler, hvs handlerMap) handlerMap {
	handlers := make(handlerMap)

	for method, h := range hvs {
		handlers[method] = chainMiddlewares(h, mws...)
	}
	return handlers
}

func chainMiddlewares(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	if len(mws) == 0 {
		return h
	}

	h = mws[len(mws)-1](h)
	for i := len(mws) - 2; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func buildPath(parent string, path string) string {
	var p string
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

func (h handlerMap) IsPartial() bool {
	return false
}

// isHTTPHandlerProducer check if the function produces a http.HandlerFunc
func isHTTPHandlerProducer(v reflect.Value) bool {
	if v.Kind() != reflect.Func {
		return false
	}

	vt := v.Type()
	if vt.NumOut() != 1 {
		return false
	}

	of := vt.Out(0)
	if of.Kind() == reflect.Func {
		ht := reflect.TypeOf(func(http.ResponseWriter, *http.Request) {})
		return of.AssignableTo(ht)
	}

	if of.Kind() == reflect.Interface {
		ht := reflect.TypeOf(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		return ht.Implements(of)
	}
	return false

}

// isHTTPHandlerMethod check if the method is a http.HandlerFunc
func isHTTPHandlerMethod(v reflect.Value) bool {
	if v.Kind() != reflect.Func {
		return false
	}

	vt := v.Type()

	fIdx := vt.NumIn() - 2
	if fIdx < 0 || fIdx > 1 {
		return false
	}

	ht := reflect.TypeOf(func(http.ResponseWriter, *http.Request) {})
	i1 := vt.In(fIdx)
	i2 := vt.In(fIdx + 1)

	c1 := ht.In(0)
	c2 := ht.In(1)

	return i1.Kind() == c1.Kind() &&
		i2.Kind() == c2.Kind() &&
		i1.Implements(c1) &&
		i2.AssignableTo(c2)
}

func parseHost(str string) string {
	arr := strings.Split(str, ":")

	if len(arr) == 0 {
		return ""
	}

	return arr[0]
}

type hostHandler struct{}

func (h hostHandler) IsPartial() bool {
	return false
}

type pprofHandler struct{}

func (p pprofHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pprof.Index(w, req)
}
