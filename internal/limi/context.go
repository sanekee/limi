package limi

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

type ctxKey string

var limiContextKey ctxKey = "limi context"

type limiContext struct {
	URLParams map[string]string

	routingPath string
	paramsType  reflect.Type
}

func NewContext(ctx context.Context) context.Context {
	_, ok := ctx.Value(limiContextKey).(*limiContext)
	if ok {
		return ctx
	}

	lCtx := &limiContext{
		URLParams: make(map[string]string),
	}
	return context.WithValue(ctx, limiContextKey, lCtx)
}

func IsContextSet(ctx context.Context) bool {
	_, ok := ctx.Value(limiContextKey).(*limiContext)
	return ok
}

func ResetContext(ctx context.Context) {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return
	}

	lCtx.URLParams = make(map[string]string)
	lCtx.routingPath = ""
	lCtx.paramsType = nil
}

func GetURLParam(ctx context.Context, key string) string {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return ""
	}

	return lCtx.URLParams[key]
}

func SetURLParam(ctx context.Context, key, val string) {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return
	}

	lCtx.URLParams[key] = val
}

func GetRoutingPath(ctx context.Context) string {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return ""
	}

	return lCtx.routingPath
}

func SetRoutingPath(ctx context.Context, path string) {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return
	}

	lCtx.routingPath = path
}

type stringer interface {
	FromString(str string) error
}

func ParseURLParam(ctx context.Context, key string, data any) error {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return errors.New("invalid context")
	}

	value, ok := lCtx.URLParams[key]
	if !ok {
		return fmt.Errorf("value not found for key %s", key)
	}

	vData := reflect.ValueOf(data)
	if vData.Kind() != reflect.Pointer {
		return fmt.Errorf("data must be a pointer")
	}

	vValue := reflect.Indirect(vData)
	ptr := reflect.NewAt(vValue.Type(), unsafe.Pointer(vValue.UnsafeAddr()))
	if ptr.Type().Implements(reflect.TypeOf(((*stringer)(nil))).Elem()) {
		FromString, _ := ptr.Type().MethodByName("FromString")
		outs := FromString.Func.Call([]reflect.Value{ptr, reflect.ValueOf(value)})
		err, _ := outs[0].Interface().(error)
		if err != nil {
			return fmt.Errorf("failed to parse custom value %s %w", value, err)
		}
	}
	return fromString(value, vValue)
}

func ParseURLParams(ctx context.Context, data any) error {
	vData := reflect.ValueOf(data)
	if vData.Kind() != reflect.Pointer {
		return fmt.Errorf("data must be a pointer")
	}

	vValue := reflect.Indirect(vData)
	if vValue.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct")
	}

	for i := 0; i < vValue.NumField(); i++ {
		tField := vValue.Type().Field(i)
		vField := vValue.Field(i)

		param := getParam(tField)
		if param == "" {
			continue
		}

		ptr := reflect.NewAt(vField.Type(), unsafe.Pointer(vField.UnsafeAddr()))
		if err := ParseURLParam(ctx, param, ptr.Interface()); err != nil {
			return fmt.Errorf("failed to parse param %s %w", param, err)
		}
	}
	return nil

}

func fromString(str string, val reflect.Value) error {
	switch val.Type().Kind() {
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		pValue, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s, value %s", str, val.Type().Kind())
		}
		val.Set(reflect.ValueOf(pValue).Convert(val.Type()))
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		pValue, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s, value %s", str, val.Type().Kind())
		}
		val.Set(reflect.ValueOf(pValue).Convert(val.Type()))
	case reflect.Float64:
		pValue, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s, value %s", str, val.Type().Kind())
		}
		val.Set(reflect.ValueOf(pValue))
	case reflect.Float32:
		pValue, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s, value %s", str, val.Type().Kind())
		}
		val.Set(reflect.ValueOf(pValue).Convert(val.Type()))
	case reflect.Bool:
		pValue, err := strconv.ParseBool(str)
		if err != nil {
			return fmt.Errorf("failed to convert %s, value %s", str, val.Type().Kind())
		}
		val.Set(reflect.ValueOf(pValue))
	case reflect.String:
		val.Set(reflect.ValueOf(str))
	default:
		return fmt.Errorf("unsupported type %s", val.Type().Kind())
	}

	return nil
}

func getParam(field reflect.StructField) string {
	limiTag := field.Tag.Get("limi")
	if limiTag == "" {
		return ""
	}

	strs := strings.Split(limiTag, "=")
	if len(strs) < 1 ||
		strs[0] != "param" {
		return ""
	}

	if len(strs) == 1 {
		return field.Name
	}

	return strs[1]
}

func SetParamsType(ctx context.Context, t reflect.Type) error {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return fmt.Errorf("invalid context")
	}

	lCtx.paramsType = t
	return nil
}

func SetParamsData(ctx context.Context, data any) error {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return fmt.Errorf("invalid context")
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		v = reflect.Indirect(v)
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct, %v", v.Kind())
	}

	lCtx.paramsType = v.Type()
	return nil
}

func GetParams(ctx context.Context) (any, error) {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return nil, fmt.Errorf("invalid context")
	}

	if lCtx.paramsType == nil {
		return nil, fmt.Errorf("unknown context params type")
	}

	v := reflect.New(lCtx.paramsType)
	if err := ParseURLParams(ctx, v.Interface()); err != nil {
		return nil, fmt.Errorf("failed to parse params %w", err)
	}
	return reflect.Indirect(v).Interface(), nil
}
