package require

import (
	"reflect"
	"sync"
	"testing"
)

func NoError(tb testing.TB, err error) {
	if err != nil {
		tb.Fail()
	}
}

func Error(tb testing.TB, err error) {
	if err == nil {
		tb.Fail()
	}
}

func Equal(tb testing.TB, expected any, actual any) {
	if !reflect.DeepEqual(expected, actual) {
		tb.Fail()
	}
}

func Len(tb testing.TB, v any, length int) {
	if getLength(tb, v) != length {
		tb.Fail()
	}
}

func True(tb testing.TB, v bool) {
	if !v {
		tb.Fail()
	}
}

func False(tb testing.TB, v bool) {
	if v {
		tb.Fail()
	}
}

func NotEmpty(tb testing.TB, v any) {
	if getLength(tb, v) <= 0 {
		tb.Fail()
	}
}

func Empty(tb testing.TB, v any) {
	if getLength(tb, v) > 0 {
		tb.Fail()
	}
}

func getLength(tb testing.TB, v any) int {
	rt := reflect.TypeOf(v)
	if rt.Kind() != reflect.Array &&
		rt.Kind() != reflect.Slice &&
		rt.Kind() != reflect.String &&
		rt.Kind() != reflect.Chan &&
		rt.Kind() != reflect.Map {
		tb.Fatalf("unsupport type for length check: %v", rt.Kind())
	}

	rv := reflect.ValueOf(v)
	return rv.Len()
}

func Nil(tb testing.TB, v any) {
	rv := reflect.ValueOf(v)
	if !rv.IsNil() {
		tb.Fail()
	}
}

func NotNil(tb testing.TB, v any) {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		tb.Fail()
	}
}

func Panics(tb testing.TB, f func()) {
	var isPanic bool
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()

			if err := recover(); err != nil {
				isPanic = true
			}
		}()
		f()
	}()

	wg.Wait()

	if !isPanic {
		tb.Fail()
	}

}
