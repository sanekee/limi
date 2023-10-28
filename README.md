# Limi

Limit is a lightweight go http router. The goal of the project is to make writing REST application easier with composable middlewares and idiomatic handler interface.

## Features

- Lightweight with only go standard library dependencies.
- Idiomatic handler, handler's path is automaitically discovered based on its package path.
- Similar syntax for host and path matching.
- Cascading middlewares support at router, subrouter and handler level.

## Components

| Name | Description |
| --- | --- |
| Router | Router is the core of Limit, a router handles http request with customizable `host` and `path`. Both `host` and `path` a match with a Radix Tree with custom matchers to provide *O(k)* time complexity lookup. Limi Router is fully compatible with `net/http`. |
| Handler | Handler is the function that handle the http request, Limi Handler is fully compaitble with `net/http` `http.Handler`. |
| Middleware | Middlewares are chainable functions injected in router or handler lever to customize the handler functionality. Limi middlewares are compaible with middlewares used in other http routers such as `go-chi`, `gorilla-mux`. |
| Mux | Mux is a router multiplexer, mux is used to serve single HTTP listener with multiple routers. Requests are handled by the order the routers were added. |

## Usage

### Router

#### Router Options

| Option Function | Description |
| --- | --- |
| WithHosts | Create router with `host` matching. Supports multiple hosts with common pattern matching. |
| WithMiddlewares | Attach middlewares to router. |
| WithNotFoundHandler | Set `not found`` handler. |
| WithMethodNotAllowedHandler| Set the `method not allowed` handler. |
| WithProfiler | Attach golang profiler to router at `/debug/pprof/`. |
| WithHandlerPath | Set the base path for Handler, default is `handler`. |

#### Examples

Creating a basic router.

```golang
r, err := limi.NewRouter("/") // a new router processing request on /
if err != nil {
    panic(err)
}
```

Creating a router with options.

```golang
r, err := limi.NewRouter(
    // handling on path /v1/api
    "/v1/api",
    // matching host = localhost, or subdomains on .example.com, or api.limiiscool.com
    WithHosts("localhost","{subdomain:.*}.example.com","api.limiiscool.com"),
    // with a logging middleware
    WithMiddlewares(middleware.Log(log.Default())),
    // with golang profiler /v1/api/debug/pprof
    WithProfiler(),
    // with handlers in <src>/pkg/api
    WithHandlerPath("api")
)
if err != nil {
    panic(err)
}
```

Adding a subrouter.

```golang
r, err := limi.NewRouter(
    "/v1",
)
if err != nil {
    panic(err)
}

sr, err := r.AddRouter(
    "/api",
    WithHandlerPath("api")
)
if err != nil {
    panic(err)
}
```

### Handlers

#### Adding Handlers

| Method | Type | Description |
| --- | --- | --- |
| AddHandler | Handler | Handler is any struct with http methods (i.e. `GET`, `POST`) as method.<br>- Methods with http.HandlerFunc signature are automaticaly added as method handler.<br>- Routing path is automatically discovered based on relative path to the router's `HandlerPath`.<br>- Custom routing path (*absolute* or *relative*) can be set using a path tag to a `limi struct{} `path:"/custom-path"` field in the Handler struct. |
| AddHandlerFunc | http.HandlerFunc | `http.HandlerFunc` is `net/http` handler function. |
| AddHTTPHandler | http.Handler | `http.Handler` is `net/http` handler with `ServeHTTP` method, using this as a catch all handler. |

#### Example

Full example can be found in [example/blog](example/blog).

`AddHandler`

