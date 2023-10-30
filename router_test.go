package limi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/sanekee/limi/internal/testing/handler"
	"github.com/sanekee/limi/internal/testing/handler/foo"
	"github.com/sanekee/limi/internal/testing/require"
)

func TestBuildPath(t *testing.T) {
	type test struct {
		parent   string
		path     string
		expected string
	}

	tests := []test{
		{
			parent:   "/parent/",
			path:     "/foo",
			expected: "/parent/foo",
		},
		{
			parent:   "/parent",
			path:     "/foo/",
			expected: "/parent/foo/",
		},
		{
			parent:   "/",
			path:     "/foo/",
			expected: "/foo/",
		},
		{
			parent:   "",
			path:     "",
			expected: "",
		},
		{
			parent:   "",
			path:     "foo",
			expected: "/foo",
		},
		{
			parent:   "/",
			path:     "/",
			expected: "/",
		},
		{
			parent:   "/",
			path:     "/foo",
			expected: "/foo",
		},
	}

	t.Parallel()

	for _, r := range tests {
		t.Run("", func(t *testing.T) {
			actual := buildPath(r.parent, r.path)
			require.Equal(t, r.expected, actual)
		})
	}
}

func TestPathHelpers(t *testing.T) {
	type test struct {
		input    string
		expected string
	}

	t.Run("ensureLeadingSlash", func(t *testing.T) {
		tests := []test{
			{
				input:    "foo",
				expected: "/foo",
			},
			{
				input:    "/foo",
				expected: "/foo",
			},
		}

		for _, test := range tests {
			actual := ensureLeadingSlash(test.input)
			require.Equal(t, test.expected, actual)
		}
	})

	t.Run("removeTraillingSlash", func(t *testing.T) {
		tests := []test{
			{
				input:    "foo/",
				expected: "foo",
			},
			{
				input:    "foo",
				expected: "foo",
			},
		}

		for _, test := range tests {
			actual := removeTraillingSlash(test.input)
			require.Equal(t, test.expected, actual)
		}
	})

	t.Run("ensureTrailingSlash", func(t *testing.T) {
		tests := []test{
			{
				input:    "foo",
				expected: "foo/",
			},
			{
				input:    "foo/",
				expected: "foo/",
			},
		}

		for _, test := range tests {
			actual := ensureTrailingSlash(test.input)
			require.Equal(t, test.expected, actual)
		}
	})

	t.Run("removeLeadingSlash", func(t *testing.T) {
		tests := []test{
			{
				input:    "/foo",
				expected: "foo",
			},
			{
				input:    "foo",
				expected: "foo",
			},
		}

		for _, test := range tests {
			actual := removeLeadingSlash(test.input)
			require.Equal(t, test.expected, actual)
		}
	})
}

func TestFindHandlerPath(t *testing.T) {
	type test struct {
		pkgPath  string
		expected string
	}

	tests := []test{
		{
			pkgPath:  "base/handler/foo/",
			expected: "/foo/",
		},
		{
			pkgPath:  "base/handler/handlerfoo/",
			expected: "/handler/foo/",
		},
		{
			pkgPath:  "/foo",
			expected: "/foo",
		},
	}

	t.Parallel()

	for _, r := range tests {
		t.Run("", func(t *testing.T) {
			actual := findHandlerPath("handler", r.pkgPath)
			require.Equal(t, r.expected, actual)
		})
	}
}

