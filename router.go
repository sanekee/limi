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
	defaultHandlerPath     = "handler"
	defaultIndexStructName = "index"
)

type Handler any

// Router is a http router. Router matches host and path with an internal Radix Tree with O(k) complexity.
//
// # Pattern
//
// Host & path support three types of pattern matching.
// - String - matches the string as is.
// - Regexp - matches the string with Regular Expression, and sets the URLParam with the matched value.
// - Label  - matches the string with wildcard, and sets the URLParam with the matched value.
//
// # Example
//
// *Host*
//
// "static.domain.com",                 // matches the host static.domain.com
// "{apiVer:v[0-9]+}.api.domain.com",   // matches hosts v1.api.domain.com, v2.api.domain.com ... and sets URLParams["apiVer"] = value
// "{subdomain}.domain.com",            // matches hosts subdomain1.domain.com, subdomain2.domain.com ... and sets URLParams["subdomain"] = value
//
// *Path*
//
// "/blog/top" ..              // matches the exact path /blog/top
// "/blog/{id:[0-9]+}" ..      // matches paths /blog/1, /blog/2 ..., sets URLParams["id"] = <value>
// "/blog/{slug}" ..           // matches paths /blog/cool-article-1, /blog/cool-article-2 ..., sets URLParam["slug"] = <value>
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

// NewRouter returns Router with path and list of RouterOptions.
func NewRouter(path string, opts ...RouterOptions) (*Router, error) {
	r, err := newRouter(path, opts...)
	if err != nil {
		return nil, err
	}

	r.notFoundHandler = attachMiddlewares(r.notFoundHandler, r.middlewares...)

	h := r.methodNotAllowedHandler
	r.methodNotAllowedHandler = func(allowedMethods ...string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handler := h(allowedMethods...)
			handler = attachMiddlewares(handler, r.middlewares...)
			handler.ServeHTTP(w, req)
		})
	}
	return r, nil
}

type RouterOptions func(r *Router) error

// WithHosts set Router's hosts matcher.
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

// WithMiddlewares set Router's middlewares.
func WithMiddlewares(mws ...func(http.Handler) http.Handler) RouterOptions {
	return func(r *Router) error {
		r.middlewares = append(r.middlewares, mws...)
		return nil
	}
}

// WithNotFoundHandler set the not found handler.
func WithNotFoundHandler(h http.Handler) RouterOptions {
	return func(r *Router) error {
		if r.isSubRoute {
			return fmt.Errorf("setting subroute with not found handler is not supported %w", limi.ErrUnsupportedOperation)
		}
		r.notFoundHandler = h
		return nil
	}
}

// WithMethodNotAllowedHandler set the method not allowed handler with list of allowed methods.
func WithMethodNotAllowedHandler(h func(...string) http.Handler) RouterOptions {
	return func(r *Router) error {
		r.methodNotAllowedHandler = h
		return nil
	}
}

// WithProfiler add golang profiler reports under /debug/pprof/.
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

// WithHandlerPath set Router's handler package base path to find handler's routing path.
func WithHandlerPath(path string) RouterOptions {
	return func(r *Router) error {
		r.handlerPath = path
		return nil
	}
}

// ServeHTTP handles http request from net/http server.
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

		limi.SetQueries(ctx, req.URL.Query())

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

// AddHandler adds handler with a list of middlewares
//
// # Handler
//
// Handler is any struct with http methods (i.e. `GET`, `POST`) as method.
// Methods with http.HandlerFunc signature are automaticaly added as a HTTP method handler.
// - Routing path is automatically discovered based on relative path to the router's `HandlerPath`.
// - Custom routing path (*absolute* or *relative*) can be set using a struct tag, e.g. `_ struct{} `limi:"path:/custom-path"` field in the Handler struct.
// - Multiple paths can be added to handle multiple paths, e.g. `_ struct{} `limi:"path=/story/cool-path,/story/strange-path,/best-path"`.
func (r *Router) AddHandler(handler Handler, mws ...func(http.Handler) http.Handler) error {
	rt := reflect.TypeOf(handler)
	baseRT := rt
	if rt.Kind() == reflect.Pointer {
		baseRT = rt.Elem()
	}
	rv := reflect.ValueOf(handler)

	if baseRT.Kind() != reflect.Struct {
		return fmt.Errorf("unsupported handler type %s %w", baseRT.Kind(), limi.ErrUnsupportedOperation)
	}

	methods := httpMethodHandlers{
		m:                       make(map[string]http.Handler),
		methodNotAllowedHandler: r.methodNotAllowedHandler,
		paramsType:              getParamsType(baseRT),
	}

	for i := 0; i < rt.NumMethod(); i++ {
		m := rt.Method(i)
		lName := strings.ToUpper(m.Name)
		if isHTTPHandlerProducer(m.Func) {
			vs := m.Func.Call([]reflect.Value{rv})
			v := vs[0]
			if v.Kind() == reflect.Func {
				methods.m[lName] = attachMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					v.Call([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(req)})
				}), mws...)
			}
		} else if isHTTPHandlerMethod(m.Func) {
			methods.m[lName] = attachMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				m.Func.Call([]reflect.Value{rv, reflect.ValueOf(w), reflect.ValueOf(req)})
			}), mws...)
		}
	}

	if len(methods.m) == 0 {
		return nil
	}

	for _, path := range resolvePaths(baseRT, r.handlerPath) {
		if err := r.insertMethodHandler(path, methods); err != nil {
			return fmt.Errorf("failed to insert methods handler with path %s %w", path, err)
		}
	}

	return nil
}

