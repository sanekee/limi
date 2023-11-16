package limi

import (
	"context"
)

type LabelMatcher struct {
	data  string
	label string
	trail byte
}

func NewLabelMatcher(str string) *LabelMatcher {
	label := str[1 : len(str)-1]
	return &LabelMatcher{data: str, label: label}
}

func (s *LabelMatcher) Match(ctx context.Context, str string) (bool, string) {
	var matched []byte
	for _, b := range str {
		if s.trail != 0 &&
			s.trail == byte(b) {
			break
		}
		matched = append(matched, byte(b))
	}

	isMatched := len(matched) != 0
	trail := str[len(matched):]

	if len(matched) > 0 {
		SetURLParam(ctx, s.label, string(matched))
	}
	return isMatched, trail
}

func (s *LabelMatcher) Parse(p Parser) (bool, string, string, string) {
	str := p.Str
	if TypeLabel != p.Type {
		return false, "", str, s.data
	}

	if s.data != p.Str {
		return false, "", str, s.data
	}

	return true, str, "", ""
}

func (s *LabelMatcher) Data() string {
	ret := "label:" + s.label
	if s.trail != 0 {
		ret += ":" + string(s.trail)
	}
	return ret
}

func (s *LabelMatcher) Type() MatcherType {
	return TypeLabel
}

func (s *LabelMatcher) SetTrail(trail byte) {
	s.trail = trail
}

func (s *LabelMatcher) Label() string {
	return s.label
}
