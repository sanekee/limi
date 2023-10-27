package limi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sanekee/limi/internal/testing/handler/foo"
	"github.com/stretchr/testify/require"
)

func TestMux(t *testing.T) {
	t.Run("global route", func(t *testing.T) {
		r, err := AddRouter("/")
		require.NoError(t, err)

		testFoo := foo.Foo{}
		err = r.AddHandler(testFoo)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:9090/foo", nil)

		ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("single route", func(t *testing.T) {
		m := &mux{}
		r, err := m.addRouter("/")
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
		r1, err := m.addRouter("/")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo")) //nolint:errcheck
		}))
		require.NoError(t, err)

		r2, err := m.addRouter("/")
		require.NoError(t, err)

		err = r2.AddHandlerFunc("/bar", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("bar")) //nolint:errcheck
		}))
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
		r1, err := m.addRouter("/")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo1")) //nolint:errcheck
		}))
		require.NoError(t, err)

		r2, err := m.addRouter("/")
		require.NoError(t, err)

		err = r2.AddHandlerFunc("/foo", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo2")) //nolint:errcheck
		}))
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
		r1, err := m.addRouter("/", WithHost("host1"))
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo1")) //nolint:errcheck
		}))
		require.NoError(t, err)

		r2, err := m.addRouter("/", WithHost("host2"))
		require.NoError(t, err)

		err = r2.AddHandlerFunc("/foo", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo2")) //nolint:errcheck
		}))
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
		r1, err := m.addRouter("/")
		require.NoError(t, err)

		err = r1.AddHandlerFunc("/foo", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo1")) //nolint:errcheck
		}))
		require.NoError(t, err)

		r2, err := m.addRouter("/")
		require.NoError(t, err)

		err = r2.AddHandlerFunc("/foo", http.MethodPost, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo2")) //nolint:errcheck
		}))
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
}