func TestIsHTTPHandlerProducer(t *testing.T) {
	t.Run("handler is http handler func producer", func(t *testing.T) {
		fn := func() http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
		}
		v := reflect.ValueOf(fn)

		actual := isHTTPHandlerProducer(v)
		require.True(t, actual)
	})

	t.Run("func is http handler func producer", func(t *testing.T) {
		fn := func() http.HandlerFunc {
			return func(w http.ResponseWriter, req *http.Request) {}
		}
		v := reflect.ValueOf(fn)

		actual := isHTTPHandlerProducer(v)
		require.True(t, actual)
	})

	t.Run("returning func is not http handler func producer", func(t *testing.T) {
		fn := func() func() {
			return func() {}
		}
		v := reflect.ValueOf(fn)

		actual := isHTTPHandlerProducer(v)
		require.False(t, actual)
	})

	t.Run("func is not http handler func producer", func(t *testing.T) {
		fn := func() {}

		v := reflect.ValueOf(fn)

		actual := isHTTPHandlerProducer(v)
		require.False(t, actual)
	})

	t.Run("struct is not http handler func producer", func(t *testing.T) {
		st := struct{}{}
		v := reflect.ValueOf(st)

		actual := isHTTPHandlerProducer(v)
		require.False(t, actual)
	})
}

func TestIsHTTPHandlerMethod(t *testing.T) {
	t.Run("method is http handler method ", func(t *testing.T) {
		st := testSt{}

		v := reflect.ValueOf(st.Get)

		actual := isHTTPHandlerMethod(v)
		require.True(t, actual)
	})

	t.Run("method is http handler method ", func(t *testing.T) {
		st := testSt{}

		vSt := reflect.TypeOf(st)
		mGet, ok := vSt.MethodByName("Get")
		require.True(t, ok)

		actual := isHTTPHandlerMethod(mGet.Func)
		require.True(t, actual)
	})

	t.Run("method is not http handler method ", func(t *testing.T) {
		st := testSt{}

		v := reflect.ValueOf(st.Fetch)

		actual := isHTTPHandlerMethod(v)
		require.False(t, actual)
	})

	t.Run("struct is not http handler func producer", func(t *testing.T) {
		st := struct{}{}
		v := reflect.ValueOf(st)

		actual := isHTTPHandlerMethod(v)
		require.False(t, actual)
	})
}

type testSt struct{}

func (t testSt) Get(http.ResponseWriter, *http.Request) {}

func (t testSt) Fetch() {}

func TestParseHost(t *testing.T) {
	t.Run("has port", func(t *testing.T) {
		actual := parseHost("host:8080")
		require.Equal(t, "host", actual)
	})

	t.Run("no colon", func(t *testing.T) {
		actual := parseHost("host")
		require.Equal(t, "host", actual)
	})

	t.Run("empty string", func(t *testing.T) {
		actual := parseHost("")
		require.Equal(t, "", actual)
	})
}

func TestAddHandler(t *testing.T) {
	t.Run("add with default package handler", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFoo := foo.Foo{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add with tag", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFoo := foo.FooDef{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add relative path with tag", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFooRel := foo.FooRel{}
		err = r.AddHandler(testFooRel)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/bar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add with package path", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFooPkg := foo.FooPkg{}
		err = r.AddHandler(testFooPkg)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/foopkg", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add stateful handler", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFooPtr := &foo.FooPtr{}
		err = r.AddHandler(testFooPtr)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/fooptr", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		count := rec.Result().Header.Get("X-Count-Total")
		require.Equal(t, "1", count)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/fooptr", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		count = rec.Result().Header.Get("X-Count-Total")
		require.Equal(t, "2", count)
	})

	t.Run("add with handler method", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFoo := foo.FooHdl{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/foohdl", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add stateful method handler", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFoo := &foo.FooPtrHdl{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/foo/fooptrhdl", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		count := rec.Result().Header.Get("X-Count-Total")
		require.Equal(t, "1", count)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/foo/fooptrhdl", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		count = rec.Result().Header.Get("X-Count-Total")
		require.Equal(t, "2", count)
	})

	t.Run("method not allowed", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFoo := foo.FooHdl{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "http://localhost:9090/foo/foohdl", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMethodNotAllowed, rec.Result().StatusCode)
		require.Equal(t, rec.Header().Get("Allow"), "GET")
	})

	t.Run("custom method not allowed", func(t *testing.T) {
		r, err := NewRouter("/",
			WithMethodNotAllowedHandler(func(m ...string) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusMethodNotAllowed)
					require.Equal(t, m[0], "GET")
				})
			}))
		require.NoError(t, err)

		testFoo := foo.FooHdl{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "http://localhost:9090/foo/foohdl", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMethodNotAllowed, rec.Result().StatusCode)
	})

	t.Run("add handlers", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFoo := foo.Foo{}
		testFooRel := foo.FooRel{}
		err = r.AddHandlers([]Handler{testFoo, testFooRel})
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/bar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add end slash", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFoo := foo.FooSlash{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add multiple path", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFoo := foo.FooPaths{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo1", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo2", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add with interface with handler", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		testFoo := foo.Foo{}
		var fooInterface interface{} = testFoo
		err = r.AddHandler(fooInterface)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})
}

