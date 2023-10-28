package main

import (
	"blog/pkg/handler/admin"
	"blog/pkg/handler/blog"
	"net/http"

	"github.com/sanekee/limi"
)

func main() {
	r, err := limi.NewRouter("/") // a new router processing request on /
	if err != nil {
		if err != nil {
			panic(err)
		}
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

	if err := r.AddHandlerFunc("/about", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("About Us")) // nolint:errcheck
	}); err != nil {
		panic(err)
	}

	// adds a catch all handler at /admin
	if err := r.AddHTTPHandler("/admin", admin.Admin{}); err != nil {
		panic(err)
	}

	if err := http.ListenAndServe(":3333", r); err != nil {
		panic(err)
	}
}
