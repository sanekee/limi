package limi

import (
	"testing"

	"github.com/sanekee/limi/internal/testing/require"
)

func TestLabelParse(t *testing.T) {
	t.Run("helper", func(t *testing.T) {
		s := NewLabelMatcher("{foo}")
		require.Equal(t, "label:foo", s.Data())
		require.Equal(t, TypeLabel, s.Type())

		s.SetTrail('/')
		require.Equal(t, "label:foo:/", s.Data())
	})

	t.Run("parse - exact matched", func(t *testing.T) {
		s := NewLabelMatcher("{foo}")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeLabel, Str: "{foo}"})
		require.True(t, isMatched)
		require.Equal(t, "{foo}", matched)
		require.Empty(t, trail1)
		require.Empty(t, trail2)
	})

	t.Run("parse - partial matched", func(t *testing.T) {
		s := NewLabelMatcher("{foo}")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeLabel, Str: "{foobar}"})
		require.False(t, isMatched)
		require.Empty(t, matched)
		require.Equal(t, "{foobar}", trail1)
		require.Equal(t, "{foo}", trail2)
	})

	t.Run("parse - data partial matched", func(t *testing.T) {
		s := NewLabelMatcher("{foobar}")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeLabel, Str: "{foo}"})
		require.False(t, isMatched)
		require.Empty(t, matched)
		require.Equal(t, "{foo}", trail1)
		require.Equal(t, "{foobar}", trail2)
	})

	t.Run("parse - data input partial matched", func(t *testing.T) {
		s := NewLabelMatcher("{foobaz}")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeLabel, Str: "{foobar}"})
		require.False(t, isMatched)
		require.Empty(t, matched)
		require.Equal(t, "{foobar}", trail1)
		require.Equal(t, "{foobaz}", trail2)
	})

	t.Run("parse empty label", func(t *testing.T) {
		s := NewLabelMatcher("{foo}")

		isMatched, matched, trail1, trail2 := s.Parse(Parser{Type: TypeLabel, Str: ""})
		require.False(t, isMatched)
		require.Empty(t, matched)
		require.Equal(t, "", trail1)
		require.Equal(t, "{foo}", trail2)
	})
}

func TestLabelMatch(t *testing.T) {
	t.Run("match - exact matched", func(t *testing.T) {
		s := NewLabelMatcher("{foo}")

		isMatched, matched, trail1 := s.Match("foo")
		require.True(t, isMatched)
		require.Empty(t, trail1)
		require.Equal(t, "foo", matched)
	})

	t.Run("match - consumed all", func(t *testing.T) {
		s := NewLabelMatcher("{foo}")

		isMatched, matched, trail1 := s.Match("foobar")
		require.True(t, isMatched)
		require.Empty(t, trail1)
		require.Equal(t, "foobar", matched)
	})

	t.Run("match - consumed with trail", func(t *testing.T) {
		s := NewLabelMatcher("{foo}")
		s.SetTrail('b')

		isMatched, matched, trail1 := s.Match("foobar")
		require.True(t, isMatched)
		require.Equal(t, "bar", trail1)
		require.Equal(t, "foo", matched)
	})
}
