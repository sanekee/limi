package limi

import (
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

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeString, Str: "foo"})
		require.True(t, isMatched)
		require.Equal(t, "foo", matched)
		require.Empty(t, trail1)
		require.Empty(t, trail2)
	})

	t.Run("parse - partial matched", func(t *testing.T) {
		s := NewStringMatcher("foo")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeString, Str: "foobar"})
		require.False(t, isMatched)
		require.Equal(t, "foo", matched)
		require.Equal(t, "bar", trail1)
		require.Empty(t, trail2)
	})

	t.Run("parse - data partial matched", func(t *testing.T) {
		s := NewStringMatcher("foobar")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeString, Str: "foo"})
		require.False(t, isMatched)
		require.Equal(t, "foo", matched)
		require.Empty(t, trail1)
		require.Equal(t, "bar", trail2)
	})

	t.Run("parse - data input partial matched", func(t *testing.T) {
		s := NewStringMatcher("foobaz")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeString, Str: "footar"})
		require.False(t, isMatched)
		require.Equal(t, "foo", matched)
		require.Equal(t, "tar", trail1)
		require.Equal(t, "baz", trail2)
	})
}

func TestStringMatch(t *testing.T) {
	t.Run("match - exact matched", func(t *testing.T) {
		s := NewStringMatcher("foo")

		isMatched, matched, trail1 := s.Match("foo")
		require.True(t, isMatched)
		require.Empty(t, trail1)
		require.Equal(t, "foo", matched)
	})

	t.Run("match - partial matched", func(t *testing.T) {
		s := NewStringMatcher("foo")

		isMatched, matched, trail1 := s.Match("foobar")
		require.True(t, isMatched)
		require.Equal(t, "bar", trail1)
		require.Equal(t, "foo", matched)
	})

	t.Run("match - data partial matched", func(t *testing.T) {
		s := NewStringMatcher("foobar")

		isMatched, matched, trail1 := s.Match("foo")
		require.False(t, isMatched)
		require.Equal(t, "foo", trail1)
		require.Empty(t, matched)
	})

	t.Run("match - data input partial matched", func(t *testing.T) {
		s := NewStringMatcher("foobaz")

		isMatched, matched, trail1 := s.Match("footar")
		require.False(t, isMatched)
		require.Equal(t, "footar", trail1)
		require.Empty(t, matched)
	})
}
