package main

import (
	"log"
	"net/http"

	"github.com/sanekee/limi"
	"github.com/sanekee/limi/middleware"
)

func main() {
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
}
