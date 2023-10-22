package limi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sanekee/limi/internal/testing/handler/foo"
	"github.com/stretchr/testify/require"
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
			actual := findHandlerPath(r.pkgPath)
			require.Equal(t, r.expected, actual)
		})
	}

}

func TestAddHandler(t *testing.T) {
	t.Run("add with tag", func(t *testing.T) {
		r := NewRouter("/")

		testFoo := foo.Foo{}
		r.AddHandler(testFoo)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add relative path with tag", func(t *testing.T) {
		r := NewRouter("/")

		testFooRel := foo.FooRel{}
		r.AddHandler(testFooRel)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/bar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add with package path", func(t *testing.T) {
		r := NewRouter("/")

		testFooPkg := foo.FooPkg{}
		r.AddHandler(testFooPkg)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/foopkg", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add stateful handler", func(t *testing.T) {
		r := NewRouter("/")

		testFooPtr := &foo.FooPtr{}
		r.AddHandler(testFooPtr)

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
		r := NewRouter("/")

		testFoo := foo.FooHdl{}
		r.AddHandler(testFoo)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/foohdl", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add stateful method handler", func(t *testing.T) {
		r := NewRouter("/")

		testFoo := &foo.FooPtrHdl{}
		r.AddHandler(testFoo)

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
}

func TestAddRouter(t *testing.T) {
	t.Run("add sub route", func(t *testing.T) {
		r := NewRouter("/")
		r1, err := r.AddRouter("/baz")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo"))
		})
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
		r := NewRouter("/")
		_, err := r.AddRouter("/foo")
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo"))
		})
		require.Error(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Result().StatusCode)
	})

	t.Run("add handler & sub route", func(t *testing.T) {
		r := NewRouter("/")
		err := r.AddHandlerFunc("/foo", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo"))
		})
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
		r := NewRouter("/")
		r1, err := r.AddRouter("/foo")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/baz", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo/baz"))
		})
		require.NoError(t, err)

		err = r.AddHandlerFunc("/foo/bar", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo/bar"))
		})
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
		r := NewRouter("/")
		err := r.AddHandlerFunc("/foooo/bar", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foooo/bar"))
		})
		require.NoError(t, err)

		r1, err := r.AddRouter("/foo")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/baz", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo/baz"))
		})
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

	t.Run("add middleware", func(t *testing.T) {
		var middlewareAct []string
		r := NewRouter("/", WithMiddlewares(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				middlewareAct = append(middlewareAct, "middleware1")
				next.ServeHTTP(w, req)
			})
		}))
		r1, err := r.AddRouter("/foo", WithMiddlewares(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				middlewareAct = append(middlewareAct, "middleware2")
				next.ServeHTTP(w, req)
			})
		}))
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/bar", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo"))
		})
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo/bar", nil)

		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		require.EqualValues(t, []string{"middleware1", "middleware2"}, middlewareAct)
	})

	t.Run("host", func(t *testing.T) {
		r := NewRouter("/", WithHost("abc"))

		err := r.AddHandlerFunc("/foo", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			hostname := parseHost(req.URL.Host)
			w.Write([]byte("foo" + ":" + hostname))
		})
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
		r := NewRouter("/",
			WithHost("host1"),
			WithHost("host2"),
		)

		err := r.AddHandlerFunc("/foo", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			hostname := parseHost(req.URL.Host)
			w.Write([]byte("foo" + ":" + hostname))
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
		r := NewRouter("/",
			WithHost("{host:[^.]+.hostname.com}"),
		)

		err := r.AddHandlerFunc("/foo", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			hostname := parseHost(req.URL.Host)
			w.Write([]byte("foo" + ":" + hostname))
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
}
