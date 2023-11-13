package limi

import (
	"context"
	"testing"

	"github.com/sanekee/limi/internal/testing/require"
)

func TestStringParse(t *testing.T) {
	t.Run("helper", func(t *testing.T) {
		s := NewStringMatcher("foo")
		require.Equal(t, "foo", s.Data())
		require.Equal(t, TypeString, s.Type())
	})

	t.Run("parse - exact matched", func(t *testing.T) {
		s := NewStringMatcher("foo")

		isMatched, matched, trail1, trail2 := s.Parse("foo")
		require.True(t, isMatched)
		require.Equal(t, "foo", matched)
		require.Empty(t, trail1)
		require.Empty(t, trail2)
	})

	t.Run("parse - partial matched", func(t *testing.T) {
		s := NewStringMatcher("foo")

		isMatched, matched, trail1, trail2 := s.Parse("foobar")
		require.False(t, isMatched)
		require.Equal(t, "foo", matched)
		require.Equal(t, "bar", trail1)
		require.Empty(t, trail2)
	})

	t.Run("parse - data partial matched", func(t *testing.T) {
		s := NewStringMatcher("foobar")

		isMatched, matched, trail1, trail2 := s.Parse("foo")
		require.False(t, isMatched)
		require.Equal(t, "foo", matched)
		require.Empty(t, trail1)
		require.Equal(t, "bar", trail2)
	})

	t.Run("parse - data input partial matched", func(t *testing.T) {
		s := NewStringMatcher("foobaz")

		isMatched, matched, trail1, trail2 := s.Parse("footar")
		require.False(t, isMatched)
		require.Equal(t, "foo", matched)
		require.Equal(t, "tar", trail1)
		require.Equal(t, "baz", trail2)
	})
}

func TestStringMatch(t *testing.T) {
	t.Run("match - exact matched", func(t *testing.T) {
		ctx := context.Background()

		s := NewStringMatcher("foo")

		isMatched, trail1 := s.Match(ctx, "foo")
		require.True(t, isMatched)
		require.Empty(t, trail1)
	})

	t.Run("match - partial matched", func(t *testing.T) {
		ctx := context.Background()

		s := NewStringMatcher("foo")

		isMatched, trail1 := s.Match(ctx, "foobar")
		require.True(t, isMatched)
		require.Equal(t, "bar", trail1)
	})

	t.Run("match - data partial matched", func(t *testing.T) {
		ctx := context.Background()

		s := NewStringMatcher("foobar")

		isMatched, trail1 := s.Match(ctx, "foo")
		require.False(t, isMatched)
		require.Equal(t, "foo", trail1)
	})

	t.Run("match - data input partial matched", func(t *testing.T) {
		ctx := context.Background()

		s := NewStringMatcher("foobaz")

		isMatched, trail1 := s.Match(ctx, "footar")
		require.False(t, isMatched)
		require.Equal(t, "footar", trail1)
	})
}
