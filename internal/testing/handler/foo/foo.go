package foo

import (
	"fmt"
	"net/http"

	"github.com/sanekee/limi/internal/testing/handler"
)

type FooPkg struct{}

func (f FooPkg) Get() http.HandlerFunc {
	return handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo"))
}

type Foo struct{}

func (f Foo) Get() http.HandlerFunc {
	return handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo"))
}

type FooDef struct {
	_ struct{} `limi:"path=/foo"`
}

func (f FooDef) Get() http.HandlerFunc {
	return handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo"))
}

type FooPtr struct {
	state int
}

func (f *FooPtr) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		f.state = f.state + 1
		w.Header().Add("X-Count-Total", fmt.Sprintf("%d", f.state))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo")) //nolint:errcheck
	}
}

type FooRel struct {
	_ struct{} `limi:"path=bar"`
}

func (f FooRel) Get() http.HandlerFunc {
	return handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo"))
}

type FooHdl struct{}

func (f FooHdl) Get(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("foo")) //nolint:errcheck
}

type FooPtrHdl struct {
	state int
}

func (f *FooPtrHdl) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		f.state = f.state + 1
		w.Header().Add("X-Count-Total", fmt.Sprintf("%d", f.state))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo")) //nolint:errcheck
	}
}

type Foo1 struct{}

func (f Foo1) Get() http.HandlerFunc {
	return handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo1"))
}

type Foo2 struct{}

func (f Foo2) Get() http.HandlerFunc {
	return handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo2"))
}

type FooMulti struct{}

func (f FooMulti) Get() http.HandlerFunc {
	return handler.NewHandlerFunc(http.StatusOK, nil, []byte("get foomulti"))
}

func (f FooMulti) Post() http.HandlerFunc {
	return handler.NewHandlerFunc(http.StatusOK, nil, []byte("post foomulti"))
}

type FooSlash struct {
	_ struct{} `limi:"path=./"`
}

func (f FooSlash) Get() http.HandlerFunc {
	return handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo"))
}

type FooPaths struct {
	_ struct{} `limi:"path=foo1,foo2"`
}

func (f FooPaths) Get() http.HandlerFunc {
	return handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo"))
}
