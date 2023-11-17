package limi

import (
	"testing"

	"github.com/sanekee/limi/internal/testing/require"
)

func TestRegexpParse(t *testing.T) {
	t.Run("helper", func(t *testing.T) {
		s := NewRegexpMatcher("{foo:.*}")
		require.Equal(t, "regexp:foo:.*", s.Data())
		require.Equal(t, TypeRegexp, s.Type())

		s.SetTrail('/')
		require.Equal(t, "regexp:foo:.*:/", s.Data())
	})

	t.Run("parse - exact matched", func(t *testing.T) {
		s := NewRegexpMatcher("{foo:.*}")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeRegexp, Str: "{foo:.*}"})
		require.True(t, isMatched)
		require.Equal(t, "{foo:.*}", matched)
		require.Empty(t, trail1)
		require.Empty(t, trail2)
	})

	t.Run("parse - different regexp", func(t *testing.T) {
		s := NewRegexpMatcher("{foo:.*}")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeRegexp, Str: "{foo:[a-z]+}"})
		require.False(t, isMatched)
		require.Empty(t, matched)
		require.Equal(t, "{foo:[a-z]+}", trail1)
		require.Equal(t, "{foo:.*}", trail2)
	})

	t.Run("parse empty label", func(t *testing.T) {
		s := NewRegexpMatcher("{foo:.*}")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeRegexp, Str: ""})
		require.False(t, isMatched)
		require.Empty(t, matched)
		require.Equal(t, "", trail1)
		require.Equal(t, "{foo:.*}", trail2)
	})

	t.Run("inavalid regexp format", func(t *testing.T) {
		require.Panics(t, func() {
			NewRegexpMatcher("{foo}")
		})

	})

	t.Run("inavalid regexp", func(t *testing.T) {
		require.Panics(t, func() {
			NewRegexpMatcher("{foo:[}")
		})
	})
}

func TestRegexpMatch(t *testing.T) {
	t.Run("match - exact matched", func(t *testing.T) {
		s := NewRegexpMatcher("{foo:.*}")

		isMatched, matched, trail1 := s.Match("foo")
		require.True(t, isMatched)
		require.Empty(t, trail1)
		require.Equal(t, "foo", matched)
	})

	t.Run("match - consumed all", func(t *testing.T) {
		s := NewRegexpMatcher("{foo:.*}")

		isMatched, matched, trail1 := s.Match("foobar")
		require.True(t, isMatched)
		require.Empty(t, trail1)
		require.Equal(t, "foobar", matched)
	})

	t.Run("match - consumed with trail", func(t *testing.T) {
		s := NewRegexpMatcher("{foo:.*}")
		s.SetTrail('b')

		isMatched, matched, trail1 := s.Match("foobar")
		require.True(t, isMatched)
		require.Equal(t, "bar", trail1)
		require.Equal(t, "foo", matched)
	})

	t.Run("match - consumed with pattern", func(t *testing.T) {
		s := NewRegexpMatcher("{foo:[a-z]+}")

		isMatched, matched, trail1 := s.Match("foobar012345")
		require.True(t, isMatched)
		require.Equal(t, "012345", trail1)
		require.Equal(t, "foobar", matched)
	})
}
