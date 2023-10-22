package limi

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type routePath struct {
	path     string
	children []*routePath
}

func TestStaticRoute(t *testing.T) {
	t.Run("new branch", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo", funcHandler(func() string { return "i'm /foo" }))
		require.NoError(t, err)

		err = root.Insert("/bar", funcHandler(func() string { return "i'm /bar" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/",
			children: []*routePath{
				{
					path: "foo",
				},
				{
					path: "bar",
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo", h1())

		h2 := lookupFunc(root.Lookup(ctx, "/bar"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /bar", h2())

		h3 := lookupFunc(root.Lookup(ctx, "/baz"))
		require.Nil(t, h3)

	})

	t.Run("split longer parent", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/bar", funcHandler(func() string { return "i'm /foo/bar" }))
		require.NoError(t, err)

		err = root.Insert("/foo", funcHandler(func() string { return "i'm /foo" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo",
			children: []*routePath{
				{
					path: "/bar",
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo", h1())

		h2 := lookupFunc(root.Lookup(ctx, "/foo/bar"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/bar", h2())
	})

	t.Run("new child", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo", funcHandler(func() string { return "i'm /foo" }))
		require.NoError(t, err)

		err = root.Insert("/foo/bar", funcHandler(func() string { return "i'm /foo/bar" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo",
			children: []*routePath{
				{
					path: "/bar",
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo", h1())

		h2 := lookupFunc(root.Lookup(ctx, "/foo/bar"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/bar", h2())
	})

	t.Run("add presplit handle", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/bar", funcHandler(func() string { return "i'm /foo/bar" }))
		require.NoError(t, err)

		err = root.Insert("/foo/car", funcHandler(func() string { return "i'm /foo/car" }))
		require.NoError(t, err)

		err = root.Insert("/foo/", funcHandler(func() string { return "i'm /foo/" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo/",
			children: []*routePath{
				{
					path: "bar",
				},
				{
					path: "car",
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo/"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo/", h1())

		h2 := lookupFunc(root.Lookup(ctx, "/foo/bar"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/bar", h2())

		h3 := lookupFunc(root.Lookup(ctx, "/foo/car"))
		require.NotNil(t, h3)
		require.Equal(t, "i'm /foo/car", h3())
	})

	t.Run("failed with duplicated handle", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo", funcHandler(func() string { return "i'm /foo" }))
		require.NoError(t, err)

		err = root.Insert("/foo", funcHandler(func() string { return "i'm /foo too" }))
		require.Error(t, err)

		expected := &routePath{
			path: "/foo",
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo", h1())

	})

	t.Run("new branches", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/bar", funcHandler(func() string { return "i'm /foo/bar" }))
		require.NoError(t, err)

		err = root.Insert("/foo", funcHandler(func() string { return "i'm /foo" }))
		require.NoError(t, err)

		err = root.Insert("/bar/", funcHandler(func() string { return "i'm /bar/" }))
		require.NoError(t, err)

		err = root.Insert("/bar/foo", funcHandler(func() string { return "i'm /bar/foo" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/",
			children: []*routePath{
				{
					path: "foo",
					children: []*routePath{
						{
							path: "/bar",
						},
					},
				},
				{
					path: "bar/",
					children: []*routePath{
						{
							path: "foo",
						},
					},
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo", h1())

		h2 := lookupFunc(root.Lookup(ctx, "/foo/bar"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/bar", h2())

		h3 := lookupFunc(root.Lookup(ctx, "/bar/"))
		require.NotNil(t, h3)
		require.Equal(t, "i'm /bar/", h3())

		h4 := lookupFunc(root.Lookup(ctx, "/bar/foo"))
		require.NotNil(t, h4)
		require.Equal(t, "i'm /bar/foo", h4())
	})

	t.Run("long branch", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("abcdefg", funcHandler(func() string { return "i'm abcdefg" }))
		require.NoError(t, err)

		err = root.Insert("abcdef", funcHandler(func() string { return "i'm abcdef" }))
		require.NoError(t, err)

		err = root.Insert("abcde", funcHandler(func() string { return "i'm abcde" }))
		require.NoError(t, err)

		err = root.Insert("abcd", funcHandler(func() string { return "i'm abcd" }))
		require.NoError(t, err)

		err = root.Insert("abc", funcHandler(func() string { return "i'm abc" }))
		require.NoError(t, err)

		err = root.Insert("ab", funcHandler(func() string { return "i'm ab" }))
		require.NoError(t, err)

		err = root.Insert("a", funcHandler(func() string { return "i'm a" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "a",
			children: []*routePath{{path: "b",
				children: []*routePath{{path: "c",
					children: []*routePath{{path: "d",
						children: []*routePath{{path: "e",
							children: []*routePath{{path: "f",
								children: []*routePath{{path: "g"}},
							}},
						}},
					}},
				}},
			}},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		root.Walk(func(level int, str string, h any) {
			hf := lookupFunc(h, "")
			require.NotNil(t, hf)
			t.Log(level, str, hf())
		})

		h1 := lookupFunc(root.Lookup(ctx, "a"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm a", h1())

		h2 := lookupFunc(root.Lookup(ctx, "ab"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm ab", h2())

		h3 := lookupFunc(root.Lookup(ctx, "abc"))
		require.NotNil(t, h3)
		require.Equal(t, "i'm abc", h3())

		h4 := lookupFunc(root.Lookup(ctx, "abcd"))
		require.NotNil(t, h4)
		require.Equal(t, "i'm abcd", h4())

		h5 := lookupFunc(root.Lookup(ctx, "abcde"))
		require.NotNil(t, h5)
		require.Equal(t, "i'm abcde", h5())

		h6 := lookupFunc(root.Lookup(ctx, "abcdef"))
		require.NotNil(t, h6)
		require.Equal(t, "i'm abcdef", h6())

		h7 := lookupFunc(root.Lookup(ctx, "abcdefg"))
		require.NotNil(t, h7)
		require.Equal(t, "i'm abcdefg", h7())

	})

	t.Run("long branch reverse", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("a", funcHandler(func() string { return "i'm a" }))
		require.NoError(t, err)

		err = root.Insert("ab", funcHandler(func() string { return "i'm ab" }))
		require.NoError(t, err)

		err = root.Insert("abc", funcHandler(func() string { return "i'm abc" }))
		require.NoError(t, err)

		err = root.Insert("abcd", funcHandler(func() string { return "i'm abcd" }))
		require.NoError(t, err)

		err = root.Insert("abcde", funcHandler(func() string { return "i'm abcde" }))
		require.NoError(t, err)

		err = root.Insert("abcdef", funcHandler(func() string { return "i'm abcdef" }))
		require.NoError(t, err)

		err = root.Insert("abcdefg", funcHandler(func() string { return "i'm abcdefg" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "a",
			children: []*routePath{{path: "b",
				children: []*routePath{{path: "c",
					children: []*routePath{{path: "d",
						children: []*routePath{{path: "e",
							children: []*routePath{{path: "f",
								children: []*routePath{{path: "g"}},
							}},
						}},
					}},
				}},
			}},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		root.Walk(func(level int, str string, h any) {
			hf := lookupFunc(h, "")
			require.NotNil(t, hf)
			t.Log(level, str, hf())
		})

		h1 := lookupFunc(root.Lookup(ctx, "a"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm a", h1())

		h2 := lookupFunc(root.Lookup(ctx, "ab"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm ab", h2())

		h3 := lookupFunc(root.Lookup(ctx, "abc"))
		require.NotNil(t, h3)
		require.Equal(t, "i'm abc", h3())

		h4 := lookupFunc(root.Lookup(ctx, "abcd"))
		require.NotNil(t, h4)
		require.Equal(t, "i'm abcd", h4())

		h5 := lookupFunc(root.Lookup(ctx, "abcde"))
		require.NotNil(t, h5)
		require.Equal(t, "i'm abcde", h5())

		h6 := lookupFunc(root.Lookup(ctx, "abcdef"))
		require.NotNil(t, h6)
		require.Equal(t, "i'm abcdef", h6())

		h7 := lookupFunc(root.Lookup(ctx, "abcdefg"))
		require.NotNil(t, h7)
		require.Equal(t, "i'm abcdefg", h7())
	})
}

func TestLabelRoute(t *testing.T) {
	t.Run("new label", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/{id}", funcHandler(func() string { return "i'm /foo/{id}" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo/",
			children: []*routePath{
				{
					path: "label:id",
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo/"))
		require.Nil(t, h1)

		ctx = NewContext(ctx)

		h2 := lookupFunc(root.Lookup(ctx, "/foo/123"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/{id}", h2())
		require.Equal(t, "123", GetURLParam(ctx, "id"))

	})

	t.Run("label only", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("{id}", funcHandler(func() string { return "i'm {id}" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "label:id",
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo/"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm {id}", h1())
	})

	t.Run("list & label", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo", funcHandler(func() string { return "i'm /foo" }))
		require.NoError(t, err)

		err = root.Insert("/foo/{id}", funcHandler(func() string { return "i'm /foo/{id}" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo",
			children: []*routePath{
				{
					path: "/",
					children: []*routePath{
						{
							path: "label:id",
						},
					},
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo", h1())

		ctx = NewContext(ctx)

		h2 := lookupFunc(root.Lookup(ctx, "/foo/abc"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/{id}", h2())
		require.Equal(t, "abc", GetURLParam(ctx, "id"))
	})

	t.Run("list with slash & label", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/", funcHandler(func() string { return "i'm /foo/" }))
		require.NoError(t, err)

		err = root.Insert("/foo/{id}", funcHandler(func() string { return "i'm /foo/{id}" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo/",
			children: []*routePath{
				{
					path: "label:id",
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo/"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo/", h1())

		ctx = NewContext(ctx)

		h2 := lookupFunc(root.Lookup(ctx, "/foo/abc"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/{id}", h2())
		require.Equal(t, "abc", GetURLParam(ctx, "id"))
	})

	t.Run("list with slash & label with slash", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/", funcHandler(func() string { return "i'm /foo/" }))
		require.NoError(t, err)

		err = root.Insert("/foo/{id}/", funcHandler(func() string { return "i'm /foo/{id}/" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo/",
			children: []*routePath{
				{
					path: "label:id:/",
					children: []*routePath{
						{
							path: "/",
						},
					},
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo/"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo/", h1())

		h2 := lookupFunc(root.Lookup(ctx, "/foo/abc"))
		require.Nil(t, h2)

		ctx = NewContext(ctx)

		h3 := lookupFunc(root.Lookup(ctx, "/foo/abc/"))
		require.NotNil(t, h3)
		require.Equal(t, "i'm /foo/{id}/", h3())
		require.Equal(t, "abc", GetURLParam(ctx, "id"))
	})

	t.Run("list & static with label", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/", funcHandler(func() string { return "i'm /foo/" }))
		require.NoError(t, err)

		err = root.Insert("/foo/bar", funcHandler(func() string { return "i'm /foo/bar" }))
		require.NoError(t, err)

		err = root.Insert("/foo/{id}", funcHandler(func() string { return "i'm /foo/{id}" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo/",
			children: []*routePath{
				{
					path: "bar",
				},
				{
					path: "label:id",
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo/"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo/", h1())

		h2 := lookupFunc(root.Lookup(ctx, "/foo/bar"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/bar", h2())

		ctx = NewContext(ctx)

		h3 := lookupFunc(root.Lookup(ctx, "/foo/abc"))
		require.NotNil(t, h3)
		require.Equal(t, "i'm /foo/{id}", h3())
		require.Equal(t, "abc", GetURLParam(ctx, "id"))
	})

	t.Run("list & static with label priority", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/", funcHandler(func() string { return "i'm /foo/" }))
		require.NoError(t, err)

		err = root.Insert("/foo/{id}", funcHandler(func() string { return "i'm /foo/{id}" }))
		require.NoError(t, err)

		err = root.Insert("/foo/bar", funcHandler(func() string { return "i'm /foo/bar" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo/",
			children: []*routePath{
				{
					path: "bar",
				},
				{
					path: "label:id",
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo/"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo/", h1())

		h2 := lookupFunc(root.Lookup(ctx, "/foo/bar"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/bar", h2())

		ctx = NewContext(ctx)

		h3 := lookupFunc(root.Lookup(ctx, "/foo/abc"))
		require.NotNil(t, h3)
		require.Equal(t, "i'm /foo/{id}", h3())
		require.Equal(t, "abc", GetURLParam(ctx, "id"))
	})

	t.Run("multi labels", func(t *testing.T) {
		root := &Node{}

		err := root.Insert("/foo/{id}/bar/{id}", funcHandler(func() string { return "i'm /foo/{id}/bar/{id}" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo/",
			children: []*routePath{
				{
					path: "label:id:/",
					children: []*routePath{
						{
							path: "/bar/",
							children: []*routePath{
								{
									path: "label:id",
								},
							},
						},
					},
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)
	})

	t.Run("label url param", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/{foo_id}/bar/{bar_id}", funcHandler(func() string { return "i'm /foo/{foo_id}/bar/{bar_id}" }))
		require.NoError(t, err)

		ctx = NewContext(ctx)

		h := lookupFunc(root.Lookup(ctx, "/foo/1/bar/2"))
		require.NotNil(t, h)
		require.Equal(t, "i'm /foo/{foo_id}/bar/{bar_id}", h())

		fooID := GetURLParam(ctx, "foo_id")
		barID := GetURLParam(ctx, "bar_id")
		require.Equal(t, "1", fooID)
		require.Equal(t, "2", barID)

	})

	t.Run("failed concatenated labels", func(t *testing.T) {
		root := &Node{}

		err := root.Insert("{id}{id2}", funcHandler(func() string { return "i'm invalid" }))
		require.Error(t, err)
	})
}

func TestRegexpRoute(t *testing.T) {
	t.Run("new regexp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/{id:[\\d]+}", funcHandler(func() string { return "i'm /foo/{id}" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo/",
			children: []*routePath{
				{
					path: "regexp:id:[\\d]+",
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		h1 := lookupFunc(root.Lookup(ctx, "/foo/"))
		require.Nil(t, h1)

		ctx = NewContext(ctx)

		h2 := lookupFunc(root.Lookup(ctx, "/foo/123"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/{id}", h2())
		require.Equal(t, "123", GetURLParam(ctx, "id"))

		h3 := lookupFunc(root.Lookup(ctx, "/foo/abc"))
		require.Nil(t, h3)

		h4 := lookupFunc(root.Lookup(ctx, "/foo/123abc"))
		require.Nil(t, h4)
	})

	t.Run("static, label & regexp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo", funcHandler(func() string { return "i'm /foo" }))
		require.NoError(t, err)

		err = root.Insert("/foo/{slug:[a-z]+}", funcHandler(func() string { return "i'm /foo/{slug}" }))
		require.NoError(t, err)

		err = root.Insert("/foo/{id}", funcHandler(func() string { return "i'm /foo/{id}" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo",
			children: []*routePath{
				{
					path: "/",
					children: []*routePath{
						{
							path: "regexp:slug:[a-z]+",
						},
						{
							path: "label:id",
						},
					},
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		ctx = NewContext(ctx)

		h1 := lookupFunc(root.Lookup(ctx, "/foo"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo", h1())

		h2 := lookupFunc(root.Lookup(ctx, "/foo/abc"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/{slug}", h2())

		slug := GetURLParam(ctx, "slug")
		require.Equal(t, "abc", slug)

		h3 := lookupFunc(root.Lookup(ctx, "/foo/123"))
		require.NotNil(t, h3)
		require.Equal(t, "i'm /foo/{id}", h3())

		id := GetURLParam(ctx, "id")
		require.Equal(t, "123", id)
	})

	t.Run("static, label & regexp - priority", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		root := &Node{}

		err := root.Insert("/foo/{id}", funcHandler(func() string { return "i'm /foo/{id}" }))
		require.NoError(t, err)

		err = root.Insert("/foo/{slug:[a-z]+}", funcHandler(func() string { return "i'm /foo/{slug}" }))
		require.NoError(t, err)

		err = root.Insert("/foo", funcHandler(func() string { return "i'm /foo" }))
		require.NoError(t, err)

		expected := &routePath{
			path: "/foo",
			children: []*routePath{
				{
					path: "/",
					children: []*routePath{
						{
							path: "regexp:slug:[a-z]+",
						},
						{
							path: "label:id",
						},
					},
				},
			},
		}

		actual := buildTree(root)
		require.EqualValues(t, expected, actual)

		ctx = NewContext(ctx)

		h1 := lookupFunc(root.Lookup(ctx, "/foo"))
		require.NotNil(t, h1)
		require.Equal(t, "i'm /foo", h1())

		h2 := lookupFunc(root.Lookup(ctx, "/foo/abc"))
		require.NotNil(t, h2)
		require.Equal(t, "i'm /foo/{slug}", h2())

		slug := GetURLParam(ctx, "slug")
		require.Equal(t, "abc", slug)

		h3 := lookupFunc(root.Lookup(ctx, "/foo/123"))
		require.NotNil(t, h3)
		require.Equal(t, "i'm /foo/{id}", h3())

		id := GetURLParam(ctx, "id")
		require.Equal(t, "123", id)
	})
}

func buildTree(n *Node) *routePath {
	if n == nil {
		return nil
	}
	path := &routePath{path: n.matcher.Data()}
	for _, nn := range n.children {
		path.children = append(path.children, buildTree(nn))
	}
	return path
}

type funcHandler func() string

func (f funcHandler) IsPartial() bool {
	return false
}

func lookupFunc(h any, str string) func() string {
	if h == nil {
		return nil
	}

	fn, ok := h.(funcHandler)
	if !ok {
		return nil
	}

	return fn
}
