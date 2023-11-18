# Limi

Limi is a lightweight go http router. The goal of the project is to make writing REST application easier with composable middlewares and idiomatic handler interface.

## Table of Contents

- [Features](#features)
- [Components](#components)
- [Usage](#usage)
  - [Setup](#setup)
  - [Router](#router)
  - [Handlers](#handlers)
  - [Middlewares](#middlewares)
  - [Mux](#mux)
- [Pattern Matching](#pattern-matching)
- [URL Parameters Binding](#url-parameters-binding)

## Features

- Lightweight with only go standard library dependencies.
- Idiomatic handler, automatic handler's path discovery, reflection based url params binding.
- Similar syntax for host and path matching.
- Cascading middlewares support at router, subrouter and handler level.

## Components

| Name       | Description                                                     |
| ----       | --------------------------------------------------------------- |
| Router     | Router is the core of limi, a router handles http request with customizable `host` and `path`. Both `host` and `path` a match with a Radix Tree with custom matchers to provide *O(k)* time complexity lookup. limi Router is fully compatible with `net/http`.            |
| Handler    | Handler is the function that handle the http request, limi Handler is fully compaitble with `net/http` `http.Handler`.                                                |
| Middleware | Middlewares are chainable functions injected in router or handler lever to customize the handler functionality. Limi middlewares are compaible with generic middlewares used in other http routers such as `go-chi`, `gorilla-mux`.                                               |
| Mux        | Mux is a router multiplexer, mux is used to serve single HTTP listener with multiple routers. Requests are handled by the order the routers were added.             |

## Usage

### Setup

Import limi in your project.

```shell
go get github.com/sanekee/limi
```

### Router

Adding a router.

#### Router Options

| Option Function            | Description                                                |
| -------------------------- | ---------------------------------------------------------- |
| WithHosts                  | Create router with `host` matching. Supports multiple hosts with common pattern matching. |
| WithMiddlewares            | Attach middlewares to router.                              |
| WithNotFoundHandler        | Set `not found`` handler.                                  |
| WithMethodNotAllowedHandler| Set the `method not allowed` handler.                      |
| WithProfiler               | Attach golang profiler to router at `/debug/pprof/`.       |
| WithHandlerPath            | Set the base path for Handler, default is `handler`.       |

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

| Method         | Type             | Description                              |
| -------------- | ---------------- | ---------------------------------------- |
| AddHandler     | Handler | Handler is any struct with http methods (i.e. `GET`, `POST`) as method.<br>- Methods with http.HandlerFunc signature are automaticaly added as method handler.<br>- Routing path is automatically discovered based on relative path to the router's `HandlerPath`.<br>- Custom routing path (*absolute* or *relative*) can be set using a struct tag, e.g. *_ struct{} \`limi:"path=/custom-path"\`*<br>- Multiple paths can be added to handle multiple paths, e.g. *_ struct{} \`limi:"path=/story/cool-path,/story/strange-path,/best-path"\`*<br>- URL Params binding with custom params struct, e.g. *_ commentParams{} \`limi:"path=/author/{id}/story/{slug}/comments/{commendId}"\`* |
| AddHandlerFunc | http.HandlerFunc | `http.HandlerFunc` is `net/http` handler function. |
| AddHTTPHandler | http.Handler | `http.Handler` is `net/http` handler with `ServeHTTP` method, using this as a catch all handler.                                                   |

#### Path Discovery

Below are path discovery under different scenarios using the router's default HandlerPath (`handler`)

- Struct with the same name or named "Index" is added as the index handler.

```golang
// package /pkg/handler
package handler

type Index struct{}          // path => /

// package /pkg/handler
package handler

type Handler struct{}        // path => /

// package /pkg/handler/bar
package bar

type Index struct{}          // path => /bar

// package /pkg/handler/foo
package foo

type Foo struct{}            // path => /foo
```

- Other struct is added as a handler with their lowercase name.

```golang
// package /pkg/handler
package handler

type Foo struct{}          // path => /foo

// package /pkg/handler/foo
package handler

type Bar struct{}          // path => /foo/bar
```

- Path definition with the `limi` struct tag.

```golang
// package /pkg/handler
package handler

type Foo struct{
    _ struct{} `limi:"path=./"`                // path => /                 // ./ is added as index handler
}          

// package /pkg/handler/foo
package handler

type Bar struct{
    _ struct{} `limi:"path=foo"`                // path => /foo/foo         // relative path to the package
}          

// package /pkg/handler/foo
package handler

type HandleByID struct{
    _ struct{} `limi:"path={id:[0-9]+}"`        // path => /foo/{id:[0-9]+} // relative regexp path to the package
}          


// package /pkg/handler/foo
package handler

type Bar struct{
    _ struct{} `limi:"path=/"`                 // path => /                 // absolute path
}          

// package /pkg/handler/foo
package handler

type Bar struct{
    _ struct{} `limi:"path=/tar/bar/zar"`     // path => /tar/bar/zar       // absolute path #2
}          

// package /pkg/handler/foo
package handler

type Bar struct{
    _ struct{} `limi:"path=/user/{id}/story/{slug:.*}"`  // path => /foo/{id}/bar/{slug:.*}
                                                        // absolute path with a label matcher {id} and a regexp matcher {slug:.*}       
}          
```

- Multiple paths definition with the `limi` struct tag. (Non standard struct tag, will break some linter.)

```golang
// package /pkg/handler
package handler

type Foo struct{
    _ struct{} `limi:"path=./,/good/foo,/bad/foo"`  // path => /, /good/foo, /bad/foo
}          
```

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
    _ authorParams `limi:"path={storyId:[0-9]+}/author"` // custom relative path
}

type authorParams struct {
    storyID int `limi:"param=storyId"`
}


// Get handles HTTP GET request
func (s Author) Get(w http.ResponseWriter, req *http.Request) {
    params, err := limi.GetParams[authorParams](req.Context())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    // retrieve author and response
    author := getAuthor(params.storyID)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("The auther is " + author))
}

type Copyright struct {
    _ struct{} `limi:"path=/copyright"` // custom absolute path
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

Mux is the router multiplexer. Using mux when we need multiple routers in a single listener.

#### Example

Full example can be found in [example/mux](example/mux).

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

m := limi.NewMux(r1, r2)

if err := http.ListenAndServe(":3333", m); err != nil {
    panic(err)
}
```

## Pattern Matching

Pattern matcher is an internal component in limi router. It's used in conjuction of the Radix Tree to lookup a `host` or `path` to find the right handler.

| Matcher Type | Syntax | Priority | Descrption |
| --- | --- | --- | --- |
| String | mypath | 1 | A string matcher matches the exact string (case sensitive). |
| Regexp | {myid:[0-9]+} | 2 | A regular expression matcher uses the regular expression syntax defined after the colon (e.g. `[0-9]+`) to match string. Matched value will be set in the value context. |
| Label | {slug} | 3 | A label wildcard matcher matches everything. Matched value is set in the value context. |

When a string matches multiple matchers, they are matched according to the priority.

#### Example

```golang
r, _ := NewRouter("/", 
    WithHosts(
        "static.domain.com",                 // matches the host static.domain.com 
        "{apiVer:v[0-9]+}.api.domain.com",   // matches hosts v1.api.domain.com, v2.api.domain.com ... and sets URLParams["apiVer"] = value
        "{subdomain}.domain.com",            // matches hosts subdomain1.domain.com, subdomain2.domain.com ... and sets URLParams["subdomain"] = value
))

r.AddHandlerFunc("/blog/top" ..              // matches the path /blog/top

r.AddHandlerFunc("/blog/{id:[0-9]+}" ..      // matches paths /blog/1, /blog/2 ..., sets URLParams["id"] = <value>

r.AddHandlerFunc("/blog/{slug}" ..           // matches paths /blog/cool-article-1, /blog/cool-article-2 ..., sets URLParam["slug"] = <value>
```

## URL Parameters Binding

Limi supports binding custom struct with common data types or custom `stringer` types.

```golang
type stringer interface {
    FromString(string) error
}
```

To bind a struct to a handler with URL parameters,

1. Declare a limi tagged field in the handler with a parameters struct.

### Example

```golang
type CommentsHandler struct {
    _ commentParams `limi:"path=/author/{id}/blog/{slug}/comment/{commentId}`
}

type commentParams struct {
    id int `limi:"param"`                       // url param is the same as field name = {id}
    blog string `limi:"param=slug"`             // url param is {slug}
    commentID myuuid `limi:"param=commentId"`   // url param is a custom type myuuid {commentId}
}

type myuuid string

func (m *myuuid) FromString(val string) error {
    *m = val
    return nil
}

func (c CommentHandler) Get(w http.ResponseWriter, req *http.Request) {
    params, err := limi.GetParams[commentParams](req.Context())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    
    fmt.Println(params)
}
```

2. Using SetURLParamsData middleware.

#### Example

```golang
r, _ := limi.NewRouter("/") // a new router processing request on /

urlparams,err := SetURLParamsData(commentParams{})
if err != nil {
    panic(err)
}

if err := r.AddHandlerFunc("/author/{id}/blog/{slug}/comment/{commentId}", 
    http.MethodGet,
    GetComment,
    urlparams,
); err != nil {
    panic(err)
}

type commentParams struct {
    id int `limi:"param"`                       // url param is the same as field name = {id}
    blog string `limi:"param=slug"`             // url param is {slug}
    commentID myuuid `limi:"param=commentId"`   // url param is a custom type myuuid {commentId}
}

type myuuid string

func (m *myuuid) FromString(val string) error {
    *m = val
    return nil
}

func GetComment(w http.ResponseWriter, req *http.Request) {
    params, err := limi.GetParams[commentParams]()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    
    fmt.Println(params)
}
```
