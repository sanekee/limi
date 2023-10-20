package limi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