func TestAddRouter(t *testing.T) {
	t.Run("add sub route", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		r1, err := r.AddRouter("/baz")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/baz/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add sub route & handler", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		_, err = r.AddRouter("/foo")
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo")))
		require.Error(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Result().StatusCode)
	})

	t.Run("add handler & sub route", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo")))
		require.NoError(t, err)

		_, err = r.AddRouter("/foo")
		require.Error(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add sub route & handler under subroute", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		r1, err := r.AddRouter("/foo")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/baz", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo/baz")))
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo/bar", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo/bar")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/baz", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo/baz", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/bar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo/bar", string(body))
	})

	t.Run("add sub route name part of longer handler", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foooo/bar", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foooo/bar")))
		require.NoError(t, err)

		r1, err := r.AddRouter("/foo")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/baz", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo/baz")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/baz", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo/baz", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/foooo/bar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foooo/bar", string(body))
	})

	t.Run("lookup inexistence route", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Result().StatusCode)
	})

	t.Run("lookup inexistence sub route", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		_, err = r.AddRouter("/foo")
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo")))
		require.Error(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/1", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Result().StatusCode)
	})

	t.Run("host", func(t *testing.T) {
		r, err := NewRouter("/", WithHosts("abc"))
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo:abc")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Result().StatusCode)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://abc:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo:abc", string(body))

	})

	t.Run("multi host", func(t *testing.T) {
		r, err := NewRouter("/",
			WithHosts("host1", "host2"),
		)
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			hostname := parseHost(req.URL.Host)
			w.Write([]byte("foo" + ":" + hostname)) // nolint:errcheck
		})
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://host1:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo:host1", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://host2:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo:host2", string(body))

	})

	t.Run("regex host", func(t *testing.T) {
		r, err := NewRouter("/",
			WithHosts("{host:[^.]+.hostname.com}"),
		)
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			hostname := parseHost(req.URL.Host)
			w.Write([]byte("foo" + ":" + hostname)) // nolint:errcheck
		})
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://subdomain1.hostname.com:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo:subdomain1.hostname.com", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://subdomain2.hostname.com:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo:subdomain2.hostname.com", string(body))
	})

	t.Run("add route path", func(t *testing.T) {
		r, err := NewRouter("/base")
		require.NoError(t, err)

		r1, err := r.AddRouter("/baz")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/baz/foo", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Result().StatusCode)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/base/baz/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add profiler", func(t *testing.T) {
		r, err := NewRouter("/admin", WithProfiler())
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/admin/debug/pprof/", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/admin/debug/pprof/heap?seconds=1", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		contentType := rec.Header().Get("Content-Type")
		require.Equal(t, "application/octet-stream", contentType)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/admin/debug/pprof/profile?seconds=1", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		contentType = rec.Header().Get("Content-Type")
		require.Equal(t, "application/octet-stream", contentType)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/admin/debug/pprof/trace?seconds=1", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		contentType = rec.Header().Get("Content-Type")
		require.Equal(t, "application/octet-stream", contentType)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/admin/debug/pprof/symbol?12345678", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		contentType = rec.Header().Get("Content-Type")
		require.Equal(t, "text/plain; charset=utf-8", contentType)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
	})

	t.Run("add profiler in subrouter", func(t *testing.T) {
		r, err := NewRouter("/admin")
		require.NoError(t, err)

		_, err = r.AddRouter("/profiler", WithProfiler())
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/admin/profiler/debug/pprof/", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/admin/profiler/debug/pprof/heap?seconds=1", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		contentType := rec.Header().Get("Content-Type")
		require.Equal(t, "application/octet-stream", contentType)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/admin/profiler/debug/pprof/profile?seconds=1", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		contentType = rec.Header().Get("Content-Type")
		require.Equal(t, "application/octet-stream", contentType)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/admin/profiler/debug/pprof/trace?seconds=1", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		contentType = rec.Header().Get("Content-Type")
		require.Equal(t, "application/octet-stream", contentType)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/admin/profiler/debug/pprof/symbol?12345678", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		contentType = rec.Header().Get("Content-Type")
		require.Equal(t, "text/plain; charset=utf-8", contentType)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
	})
}

