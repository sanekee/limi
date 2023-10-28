package limi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sanekee/limi/internal/testing/handler"
	"github.com/sanekee/limi/internal/testing/handler/foo"
	"github.com/sanekee/limi/internal/testing/require"
)

func TestMux(t *testing.T) {
	t.Run("global route", func(t *testing.T) {
		r, err := defaultMux.AddRouter("/")
		require.NoError(t, err)

		testFoo := foo.Foo{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		defaultMux.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("single route", func(t *testing.T) {
		m := &mux{}
		r, err := m.AddRouter("/")
		require.NoError(t, err)

		testFoo := foo.Foo{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("multi routes", func(t *testing.T) {
		m := &mux{}
		r1, err := m.AddRouter("/")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo")))
		require.NoError(t, err)

		r2, err := m.AddRouter("/")
		require.NoError(t, err)

		err = r2.AddHandlerFunc("/bar", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("bar")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/bar", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "bar", string(body))
	})

	t.Run("multi routes - same path", func(t *testing.T) {
		m := &mux{}
		r1, err := m.AddRouter("/")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo1")))
		require.NoError(t, err)

		r2, err := m.AddRouter("/")
		require.NoError(t, err)

		err = r2.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo2")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo1", string(body))
	})

	t.Run("multi routes - same path, different host", func(t *testing.T) {
		m := &mux{}
		r1, err := m.AddRouter("/", WithHosts("host1"))
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo1")))
		require.NoError(t, err)

		r2, err := m.AddRouter("/", WithHosts("host2"))
		require.NoError(t, err)

		err = r2.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo2")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://host1:9090/foo", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo1", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://host2:9090/foo", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo2", string(body))
	})

	t.Run("multi routes - same path, different methods", func(t *testing.T) {
		m := &mux{}
		r1, err := m.AddRouter("/")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo1")))
		require.NoError(t, err)

		r2, err := m.AddRouter("/")
		require.NoError(t, err)

		err = r2.AddHandlerFunc("/foo", http.MethodPost, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo2")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo1", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "http://localhost:9090/foo", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo2", string(body))
	})

	t.Run("multi routes - method not allowed", func(t *testing.T) {
		m := &mux{}
		var notAllowedRouter int
		r1, err := m.AddRouter("/", WithMethodNotAllowedHandler(func(allowed ...string) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				require.Len(t, allowed, 1)
				require.Equal(t, http.MethodGet, allowed[0])
				notAllowedRouter = 1
				w.WriteHeader(http.StatusMethodNotAllowed)
			})
		}))
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, handler.NewHandlerFunc(http.StatusOK, nil, []byte("foo1")))
		require.NoError(t, err)

		r2, err := m.AddRouter("/", WithMethodNotAllowedHandler(func(allowed ...string) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				require.Len(t, allowed, 1)
				require.Equal(t, http.MethodPost, allowed[0])
				notAllowedRouter = 2
				w.WriteHeader(http.StatusMethodNotAllowed)
			})
		}))

		require.NoError(t, err)

		err = r2.AddHandlerFunc("/bar", http.MethodPost, handler.NewHandlerFunc(http.StatusOK, nil, []byte("bar")))
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo1", string(body))

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "http://localhost:9090/foo", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMethodNotAllowed, rec.Result().StatusCode)
		require.Equal(t, 1, notAllowedRouter)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "http://localhost:9090/bar", nil)

		m.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMethodNotAllowed, rec.Result().StatusCode)
		require.Equal(t, 2, notAllowedRouter)
	})
}
