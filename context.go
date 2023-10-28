package limi

import (
	"context"

	"github.com/sanekee/limi/internal/limi"
)

func GetURLParam(ctx context.Context, key string) string {
	return limi.GetURLParam(ctx, key)
}