func TestAddHandlerFunc(t *testing.T) {
	t.Run("add single func", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("merge same path", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(req.Method + ":foo")) // nolint:errcheck
		})
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodPost, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(req.Method + ":foo")) // nolint:errcheck
		})
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "GET:foo", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "POST:foo", string(body))
	})

	t.Run("merge same path with conflict", func(t *testing.T) {
		r, err := NewRouter("/")
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo")))
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo")))
		require.Error(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})
}

func TestMiddleware(t *testing.T) {
	addLayer := func(layers *[]int, layer int) {
		*layers = append(*layers, layer)
	}

	newMiddleware := func(layers *[]int, layer int) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				addLayer(layers, layer)
				next.ServeHTTP(w, req)
			})
		}
	}

	newHandler := func(layers *[]int, layer int, status int, body []byte) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			addLayer(layers, layer)
			w.WriteHeader(status)
			if len(body) > 0 {
				w.Write(body) //nolint:errcheck
			}
		}
	}

	t.Run("router, subrouter, handler", func(t *testing.T) {
		var layers []int
		r, err := NewRouter("/", WithMiddlewares(newMiddleware(&layers, 1)))
		require.NoError(t, err)

		r1, err := r.AddRouter("/foo", WithMiddlewares(newMiddleware(&layers, 2)))
		require.NoError(t, err)

		err = r1.AddHandlerFunc(
			"/bar",
			http.MethodGet,
			newHandler(&layers, 4, http.StatusOK, []byte("bar")),
			newMiddleware(&layers, 3),
		)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/bar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "bar", string(body))

		require.Equal(t, []int{1, 2, 3, 4}, layers)
	})

	t.Run("router, different subrouter, handler", func(t *testing.T) {
		var layers []int
		r, err := NewRouter("/", WithMiddlewares(newMiddleware(&layers, 1)))
		require.NoError(t, err)

		_, err = r.AddRouter("/foo", WithMiddlewares(newMiddleware(&layers, 2)))
		require.NoError(t, err)

		r2, err := r.AddRouter("/bar", WithMiddlewares(newMiddleware(&layers, 3)))
		require.NoError(t, err)

		err = r2.AddHandlerFunc(
			"/bar",
			http.MethodGet,
			newHandler(&layers, 5, http.StatusOK, []byte("bar")),
			newMiddleware(&layers, 4),
		)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/bar/bar", nil)

		layers = []int{}
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "bar", string(body))

		require.Equal(t, []int{1, 3, 4, 5}, layers)
	})

	t.Run("router, subrouter, merged method handlers", func(t *testing.T) {
		var layers []int
		r, err := NewRouter("/", WithMiddlewares(newMiddleware(&layers, 1)))
		require.NoError(t, err)

		_, err = r.AddRouter("/foo", WithMiddlewares(newMiddleware(&layers, 2)))
		require.NoError(t, err)

		r2, err := r.AddRouter("/bar", WithMiddlewares(newMiddleware(&layers, 3)))
		require.NoError(t, err)

		err = r2.AddHandlerFunc(
			"/bar",
			http.MethodGet,
			newHandler(&layers, 6, http.StatusOK, []byte("get bar")),
			newMiddleware(&layers, 4),
			newMiddleware(&layers, 5),
		)
		require.NoError(t, err)

		err = r2.AddHandlerFunc(
			"/bar",
			http.MethodPost,
			newHandler(&layers, 8, http.StatusOK, []byte("post bar")),
			newMiddleware(&layers, 7),
		)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/bar/bar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "get bar", string(body))

		require.Equal(t, []int{1, 3, 4, 5, 6}, layers)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "http://localhost:9090/bar/bar", nil)

		layers = []int{}
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "post bar", string(body))

		require.Equal(t, []int{1, 3, 7, 8}, layers)
	})

	t.Run("router, subrouter, merged method handlers from Handler", func(t *testing.T) {
		var layers []int
		r, err := NewRouter("/", WithMiddlewares(newMiddleware(&layers, 1)))
		require.NoError(t, err)

		r1, err := r.AddRouter("/tar", WithMiddlewares(newMiddleware(&layers, 2)))
		require.NoError(t, err)

		_, err = r.AddRouter("/bar", WithMiddlewares(newMiddleware(&layers, 3)))
		require.NoError(t, err)

		err = r1.AddHandler(foo.FooMulti{},
			newMiddleware(&layers, 4),
			newMiddleware(&layers, 5))
		require.NoError(t, err)

		err = r1.AddHandlerFunc(
			"/foo/foomulti",
			http.MethodDelete,
			newHandler(&layers, 8, http.StatusOK, []byte("delete foo")),
			newMiddleware(&layers, 6),
			newMiddleware(&layers, 7),
		)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/tar/foo/foomulti", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "get foomulti", string(body))

		require.Equal(t, []int{1, 2, 4, 5}, layers)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "http://localhost:9090/tar/foo/foomulti", nil)

		layers = []int{}
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "post foomulti", string(body))

		require.Equal(t, []int{1, 2, 4, 5}, layers)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodDelete, "http://localhost:9090/tar/foo/foomulti", nil)

		layers = []int{}
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "delete foo", string(body))

		require.Equal(t, []int{1, 2, 6, 7, 8}, layers)
	})

	t.Run("router, subrouter, http.Handler", func(t *testing.T) {
		var layers []int
		r, err := NewRouter("/", WithMiddlewares(newMiddleware(&layers, 1)))
		require.NoError(t, err)

		r1, err := r.AddRouter("/foo", WithMiddlewares(newMiddleware(&layers, 2)))
		require.NoError(t, err)

		err = r1.AddHTTPHandler(
			"/bar",
			newHandler(&layers, 4, http.StatusOK, []byte("bar")),
			newMiddleware(&layers, 3),
		)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/bar/baz/foobar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "bar", string(body))

		require.Equal(t, []int{1, 2, 3, 4}, layers)
	})

	t.Run("router, subrouter, not found", func(t *testing.T) {
		var layers []int
		r, err := NewRouter("/", WithMiddlewares(newMiddleware(&layers, 1)))
		require.NoError(t, err)

		r1, err := r.AddRouter("/foo", WithMiddlewares(newMiddleware(&layers, 2)))
		require.NoError(t, err)

		err = r1.AddHandlerFunc(
			"/bar",
			http.MethodGet,
			newHandler(&layers, 4, http.StatusOK, []byte("bar")),
			newMiddleware(&layers, 3),
		)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/raz", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Result().StatusCode)

		require.Equal(t, []int{1}, layers)
	})

	t.Run("router, subrouter, method not allowed", func(t *testing.T) {
		var layers []int
		r, err := NewRouter("/", WithMiddlewares(newMiddleware(&layers, 1)))
		require.NoError(t, err)

		r1, err := r.AddRouter("/foo", WithMiddlewares(newMiddleware(&layers, 2)))
		require.NoError(t, err)

		err = r1.AddHandlerFunc(
			"/bar",
			http.MethodGet,
			newHandler(&layers, 4, http.StatusOK, []byte("bar")),
			newMiddleware(&layers, 3),
		)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://localhost:9090/foo/bar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMethodNotAllowed, rec.Result().StatusCode)

		require.Equal(t, []int{1, 2}, layers)
	})

	t.Run("router, subrouter, custom not found", func(t *testing.T) {
		var layers []int
		r, err := NewRouter("/",
			WithMiddlewares(newMiddleware(&layers, 1)),
			WithNotFoundHandler(newHandler(&layers, 2, http.StatusNotFound, nil)),
		)
		require.NoError(t, err)

		r1, err := r.AddRouter("/foo", WithMiddlewares(newMiddleware(&layers, 3)))
		require.NoError(t, err)

		err = r1.AddHandlerFunc(
			"/bar",
			http.MethodGet,
			newHandler(&layers, 5, http.StatusOK, []byte("bar")),
			newMiddleware(&layers, 4),
		)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/raz", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Result().StatusCode)

		require.Equal(t, []int{1, 2}, layers)
	})

	t.Run("router, subrouter, custom method not allowed", func(t *testing.T) {
		var layers []int
		r, err := NewRouter("/", WithMiddlewares(newMiddleware(&layers, 1)))
		require.NoError(t, err)

		r1, err := r.AddRouter("/foo",
			WithMiddlewares(newMiddleware(&layers, 2)),
			WithMethodNotAllowedHandler(func(...string) http.Handler {
				return newHandler(&layers, 3, http.StatusMethodNotAllowed, nil)
			}),
		)
		require.NoError(t, err)

		err = r1.AddHandlerFunc(
			"/bar",
			http.MethodGet,
			newHandler(&layers, 5, http.StatusOK, []byte("bar")),
			newMiddleware(&layers, 4),
		)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://localhost:9090/foo/bar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMethodNotAllowed, rec.Result().StatusCode)

		require.Equal(t, []int{1, 2, 3}, layers)
	})
}

