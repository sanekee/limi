package limi

import (
	"context"
	"net/url"
	"strconv"
	"testing"

	"github.com/sanekee/limi/internal/testing/require"
)

func TestContext(t *testing.T) {
	t.Run("new context", func(t *testing.T) {
		ctx := NewContext(context.Background())
		actual := IsContextSet(ctx)
		require.True(t, actual)
	})

	t.Run("new context from existing", func(t *testing.T) {
		ctx := NewContext(context.Background())
		actual := IsContextSet(ctx)
		require.True(t, actual)

		SetURLParam(ctx, "foo", "bar")

		ctx = NewContext(ctx)
		actualStr := GetURLParam(ctx, "foo")
		require.Equal(t, "bar", actualStr)
	})

	t.Run("reset context", func(t *testing.T) {
		ctx := NewContext(context.Background())
		actual := IsContextSet(ctx)
		require.True(t, actual)

		SetURLParam(ctx, "foo", "bar")
		SetRoutingPath(ctx, "foobar")

		ResetContext(ctx)

		actualStr := GetURLParam(ctx, "foo")
		require.Empty(t, actualStr)
		require.Empty(t, GetRoutingPath(ctx))
	})

	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()

		SetURLParam(ctx, "foo", "bar")
		require.Empty(t, GetURLParam(ctx, "foo"))

		SetRoutingPath(ctx, "foobar")
		require.Empty(t, GetRoutingPath(ctx))
	})
}

func TestURLParams(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "bar")

		val := GetURLParam(ctx, "foo")
		require.Equal(t, "bar", val)

	})
}

func TestParseURLParam(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "bar")

		var actual string
		err := ParseURLParam(ctx, "foo", &actual)
		require.NoError(t, err)
		require.Equal(t, "bar", actual)
	})

	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()

		var actual string
		err := ParseURLParam(ctx, "foo", &actual)
		require.Error(t, err)
		require.Empty(t, actual)
	})

	t.Run("missing key", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "bar")

		var actual string
		err := ParseURLParam(ctx, "foo1", &actual)
		require.Error(t, err)
		require.Empty(t, actual)
	})

	t.Run("invalid data", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "bar")

		var actual string
		err := ParseURLParam(ctx, "foo", actual)
		require.Error(t, err)
	})

	t.Run("int", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168")

		var actual int
		err := ParseURLParam(ctx, "foo", &actual)
		require.NoError(t, err)
		require.Equal(t, 168, actual)
	})

	t.Run("invalid int", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168a")

		var actual int
		err := ParseURLParam(ctx, "foo", &actual)
		require.Error(t, err)
	})

	t.Run("uint", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168")

		var actual uint
		err := ParseURLParam(ctx, "foo", &actual)
		require.NoError(t, err)
		require.Equal(t, uint(168), actual)
	})

	t.Run("invalid uint", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168a")

		var actual uint
		err := ParseURLParam(ctx, "foo", &actual)
		require.Error(t, err)
	})

	t.Run("float32", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168")

		var actual float32
		err := ParseURLParam(ctx, "foo", &actual)
		require.NoError(t, err)
		require.Equal(t, float32(168), actual)
	})

	t.Run("invalid float32", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168a")

		var actual float32
		err := ParseURLParam(ctx, "foo", &actual)
		require.Error(t, err)
	})

	t.Run("float64", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168")

		var actual float64
		err := ParseURLParam(ctx, "foo", &actual)
		require.NoError(t, err)
		require.Equal(t, float64(168), actual)
	})

	t.Run("invalid float64", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168a")

		var actual float64
		err := ParseURLParam(ctx, "foo", &actual)
		require.Error(t, err)
	})

	t.Run("bool", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "true")

		var actual bool
		err := ParseURLParam(ctx, "foo", &actual)
		require.NoError(t, err)
		require.Equal(t, true, actual)
	})

	t.Run("invalid bool", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168a")

		var actual bool
		err := ParseURLParam(ctx, "foo", &actual)
		require.Error(t, err)
	})

	t.Run("custom stringer (int)", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168")

		var actual customStringer
		err := ParseURLParam(ctx, "foo", &actual)
		require.NoError(t, err)
		require.Equal(t, 168, int(actual))
	})

	t.Run("invalid custom stringer (int)", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168a")

		var actual customStringer
		err := ParseURLParam(ctx, "foo", &actual)
		require.Error(t, err)
	})

	t.Run("custom stuct", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "foo", "168")

		var actual customStruct
		err := ParseURLParam(ctx, "foo", &actual)
		require.Error(t, err)
	})
}

type customStringer int

func (c *customStringer) FromString(str string) error {
	v, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	*(*int)(c) = v
	return nil
}

type customStruct struct{}

type testParams struct {
	id        uint64 `limi:"param"`
	index     int    `limi:"param=idx"`
	slugPath  string `limi:"param=slug"`
	Exported  string `limi:"param=exported"`
	offset    uint64 `limi:"query"`
	size      int    `limi:"query=size"`
	filter    string `limi:"query=f"`
	Sex       string `limi:"query=sex"`
	skipField string
}

func TestParseURLParams(t *testing.T) {
	t.Run("parse", func(t *testing.T) {
		ctx := NewContext(context.Background())

		SetURLParam(ctx, "id", "168")
		SetURLParam(ctx, "idx", "420")
		SetURLParam(ctx, "slug", "my-path")
		SetURLParam(ctx, "exported", "my-export")
		SetURLParam(ctx, "skipField", "my-skip")
		SetQueries(ctx, url.Values{
			"offset": []string{"6"},
			"size":   []string{"9"},
			"f":      []string{"name"},
			"sex":    []string{"F"},
		})

		expected := testParams{
			id:       168,
			index:    420,
			slugPath: "my-path",
			Exported: "my-export",
			offset:   6,
			size:     9,
			filter:   "name",
			Sex:      "F",
		}
		var actual testParams
		err := ParseURLParams(ctx, &actual)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
		require.Empty(t, actual.skipField)
	})

	t.Run("not pointer", func(t *testing.T) {
		ctx := NewContext(context.Background())

		var actual testParams
		err := ParseURLParams(ctx, actual)
		require.Error(t, err)
	})

	t.Run("not struct", func(t *testing.T) {
		ctx := NewContext(context.Background())

		var actual int
		err := ParseURLParams(ctx, actual)
		require.Error(t, err)
	})

	t.Run("params", func(t *testing.T) {
		ctx := NewContext(context.Background())
		SetParamsData(ctx, testParams{})

		SetURLParam(ctx, "id", "168")
		SetURLParam(ctx, "idx", "420")
		SetURLParam(ctx, "slug", "my-path")
		SetURLParam(ctx, "exported", "my-export")
		SetURLParam(ctx, "skipField", "my-skip")
		SetQueries(ctx, url.Values{
			"offset": []string{"6"},
			"size":   []string{"9"},
			"f":      []string{"name"},
			"sex":    []string{"F"},
		})

		expected := testParams{
			id:       168,
			index:    420,
			slugPath: "my-path",
			Exported: "my-export",
			offset:   6,
			size:     9,
			filter:   "name",
			Sex:      "F",
		}

		params, err := GetParams(ctx)
		require.NoError(t, err)
		actual, ok := params.(testParams)
		require.True(t, ok)
		require.Equal(t, expected, actual)
		require.Empty(t, actual.skipField)
	})
}
