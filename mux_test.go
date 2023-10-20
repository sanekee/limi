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
	t.Run("static route", func(t *testing.T) {
		fooHandler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo"))
		}
		barHandler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("bar"))
		}
		mux := newMux("/")
		mux.AddHandler("/foo", HandlerType{
			Handlers: map[string]http.Handler{
				"GET": http.HandlerFunc(fooHandler),
			},
		})
		mux.AddHandler("/bar", HandlerType{
			Handlers: map[string]http.Handler{
				"GET": http.HandlerFunc(barHandler),
			},
		})
		svr := httptest.NewServer(mux.Serve())

		req := httptest.NewRequest(http.MethodGet, svr.URL+"/foo", nil)
		res, err := http.DefaultTransport.RoundTrip(req)

		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, http.StatusOK, res.StatusCode)

		buf, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(buf))

		req = httptest.NewRequest(http.MethodGet, svr.URL+"/bar", nil)
		res, err = http.DefaultTransport.RoundTrip(req)

		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, http.StatusOK, res.StatusCode)

		buf, err = io.ReadAll(res.Body)
		require.NoError(t, err)
		require.Equal(t, "bar", string(buf))
	})

	t.Run("static route - level 2", func(t *testing.T) {
		fooHandler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("foo"))
		}
		barHandler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("bar"))
		}
		mux := newMux("/")
		mux.AddHandler("/foo", HandlerType{
			Handlers: map[string]http.Handler{
				"GET": http.HandlerFunc(fooHandler),
			},
		})
		mux.AddHandler("/foo/bar", HandlerType{
			Handlers: map[string]http.Handler{
				"GET": http.HandlerFunc(barHandler),
			},
		})
		svr := httptest.NewServer(mux.Serve())

		req := httptest.NewRequest(http.MethodGet, svr.URL+"/foo", nil)
		res, err := http.DefaultTransport.RoundTrip(req)

		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, http.StatusOK, res.StatusCode)

		buf, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(buf))

		req = httptest.NewRequest(http.MethodGet, svr.URL+"/foo/bar", nil)
		res, err = http.DefaultTransport.RoundTrip(req)

		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, http.StatusOK, res.StatusCode)

		buf, err = io.ReadAll(res.Body)
		require.NoError(t, err)
		require.Equal(t, "bar", string(buf))
	})

}

func TestAddHandler(t *testing.T) {
	t.Run("add with tag", func(t *testing.T) {
		testFoo := foo.Foo{}
		AddHandler(testFoo)

		mux := Mux()
		require.NotNil(t, mux)

		h, ok := mux.handlers["/foo"]
		require.True(t, ok)

		hd, ok := h.Handlers["get"]
		require.True(t, ok)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		hd.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add relative path with tag", func(t *testing.T) {
		testFooRel := foo.FooRel{}
		AddHandler(testFooRel)

		mux := Mux()
		require.NotNil(t, mux)

		h, ok := mux.handlers["/foo/bar"]
		require.True(t, ok)

		hd, ok := h.Handlers["get"]
		require.True(t, ok)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		hd.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add with package path", func(t *testing.T) {
		testFooPkg := foo.FooPkg{}
		AddHandler(testFooPkg)

		mux := Mux()
		require.NotNil(t, mux)

		h, ok := mux.handlers["/foo/foopkg"]
		require.True(t, ok)

		hd, ok := h.Handlers["get"]
		require.True(t, ok)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		hd.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add stateful handler", func(t *testing.T) {
		testFooPtr := &foo.FooPtr{}
		AddHandler(testFooPtr)

		mux := Mux()
		require.NotNil(t, mux)

		h, ok := mux.handlers["/foo/fooptr"]
		require.True(t, ok)

		hd, ok := h.Handlers["get"]
		require.True(t, ok)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		hd.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		count := rec.Result().Header.Get("X-Count-Total")
		require.Equal(t, "1", count)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/", nil)

		hd.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		count = rec.Result().Header.Get("X-Count-Total")
		require.Equal(t, "2", count)
	})

	t.Run("add with handler method", func(t *testing.T) {
		testFoo := foo.FooHdl{}
		AddHandler(testFoo)

		mux := Mux()
		require.NotNil(t, mux)

		h, ok := mux.handlers["/foo/foohdl"]
		require.True(t, ok)

		hd, ok := h.Handlers["get"]
		require.True(t, ok)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		hd.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))
	})

	t.Run("add stateful method handler", func(t *testing.T) {
		testFoo := &foo.FooPtrHdl{}
		AddHandler(testFoo)

		mux := Mux()
		require.NotNil(t, mux)

		h, ok := mux.handlers["/foo/fooptrhdl"]
		require.True(t, ok)

		hd, ok := h.Handlers["get"]
		require.True(t, ok)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		hd.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err := io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		count := rec.Result().Header.Get("X-Count-Total")
		require.Equal(t, "1", count)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/", nil)

		hd.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		body, err = io.ReadAll(rec.Body)
		require.NoError(t, err)
		require.Equal(t, "foo", string(body))

		count = rec.Result().Header.Get("X-Count-Total")
		require.Equal(t, "2", count)
	})
}