// AddHandlers adds multiple handlers with a list of middlewares.
func (r *Router) AddHandlers(handlers []Handler, mws ...func(http.Handler) http.Handler) error {
	for _, h := range handlers {
		if err := r.AddHandler(h, mws...); err != nil {
			return err
		}
	}
	return nil
}

// AddHandlerFunc adds http handler with path and method.
func (r *Router) AddHandlerFunc(path string, method string, fn http.HandlerFunc, mws ...func(http.Handler) http.Handler) error {
	path = r.buildPath(path)
	return r.insertMethodHandler(path, httpMethodHandlers{
		m: map[string]http.Handler{
			method: attachMiddlewares(fn, mws...),
		},
		methodNotAllowedHandler: r.methodNotAllowedHandler,
	})
}

// AddHTTPHandler adds a catch all http handler with path.
func (r *Router) AddHTTPHandler(path string, h http.Handler, mws ...func(http.Handler) http.Handler) error {
	path = r.buildPath(path)
	middlewares := append(r.middlewares, mws...)
	h = attachMiddlewares(h, middlewares...)
	return r.node.Insert(path, limi.HTTPHandler(h.ServeHTTP))
}

// AddRouter adds a sub router.
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
			handler = attachMiddlewares(handler, nr.middlewares...)
			handler.ServeHTTP(w, req)
		})
	}

	if err := r.insertRouter(nr); err != nil {
		return nil, fmt.Errorf("error inserting router %w", err)
	}

	return nr, nil
}

// IsSupportedHost match host with supported host.
func (r *Router) IsSupportedHost(ctx context.Context, host string) bool {
	if r.host == nil {
		return true
	}

	h, _ := r.host.Lookup(ctx, host)
	return h != nil
}

// IsPartial implements Node Handle interface, returning true indicates partial match is return for futher matching.
func (r *Router) IsPartial() bool {
	return true
}

// Merge implements Node Handle interface, returning false cause a Router doesn't merge multiple handles.
func (r *Router) Merge(limi.Handle) bool {
	return false
}

// IsMethodAllowed implements Node Handle interface, returns true to allow all methods on a router Handle
func (r *Router) IsMethodAllowed(string) bool {
	return true
}

// newRouter creates a new Router with list of RouterOptions.
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

// insertMethodHandler inserts new handler.
func (r *Router) insertMethodHandler(path string, h httpMethodHandlers) error {
	path = r.buildPath(path)
	handlers := buildMethodsHandlers(h, r.middlewares...)

	return r.node.Insert(path, handlers)
}

// insertRouter inserts new router.
func (r *Router) insertRouter(r1 *Router) error {
	return r.node.Insert(r.buildPath(r1.path), r1)
}

