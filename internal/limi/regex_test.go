package limi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegexp(t *testing.T) {
	t.Run("helper", func(t *testing.T) {
		s := NewRegexpMatcher("foo:.*")
		require.Equal(t, "regexp:foo:.*", s.Data())
		require.Equal(t, TypeRegexp, s.Type())

		s.SetTrail('/')
		require.Equal(t, "regexp:foo:.*:/", s.Data())
	})

	t.Run("parse - exact matched", func(t *testing.T) {
		s := NewRegexpMatcher("foo:.*")

		isMatched, matched, trail1, trail2 := s.Parse("{foo:.*}")
		require.True(t, isMatched)
		require.Equal(t, "{foo:.*}", matched)
		require.Empty(t, trail1)
		require.Empty(t, trail2)
	})

	t.Run("parse - different regexp", func(t *testing.T) {
		s := NewRegexpMatcher("foo:.*")

		isMatched, matched, trail1, trail2 := s.Parse("{foo:[a-z]+}")
		require.False(t, isMatched)
		require.Empty(t, matched)
		require.Equal(t, "{foo:[a-z]+}", trail1)
		require.Equal(t, "{foo:.*}", trail2)
	})

	t.Run("parse empty label", func(t *testing.T) {
		s := NewRegexpMatcher("foo:.*")

		isMatched, matched, trail1, trail2 := s.Parse("{}")
		require.False(t, isMatched)
		require.Empty(t, matched)
		require.Equal(t, "{}", trail1)
		require.Equal(t, "{foo:.*}", trail2)
	})

	t.Run("inavalid regexp format", func(t *testing.T) {
		require.Panics(t, func() {
			NewRegexpMatcher("foo")
		})

	})

	t.Run("inavalid regexp", func(t *testing.T) {
		require.Panics(t, func() {
			NewRegexpMatcher("foo:[")
		})
	})

	t.Run("match - exact matched", func(t *testing.T) {
		ctx := NewContext(context.Background())

		s := NewRegexpMatcher("foo:.*")

		isMatched, matched, trail1 := s.Match(ctx, "foo")
		require.True(t, isMatched)
		require.Equal(t, "foo", matched)
		require.Empty(t, trail1)
		require.Equal(t, "foo", GetURLParam(ctx, "foo"))
	})

	t.Run("match - consumed all", func(t *testing.T) {
		ctx := NewContext(context.Background())

		s := NewRegexpMatcher("foo:.*")

		isMatched, matched, trail1 := s.Match(ctx, "foobar")
		require.True(t, isMatched)
		require.Equal(t, "foobar", matched)
		require.Empty(t, trail1)
		require.Equal(t, "foobar", GetURLParam(ctx, "foo"))
	})

	t.Run("match - consumed with trail", func(t *testing.T) {
		ctx := NewContext(context.Background())

		s := NewRegexpMatcher("foo:.*")
		s.SetTrail('b')

		isMatched, matched, trail1 := s.Match(ctx, "foobar")
		require.False(t, isMatched)
		require.Equal(t, "foo", matched)
		require.Equal(t, "bar", trail1)
		require.Equal(t, "foo", GetURLParam(ctx, "foo"))
	})

	t.Run("match - consumed with pattern", func(t *testing.T) {
		ctx := NewContext(context.Background())

		s := NewRegexpMatcher("foo:[a-z]+")

		isMatched, matched, trail1 := s.Match(ctx, "foobar012345")
		require.False(t, isMatched)
		require.Equal(t, "foobar", matched)
		require.Equal(t, "012345", trail1)
		require.Equal(t, "foobar", GetURLParam(ctx, "foo"))
	})
}