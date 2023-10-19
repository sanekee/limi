package limi

import (
	"context"
)

type LabelMatcher struct {
	data  string
	trail byte
}

func NewLabelMatcher(str string) *LabelMatcher {
	return &LabelMatcher{data: str}
}

func (s *LabelMatcher) Match(ctx context.Context, str string) (bool, string, string) {
	var matched []byte
	for _, b := range str {
		if s.trail != 0 &&
			s.trail == byte(b) {
			break
		}
		matched = append(matched, byte(b))
	}

	isMatched := len(matched) == len(str)
	trail := str[len(matched):]

	if len(matched) > 0 {
		SetURLParam(ctx, s.data, string(matched))
	}
	return isMatched, string(matched), trail
}

func (s *LabelMatcher) Parse(str string) (bool, string, string, string) {
	if len(str) < 3 ||
		str[0] != '{' ||
		str[len(str)-1] != '}' {
		return false, str, str, ""
	}

	label := str[1 : len(str)-2]
	if label == s.data {
		return true, "", "", ""
	}

	return false, str, str, ""
}

func (s *LabelMatcher) Data() string {
	ret := "label:" + s.data
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
