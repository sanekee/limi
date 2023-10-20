package foo

import (
	"fmt"
	"net/http"
)

type FooPkg struct{}

func (f FooPkg) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo"))
	}
}

type Foo struct {
	limi struct{} `path:"/foo"`
}

func (f Foo) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo"))
	}
}

type FooPtr struct {
	state int
}

func (f *FooPtr) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		f.state = f.state + 1
		w.Header().Add("X-Count-Total", fmt.Sprintf("%d", f.state))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo"))
	}
}

type FooRel struct {
	limi struct{} `path:"bar"`
}

func (f FooRel) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo"))
	}
}

type FooHdl struct{}

func (f FooHdl) Get(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("foo"))
}

type FooPtrHdl struct {
	state int
}

func (f *FooPtrHdl) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		f.state = f.state + 1
		w.Header().Add("X-Count-Total", fmt.Sprintf("%d", f.state))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo"))
	}
}