```golang
// in pkg/handler/blog
package blog

type Blog struct {}

// Get handles HTTP GET request
func (s Blog) Get(w http.ResponseWriter, req *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("A cool story."))
}

// Post handles HTTP POST request
func (s Blog) Post(w http.ResponseWriter, req *http.Request) {
    // create story   
    w.WriteHeader(http.StatusOK)
}

type Author struct {
    limi struct{} `path:"{storyId:[0-9]+}/author"` // custom relative path
}

// Get handles HTTP GET request
func (s Author) Get(w http.ResponseWriter, req *http.Request) {
    idStr := limi.GetURLParam(req.Context(), "storyId")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    // retrieve author and response
    author := getAuthor(id)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("The auther is " + author))
}

type Copyright struct {
    limi struct{} `path:"/copyright"` // custom absolute path
}

// Get handles HTTP GET request
func (s Copyright) Get(w http.ResponseWriter, req *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Copyright ©️ 2023"))
}

// in main.go
package main

import "mypkg/pkg/handler/blog"


r, err := limi.NewRouter("/") // a new router processing request on /
if err != nil {
    panic(err)
}

// add a handler at /blog
if err := r.AddHandler(blog.Blog{}); err != nil {
    panic(err)
}

// add a handler at /blog/{storyId:[0-9]+}/author
if err := r.AddHandler(blog.Author{}); err != nil {
    panic(err)
}

// add a handler at /copyright
if err := r.AddHandler(blog.Copyright{}); err != nil {
    panic(err)
}

if err := http.ListenAndServe(":3333", r); err != nil {
    panic(err)
}
```

`AddHandlerFunc`

```golang
r, err := limi.NewRouter("/") // a new router processing request on /
if err != nil {
    panic(err)
}

if err := r.AddHandlerFunc("/about", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("About Us"))
}); err != nil {
    panic(err)
}
```

`AddHTTPHandler`

```golang
// in pkg/handler/admin

type Admin struct{}

func (a Admin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Admin portal")) // nolint:errcheck
}

// in main
r, err := limi.NewRouter("/") // a new router processing request on /
if err != nil {
    panic(err)
}

// adds a catch all handler at /admin
if err := r.AddHTTPHandler("/admin", admin.Admin{}); err != nil {
    panic(err)
}
```

### Middlewares

Middlewares are chainable http.Handler.

#### Example

```golang
func MyMiddleWare(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req* http.Request) {
        // do cool stuf
        next.ServeHTTP(w, req)
    })
}
```

An example logging middleware can be found in the [middleware/log.go](middleware/log.go).

### Mux

Mux is the router multiplexer. Use when we need multiple routers in a single listener.

#### Example

Full example can be found in [example/blog](example/blog).

```golang
r1, err := limi.NewRouter(
    "/",
    limi.WithHosts("v1.example.com"),
)
if err != nil {
    panic(err)
}

r1.AddHTTPHandler("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    // handling v1 api
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("V1")) // nolint:errcheck
}))

r2, err := limi.NewRouter(
    "/",
    limi.WithMiddlewares(middleware.Log(log.Default())), // enable logging in v2 api
    limi.WithHosts("v2.example.com"),
)
if err != nil {
    panic(err)
}

r2.AddHTTPHandler("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    // handling v1 api
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("V2")) // nolint:errcheck
}))

m := limi.Mux()
m.AddRouters(r1, r2)

if err := http.ListenAndServe(":3333", m); err != nil {
    panic(err)
}
```

### Pattern Matching

Pattern matcher is an internal component in Limi router. It's used in conjuction of the Radix Tree to lookup a `host` or `path` to find the right handler.

| Matcher Type | Syntax | Priority | Descrption |
| --- | --- | --- |
| String | mypath | 1 | A string matcher matches the exact string (case sensitive). |
| Regexp | {myid:[0-9]+} | 2 | A regular expression matcher uses the regular expression syntax defined after the colon (e.g. `[0-9]+`) to match string. Matched value will be set in the value context. |
| Label | {slug} | 3 | A label matcher matches everything and value is in the value context. |

When multiple handler with similar matcher is found, they are matched accoriding to the priority.

#### Example

```golang
r.AddHandlerFunc("/mypath" ..// matches the exact path /mypath

r.AddHandlerFunc("/mypath/{id:[0-9]+}" ..// matches the path /mypath/1, /mypath/2 ... sets URLParam[id] = <value>

r.AddHandlerFunc("/mypath/{slug}" ..// matches the path /mypath/cool-article-1, /mypath/cool-article-2 ... sets URLParam[slug] = <value>
```