func TestLookupTags(t *testing.T) {
	type testSt struct {
		testName string
		input    reflect.StructTag
		expected []string
	}

	tests := []testSt{
		{
			testName: "single",
			input:    `path:"foo"`,
			expected: []string{"foo"},
		},
		{
			testName: "multiple",
			input:    `path:"foo1" path:"foo2"`,
			expected: []string{"foo1", "foo2"},
		},
		{
			testName: "multiple with spaces",
			input:    ` path:"foo1"   path:" foo2"    path:"foo3 "`,
			expected: []string{"foo1", " foo2", "foo3 "},
		},
		{
			testName: "multiple with spaces and quotes",
			input:    ` path:"fo\"o1" path:"fo\"\"o2", path:\"foo\"`,
			expected: []string{`fo"o1`, `fo""o2`},
		},
	}

	t.Parallel()

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			actual := lookupTags(test.input, "path")
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestBuildHandlerPath(t *testing.T) {
	type testSt struct {
		testName   string
		pkgPath    string
		structName string
		tag        reflect.StructTag
		expected   []string
	}

	tests := []testSt{
		{
			testName:   "package path + same struct name",
			pkgPath:    "/foo",
			structName: "Foo",
			expected:   []string{"/foo"},
		},
		{
			testName:   "package path + different struct name",
			pkgPath:    "/foo",
			structName: "FooBar",
			expected:   []string{"/foo/foobar"},
		},
		{
			testName:   "package path + struct name + tag absolute path",
			pkgPath:    "/foo",
			structName: "FooBar",
			tag:        `path:"/tagpath"`,
			expected:   []string{"/tagpath"},
		},
		{
			testName:   "package path + struct name + tag relative path",
			pkgPath:    "/foo",
			structName: "FooBar",
			tag:        `path:"tagpath"`,
			expected:   []string{"/foo/tagpath"},
		},
		{
			testName:   "package path + struct name + multiple tag paths",
			pkgPath:    "/foo",
			structName: "FooBar",
			tag:        `path:"tagpath" path:"/absolutetag", path:"./"`,
			expected:   []string{"/foo/tagpath", "/absolutetag", "/foo/"},
		},
	}

	t.Parallel()

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			actual := buildHandlerPaths(test.pkgPath, test.structName, test.tag)
			require.Equal(t, test.expected, actual)
		})
	}
}
