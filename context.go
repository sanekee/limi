package limi

import (
	"context"
	"fmt"

	"github.com/sanekee/limi/internal/limi"
)

// GetURLParam get value set by label matched in url
func GetURLParam(ctx context.Context, key string) string {
	return limi.GetURLParam(ctx, key)
}

// GetURLParam set data to the value set by label matched in url
func ParseURLParam(ctx context.Context, key string, data any) error {
	return limi.ParseURLParam(ctx, key, data)
}

// SetParamsData set context params type for parsing
func SetParamsData(ctx context.Context, data any) error {
	return limi.SetParamsData(ctx, data)
}

// GetParams get context params parsed from URL
func GetParams[T any](ctx context.Context) (T, error) {
	var ret T
	data, err := limi.GetParams(ctx)
	if err != nil {
		return ret, err
	}

	ret, ok := data.(T)
	if !ok {
		return ret, fmt.Errorf("failed to convert params type")
	}
	return ret, nil
}