// lookup lookup for a Handle to path, with the remining unmatched string.
func (r *Router) lookup(ctx context.Context, path string) (limi.Handle, string) {
	router := r
	findPath := path
	for {
		h, trail := router.node.Lookup(ctx, findPath)

		// exact match
		if h != nil && trail == "" {
			return h, ""
		}

		// handle not found or (trail is not empty and no sub matchers)
		if h == nil || !h.IsPartial() {
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

// buildPath return a subpath relative to the router's path.
func (r *Router) buildPath(path string) string {
	if r.isSubRoute {
		return path
	}
	return buildPath(r.path, path)
}

// buildMethodsHandlers returns the methods handlers map with middlewares attached.
func buildMethodsHandlers(hms httpMethodHandlers, mws ...func(http.Handler) http.Handler) httpMethodHandlers {
	handlers := hms

	for method, h := range hms.m {
		handlers.m[method] = attachMiddlewares(h, mws...)
	}
	return handlers
}

// attachMiddlewares returns a Handler with list of middlewares attached to handler.
func attachMiddlewares(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	if len(mws) == 0 {
		return h
	}

	h = mws[len(mws)-1](h)
	for i := len(mws) - 2; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// buildPath returns a path with parent prefix.
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

// ensureLeadingSlash returns a path with '/' prefix.
func ensureLeadingSlash(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

// removeTraillingSlash returns a path without '/'.
func removeTraillingSlash(path string) string {
	return strings.TrimSuffix(path, "/")
}

// ensureTrailingSlash returns a path with '/' suffix.
func ensureTrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}

// removeLeadingSlash returns a path without '/' prefix.
func removeLeadingSlash(path string) string {
	return strings.TrimPrefix(path, "/")
}

// findHandlerPath returns a string found after the handlerPath
func findHandlerPath(handlerPath, path string) string {
	arrPath := strings.SplitAfter(path, handlerPath)
	if len(arrPath) == 1 {
		return path
	}

	return strings.Join(arrPath[1:], "/")
}

// packageName returns the package name from the package path.
func packageName(path string) string {
	idx := strings.LastIndexByte(removeTraillingSlash(path), '/')
	if idx < 0 {
		return path
	}

	return path[idx+1:]
}

// methodNotAllowedHandler returns a default handler when method a not allow for a path.
func methodNotAllowedHandler(allowedMethods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, m := range allowedMethods {
			w.Header().Add("Allow", strings.ToUpper(m))
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

// getPaths return multiple paths tag in str
func getPaths(t reflectTyper) []string {
	var limiTag string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if limiTag = field.Tag.Get("limi"); limiTag != "" {
			break
		}
	}

	if limiTag == "" {
		return nil
	}

	strs := strings.Split(limiTag, "=")
	if len(strs) != 2 ||
		strings.TrimSpace(strs[0]) != "path" {
		return nil
	}

	pathStrs := limi.SplitEscape(strs[1], ',')
	var paths []string
	for _, p := range pathStrs {
		paths = append(paths, strings.TrimSpace(p))
	}

	return paths
}

type reflectTyper interface {
	Name() string
	PkgPath() string
	Field(int) reflect.StructField
	NumField() int
}

// resolvePaths build paths from a struct type with PkgPath, struct name and struct tag
func resolvePaths(t reflectTyper, handlerPath string) []string {
	pkgPath := t.PkgPath()
	structName := t.Name()

	paths := getPaths(t)
	structName = strings.ToLower(structName)
	pkgName := packageName(pkgPath)
	trimmedPkgPath := removeTraillingSlash(findHandlerPath(handlerPath, pkgPath))
	// no path tag found, default to pkgPath + structname
	if len(paths) == 0 {
		path := trimmedPkgPath
		if pkgName != structName && structName != defaultIndexStructName {
			path += ensureLeadingSlash(structName)
		}
		paths = append(paths, path)
	}

	// fix relative path
	for i, path := range paths {
		if !strings.HasPrefix(path, "/") {
			path = strings.TrimPrefix(path, ".")
			path = trimmedPkgPath + ensureLeadingSlash(path)
		}
		paths[i] = path
	}

	return paths
}

// Map of HTTP Handlers by Methods.
type httpMethodHandlers struct {
	m                       map[string]http.Handler
	methodNotAllowedHandler func(...string) http.Handler
	paramsType              reflect.Type
}

// keys returns a list of methods supported by the handler.
func (h httpMethodHandlers) keys() []string {
	var keys []string
	for k := range h.m {
		keys = append(keys, k)
	}
	return keys
}

// IsPartial implements Node Handle interface, returning false indicates partial match is not handled.
func (h httpMethodHandlers) IsPartial() bool {
	return false
}

// IsMethodAllowed implements Node Handle interface, returns true when the method is supported.
func (h httpMethodHandlers) IsMethodAllowed(method string) bool {
	_, ok := h.m[method]
	return ok
}

// ServeHTTP implements Node Handle interface, handles net/http server requests.
func (h httpMethodHandlers) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hdl, ok := h.m[req.Method]
	if !ok {
		h.methodNotAllowedHandler(h.keys()...).ServeHTTP(w, req)
		return
	}

	if h.paramsType != nil {
		limi.SetParamsType(req.Context(), h.paramsType)
	}
	hdl.ServeHTTP(w, req)
}

// Merge implements Node Handle interface, merges the handler from h1.
// Returns true when existing handler is not found.
// Returns false when existing handler is found.
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

// parseHost return the host part of the req.Host
func parseHost(str string) string {
	arr := strings.Split(str, ":")

	if len(arr) == 0 {
		return ""
	}

	return arr[0]
}

// hostHandler is a node Handle to match router's host
type hostHandler struct{}

// IsPartial implements Node Handle interface, returning false indicates partial match is not handled.
func (h hostHandler) IsPartial() bool {
	return false
}

// Merge implements Node Handle interface, returning false indicates merging is not allowed.
func (h hostHandler) Merge(limi.Handle) bool {
	return false
}

// IsMethodAllowed implements Node Handle interface, returns true to handle all http methods.
func (h hostHandler) IsMethodAllowed(string) bool {
	return true
}

// ServeHTTP implements Node Handle interface, handles net/http server requests.
func (h hostHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {}

// getParamsType check if a struct type params is set in the field
func getParamsType(t reflect.Type) reflect.Type {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		limiTag := field.Tag.Get("limi")
		if limiTag == "" {
			continue
		}

		if field.Type.Kind() != reflect.Struct {
			continue
		}

		ft := field.Type
		for j := 0; j < ft.NumField(); j++ {
			ftField := ft.Field(j)
			limiTag := ftField.Tag.Get("limi")
			if limiTag == "" {
				continue
			}

			if strings.Contains(limiTag, "param") ||
				strings.Contains(limiTag, "query") {
				return field.Type
			}
		}
	}

	return nil
}
