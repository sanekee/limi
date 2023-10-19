package limi

import (
	"context"
	"regexp"
	"strings"
)

type RegexpMatcher struct {
	label  string
	regexp *regexp.Regexp
	trail  byte
}

func NewRegexpMatcher(str string) *RegexpMatcher {
	strArr := strings.Split(str, ":")
	if len(strArr) != 2 {
		panic("invalid regexp format")
	}
	return &RegexpMatcher{label: strArr[0], regexp: regexp.MustCompile(strArr[1])}
}

func (s *RegexpMatcher) Match(ctx context.Context, str string) (bool, string, string) {
	var testStr []byte
	for _, b := range str {
		if s.trail != 0 &&
			s.trail == byte(b) {
			break
		}
		testStr = append(testStr, byte(b))
	}

	matched := s.regexp.Find(testStr)
	isMatched := len(matched) == len(str)
	trail := str[len(matched):]

	if len(matched) > 0 {
		SetURLParam(ctx, s.label, string(matched))
	}
	return isMatched, string(matched), trail
}

func (s *RegexpMatcher) Parse(str string) (bool, string, string, string) {
	if len(str) < 3 ||
		str[0] != '{' ||
		str[len(str)-1] != '}' {
		return false, str, str, ""
	}

	label := str[1 : len(str)-2]
	if label == (s.label + ":" + s.regexp.String()) {
		return true, "", "", ""
	}

	return false, str, str, ""
}

func (s *RegexpMatcher) Data() string {
	ret := "regexp:" + s.label + ":" + s.regexp.String()
	if s.trail != 0 {
		ret += ":" + string(s.trail)
	}
	return ret
}

func (s *RegexpMatcher) Type() MatcherType {
	return TypeRegexp
}

func (s *RegexpMatcher) SetTrail(trail byte) {
	s.trail = trail
}
