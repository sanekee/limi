package limi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildPath(t *testing.T) {
	type test struct {
		host     string
		parent   string
		path     string
		expected string
	}

	tests := []test{
		{
			host:     "localhost",
			parent:   "/parent/",
			path:     "/foo",
			expected: "localhost/parent/foo",
		},
		{
			host:     "localhost/",
			parent:   "/parent",
			path:     "/foo/",
			expected: "localhost/parent/foo/",
		},
		{
			host:     "localhost",
			parent:   "/",
			path:     "/foo/",
			expected: "localhost/foo/",
		},
		{
			host:     "localhost",
			parent:   "",
			path:     "",
			expected: "localhost",
		},
		{
			host:     "localhost",
			parent:   "",
			path:     "foo",
			expected: "localhost/foo",
		},
		{
			host:     "",
			parent:   "/",
			path:     "/",
			expected: "/",
		},
		{
			host:     "",
			parent:   "/",
			path:     "/foo",
			expected: "/foo",
		},
	}

	t.Parallel()

	for _, r := range tests {
		t.Run("", func(t *testing.T) {
			actual := buildPath(r.host, r.parent, r.path)
			require.Equal(t, r.expected, actual)
		})
	}

}

func TestFindHandlerPath(t *testing.T) {
	type test struct {
		pkgPath  string
		expected string
	}

	tests := []test{
		{
			pkgPath:  "base/handler/foo/",
			expected: "/foo/",
		},
		{
			pkgPath:  "base/handler/handlerfoo/",
			expected: "/handler/foo/",
		},
		{
			pkgPath:  "/foo",
			expected: "/foo",
		},
	}

	t.Parallel()

	for _, r := range tests {
		t.Run("", func(t *testing.T) {
			actual := findHandlerPath(r.pkgPath)
			require.Equal(t, r.expected, actual)
		})
	}

}
