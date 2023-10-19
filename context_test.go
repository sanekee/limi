package limi

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestURLParams(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = NewContext(ctx)

		SetURLParam(ctx, "foo", "bar")

		val := GetURLParam(ctx, "foo")
		require.Equal(t, "bar", val)

	})
}
