package limi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
)

type Handle interface {
	IsPartial() bool
	Merge(Handle) bool
	ServeHTTP(http.ResponseWriter, *http.Request)
	IsMethodAllowed(string) bool
}

type Node struct {
	children nodes
	handle   Handle
	matcher  Matcher
}

func (n *Node) Insert(str string, h Handle) error {
	if str == "" {
		return fmt.Errorf("node string cannot be empty %w", ErrInvalidInput)
	}

	parsers, err := SplitParsers(str)
	if err != nil {
		return fmt.Errorf("failed to split string, %w", err)
	}

	node := n
	for _, p := range parsers {
		if node.matcher != nil &&
			node.matcher.Type() == TypeLabel {
			if p.Type == TypeLabel {
				return errors.New("invalid label matcher without separator")
			}
			if p.Type == TypeString {
				labelMatcher, ok := node.matcher.(*LabelMatcher)
				if !ok {
					return errors.New("error casting matcher")
				}
				labelMatcher.SetTrail(p.Str[0])
			}
		}
		lastNode, _, err := insert(node, p)
		if err != nil {
			return errors.New("failed to insert handle")
		}
		node = lastNode
	}

	if node == nil {
		return errors.New("unknown node")
	}

	if node.handle != nil {
		if node.handle.Merge(h) {
			return nil
		} else {
			return fmt.Errorf("handle already existed %w", ErrHandleExists)
		}
	}
	node.handle = h

	return nil
}

func insert(n *Node, p Parser) (*Node, string, error) {
	if n.matcher == nil {
		n.matcher = NewMatcher(p)
		return n, "", nil
	}

	str := p.Str
	if n.matcher.Type() != p.Type {
		newNode := &Node{matcher: NewMatcher(p)}
		n.children = append(n.children, newNode)
		sort.Sort(n.children)
		return newNode, "", nil
	}

	isMatched, matched, trailStr, trailNode := n.matcher.Parse(str)
	if isMatched {
		return n, "", nil
	}

	if len(matched) == 0 {
		return nil, str, nil
	}

	// reparent current string's remainder
	if trailNode != "" {
		children, handle := n.children, n.handle

		n.matcher = NewStringMatcher(matched)
		n.children = append([]*Node{}, &Node{children: children, handle: handle, matcher: NewStringMatcher(trailNode)})
		n.handle = nil
	}

	// search new string's remainder
	str = trailStr
	if trailStr != "" {
		for _, nn := range n.children {
			if nn.matcher.Type() != p.Type {
				continue
			}
			nnn, str1, err := insert(nn, Parser{Str: str, Type: p.Type})
			if err != nil {
				return nil, str, fmt.Errorf("failed to insert node %w", err)
			}
			if str1 == "" {
				return nnn, "", nil
			}
			str = str1
		}
		newNode := &Node{matcher: NewStringMatcher(str)}
		n.children = append(n.children, newNode)
		sort.Sort(n.children)

		return newNode, "", nil
	}

	return n, str, nil
}

func (n *Node) Walk(fn func(level int, str string, h any)) {
	var level int

	n.walk(level, fn)
}

func (n *Node) walk(level int, fn func(level int, str string, h any)) {
	fn(level, n.matcher.Data(), n.handle)
	level++
	for _, nn := range n.children {
		nn.walk(level, fn)
	}
}

func (n *Node) Lookup(ctx context.Context, str string) (Handle, string) {
	return lookup(ctx, n, str)

}

func lookup(ctx context.Context, n *Node, str string) (Handle, string) {
	if str == "" {
		return nil, ""
	}

	if n.matcher == nil {
		return nil, str
	}

	isMatched, matched, trail := n.matcher.Match(ctx, str)
	if isMatched && n.handle != nil {
		return n.handle, trail
	}

	if len(matched) == 0 {
		return nil, trail
	}

	for _, nn := range n.children {
		h, trail := lookup(ctx, nn, trail)
		if h != nil {
			return h, trail
		}
	}

	if n.handle != nil &&
		n.handle.IsPartial() &&
		matched == n.matcher.Data() {
		return n.handle, trail
	}
	return nil, ""

}

type nodes []*Node

func (n nodes) Less(i, j int) bool {
	if n[i].matcher == nil {
		return true
	}
	if n[j].matcher == nil {
		return false
	}

	return n[i].matcher.Type() < n[j].matcher.Type()
}

func (n nodes) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n nodes) Len() int      { return len(n) }

type HTTPHandler http.HandlerFunc

func (h HTTPHandler) IsPartial() bool {
	return true
}

func (h HTTPHandler) Merge(Handle) bool {
	return false
}

func (h HTTPHandler) IsMethodAllowed(string) bool {
	return true
}

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h(w, req)
}
