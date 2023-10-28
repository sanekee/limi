package limi

import (
	"testing"

	"github.com/sanekee/limi/internal/testing/require"
)

func TestSplit(t *testing.T) {
	t.Run("split string", func(t *testing.T) {
		str := "a string"
		res, err := SplitParsers(str)

		require.NoError(t, err)
		require.Len(t, res, 1)

		require.Equal(t, str, res[0].Str)
		require.Equal(t, TypeString, res[0].Type)
	})

	t.Run("label only", func(t *testing.T) {
		str := "{label}"
		res, err := SplitParsers(str)

		require.NoError(t, err)
		require.Len(t, res, 1)

		require.Equal(t, "label", res[0].Str)
		require.Equal(t, TypeLabel, res[0].Type)
	})

	t.Run("multiple labels only", func(t *testing.T) {
		str := "{label1}{label2}"
		res, err := SplitParsers(str)

		require.NoError(t, err)
		require.Len(t, res, 2)

		require.Equal(t, "label1", res[0].Str)
		require.Equal(t, TypeLabel, res[0].Type)
		require.Equal(t, "label2", res[1].Str)
		require.Equal(t, TypeLabel, res[1].Type)
	})

	t.Run("split string with label", func(t *testing.T) {
		str := "a string{and label}"
		res, err := SplitParsers(str)

		require.NoError(t, err)
		require.Len(t, res, 2)

		require.Equal(t, "a string", res[0].Str)
		require.Equal(t, TypeString, res[0].Type)
		require.Equal(t, "and label", res[1].Str)
		require.Equal(t, TypeLabel, res[1].Type)
	})

	t.Run("error with invalid string", func(t *testing.T) {
		str := "a string{and invalid label"
		res, err := SplitParsers(str)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("regexp only", func(t *testing.T) {
		str := "{regexp:[\\d]+}"
		res, err := SplitParsers(str)

		require.NoError(t, err)
		require.Len(t, res, 1)

		require.Equal(t, "regexp:[\\d]+", res[0].Str)
		require.Equal(t, TypeRegexp, res[0].Type)
	})

	t.Run("multiple regexps only", func(t *testing.T) {
		str := "{regexp1:[a-z]+}{regexp2:[0-9]+}"
		res, err := SplitParsers(str)

		require.NoError(t, err)
		require.Len(t, res, 2)

		require.Equal(t, "regexp1:[a-z]+", res[0].Str)
		require.Equal(t, TypeRegexp, res[0].Type)
		require.Equal(t, "regexp2:[0-9]+", res[1].Str)
		require.Equal(t, TypeRegexp, res[1].Type)
	})

	t.Run("split string with regexp", func(t *testing.T) {
		str := "a string{and regexp:[a-z]+}"
		res, err := SplitParsers(str)

		require.NoError(t, err)
		require.Len(t, res, 2)

		require.Equal(t, "a string", res[0].Str)
		require.Equal(t, TypeString, res[0].Type)
		require.Equal(t, "and regexp:[a-z]+", res[1].Str)
		require.Equal(t, TypeRegexp, res[1].Type)
	})

	t.Run("split string with labels & regexp", func(t *testing.T) {
		str := "a string{and label}{and regexp:[a-z]+}"
		res, err := SplitParsers(str)

		require.NoError(t, err)
		require.Len(t, res, 3)

		require.Equal(t, "a string", res[0].Str)
		require.Equal(t, TypeString, res[0].Type)
		require.Equal(t, "and label", res[1].Str)
		require.Equal(t, TypeLabel, res[1].Type)
		require.Equal(t, "and regexp:[a-z]+", res[2].Str)
		require.Equal(t, TypeRegexp, res[2].Type)
	})

	t.Run("error with invalid string", func(t *testing.T) {
		str := "a string{and invalid regexp:[absd"
		res, err := SplitParsers(str)

		require.Error(t, err)
		require.Nil(t, res)
	})
}
