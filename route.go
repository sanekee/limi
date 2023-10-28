package limi

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // ensure import
	"reflect"
	"strings"

	"github.com/sanekee/limi/internal/limi"
)

const (
	defaultHandlerPath = "handler"
)

type Handler any

// Router is a http router
type Router struct {
	path        string
	handlerPath string
	host        *limi.Node
	node        *limi.Node
	middlewares []func(http.Handler) http.Handler

	notFoundHandler         http.Handler
	methodNotAllowedHandler func(...string) http.Handler

	isSubRoute bool
}

// NewRouter returns Router with path preset
func NewRouter(path string, opts ...RouterOptions) (*Router, error) {
	r, err := newRouter(path, opts...)
	if err != nil {
		return nil, err
	}

	r.notFoundHandler = chainMiddlewares(r.notFoundHandler, r.middlewares...)

	h := r.methodNotAllowedHandler
	r.methodNotAllowedHandler = func(allowedMethods ...string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handler := h(allowedMethods...)
			handler = chainMiddlewares(handler, r.middlewares...)
			handler.ServeHTTP(w, req)
		})
	}
	return r, nil
}

type RouterOptions func(r *Router) error

// WithHosts set Router's hosts matcher
func WithHosts(hosts ...string) RouterOptions {
	return func(r *Router) error {
		if r.isSubRoute {
			return fmt.Errorf("setting subroute with host is unsupported %w", limi.ErrUnsupportedOperation)
		}
		if r.host == nil {
			n := &limi.Node{}
			r.host = n
		}
		for _, h := range hosts {
			if err := r.host.Insert(h, hostHandler{}); err != nil {
				return err
			}
		}
		return nil
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
		if r.isSubRoute {
			return fmt.Errorf("setting subroute with not found handler is not supported %w", limi.ErrUnsupportedOperation)
		}
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

// WithProfiler add golang profiler reports under /debug/pprof/
func WithProfiler() RouterOptions {
	return func(r *Router) error {
		if err := r.AddHTTPHandler("/debug/pprof/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// create a new req with /debug/pprof as root as expected by the installed pprof handler
			req = req.Clone(req.Context())
			path := "/debug/pprof/"
			paths := strings.SplitAfter(req.URL.Path, "/debug/pprof/")
			if len(paths) > 1 {
				path += removeLeadingSlash(strings.Join(paths[1:], "/"))
			}
			req.URL.Path = path
			http.DefaultServeMux.ServeHTTP(w, req)
		})); err != nil {
			return err
		}
		return nil
	}
}

// WithHandlerPath set Router's handler package base path to find handler's routing path
func WithHandlerPath(path string) RouterOptions {
	return func(r *Router) error {
		r.handlerPath = path
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
	if !limi.IsContextSet(ctx) {
		ctx = limi.NewContext(ctx)
		req = req.WithContext(ctx)

		if !r.IsSupportedHost(ctx, parseHost(req.Host)) {
			r.notFoundHandler.ServeHTTP(w, req)
		}
		path = req.URL.Path
	} else {
		path = limi.GetRoutingPath(ctx)
	}

	h, trail := r.lookup(ctx, path)
	if h == nil {
		r.notFoundHandler.ServeHTTP(w, req)
		return
	}

	if trail != "" {
		limi.SetRoutingPath(ctx, trail)
	}
	h.ServeHTTP(w, req)
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

	// if handler's path is not absolute path, prefix with package path after a hander/ subpath
	if !strings.HasPrefix(hPath, "/") {
		rtName := strings.ToLower(baseRT.Name())
		pkgPath := baseRT.PkgPath()
		pkgPath = removeTraillingSlash(findHandlerPath(r.handlerPath, pkgPath))

		if !pathDef {
			pkgName := packageName(pkgPath)
			hPath = pkgPath
			if pkgName != rtName {
				hPath += ensureLeadingSlash(rtName)
			}
		} else {
			hPath = pkgPath + ensureLeadingSlash(hPath)
		}
	}

	methods := httpMethodHandlers{
		m:                       make(map[string]http.Handler),
		methodNotAllowedHandler: r.methodNotAllowedHandler,
	}

	for i := 0; i < rt.NumMethod(); i++ {
		m := rt.Method(i)
		lName := strings.ToUpper(m.Name)
		if isHTTPHandlerProducer(m.Func) {
			vs := m.Func.Call([]reflect.Value{rv})
			v := vs[0]
			if v.Kind() == reflect.Func {
				methods.m[lName] = chainMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					v.Call([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(req)})
				}), mws...)
			}
		} else if isHTTPHandlerMethod(m.Func) {
			methods.m[lName] = chainMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				m.Func.Call([]reflect.Value{rv, reflect.ValueOf(w), reflect.ValueOf(req)})
			}), mws...)
		}
	}

	if len(methods.m) > 0 {
		return r.insertMethodHandler(hPath, methods)
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
	path = r.buildPath(path)
	return r.insertMethodHandler(path, httpMethodHandlers{
		m: map[string]http.Handler{
			method: chainMiddlewares(fn, mws...),
		},
		methodNotAllowedHandler: r.methodNotAllowedHandler,
	})

}

