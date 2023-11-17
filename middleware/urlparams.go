package middleware

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/sanekee/limi/internal/limi"
)

func SetURLParamsData(data any) (func(http.Handler) http.Handler, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		v = reflect.Indirect(v)
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("data must be a struct, %v", v.Kind())
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			limi.SetParamsType(req.Context(), v.Type())

			next.ServeHTTP(w, req)
		})
	}, nil
}
