package limi

import (
	"context"

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
func GetParams(ctx context.Context) (any, error) {
	return limi.GetParams(ctx)
}