// AddHTTPHandler adds a catch all http handler with path
func (r *Router) AddHTTPHandler(path string, h http.Handler, mws ...func(http.Handler) http.Handler) error {
	path = r.buildPath(path)
	middlewares := append(r.middlewares, mws...)
	h = chainMiddlewares(h, middlewares...)
	return r.node.Insert(path, limi.HTTPHandler(h.ServeHTTP))
}

// AddRouter adds a sub router
func (r *Router) AddRouter(path string, opts ...RouterOptions) (*Router, error) {
	nr, err := newRouter(path)
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

	h := nr.methodNotAllowedHandler
	nr.methodNotAllowedHandler = func(allowedMethods ...string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handler := h(allowedMethods...)
			handler = chainMiddlewares(handler, nr.middlewares...)
			handler.ServeHTTP(w, req)
		})
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

// IsPartial implemens Node interface, returning true indicate partial match is return for futher matching
func (r *Router) IsPartial() bool {
	return true
}

func (r *Router) Merge(limi.Handle) bool {
	return false
}

func (r *Router) IsMethodAllowed(string) bool {
	return true
}

func newRouter(path string, opts ...RouterOptions) (*Router, error) {
	r := &Router{
		path:                    path,
		handlerPath:             defaultHandlerPath,
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

// insertMethodHandler inserts new handler
func (r *Router) insertMethodHandler(path string, h httpMethodHandlers) error {
	path = r.buildPath(path)
	handlers := buildHandlers(r.middlewares, h)

	return r.node.Insert(path, handlers)
}

// insertRouter inserts new router
func (r *Router) insertRouter(r1 *Router) error {
	return r.node.Insert(r.buildPath(r1.path), r1)
}

func (r *Router) lookup(ctx context.Context, path string) (limi.Handle, string) {
	router := r
	findPath := path
	for {
		h, trail := router.node.Lookup(ctx, findPath)

		// exact match
		if trail == "" {
			return h, ""
		}

		if !h.IsPartial() {
			return nil, trail
		}

		hr, ok := h.(*Router)
		if !ok {
			// catchall handler
			return h, trail
		}

		router, findPath = hr, trail
	}
}

func (r *Router) buildPath(path string) string {
	if r.isSubRoute {
		return path
	}
	return buildPath(r.path, path)
}

func buildHandlers(mws []func(http.Handler) http.Handler, hms httpMethodHandlers) httpMethodHandlers {
	handlers := hms

	for method, h := range hms.m {
		handlers.m[method] = chainMiddlewares(h, mws...)
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

func findHandlerPath(handlerPath, path string) string {
	arrPath := strings.SplitAfter(path, handlerPath)
	if len(arrPath) == 1 {
		return path
	}

	return strings.Join(arrPath[1:], "/")
}

func packageName(path string) string {
	idx := strings.LastIndexByte(removeTraillingSlash(path), '/')
	if idx < 0 {
		return path
	}

	return path[idx+1:]
}

func methodNotAllowedHandler(allowedMethods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, m := range allowedMethods {
			w.Header().Add("Allow", strings.ToUpper(m))
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

// Map of HTTP Handlers by Methods
type httpMethodHandlers struct {
	m                       map[string]http.Handler
	methodNotAllowedHandler func(...string) http.Handler
}

func (h httpMethodHandlers) keys() []string {
	var keys []string
	for k := range h.m {
		keys = append(keys, k)
	}
	return keys
}

func (h httpMethodHandlers) IsPartial() bool {
	return false
}

func (h httpMethodHandlers) IsMethodAllowed(method string) bool {
	_, ok := h.m[method]
	return ok
}

func (h httpMethodHandlers) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hdl, ok := h.m[req.Method]
	if !ok {
		h.methodNotAllowedHandler(h.keys()...).ServeHTTP(w, req)
		return
	}
	hdl.ServeHTTP(w, req)
}

func (h httpMethodHandlers) Merge(h1 limi.Handle) bool {
	hMap, ok := h1.(httpMethodHandlers)
	if !ok {
		return false
	}

	for method, hdl := range hMap.m {
		if _, ok := h.m[method]; ok {
			return false
		}
		h.m[method] = hdl
	}
	return true
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

func (h hostHandler) Merge(limi.Handle) bool {
	return false
}

func (h hostHandler) IsMethodAllowed(string) bool {
	return true
}

func (h hostHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {}
