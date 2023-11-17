package limi

import (
	"regexp"
	"strings"
)

type RegexpMatcher struct {
	data   string
	label  string
	regexp *regexp.Regexp
	trail  byte
}

func NewRegexpMatcher(str string) *RegexpMatcher {
	labelStr := str[1 : len(str)-1]
	strArr := strings.Split(labelStr, ":")
	if len(strArr) != 2 {
		panic("invalid regexp format")
	}
	return &RegexpMatcher{data: str, label: strArr[0], regexp: regexp.MustCompile(strArr[1])}
}

func (s *RegexpMatcher) Match(str string) (bool, string, string) {
	var testStr []byte

	if s.trail == 0 {
		testStr = []byte(str)
	} else {
		for _, b := range str {
			if s.trail == byte(b) {
				break
			}
			testStr = append(testStr, byte(b))
		}
	}

	matched := s.regexp.Find(testStr)
	isMatched := len(matched) != 0
	trail := str[len(matched):]

	return isMatched, string(matched), trail
}

func (s *RegexpMatcher) Parse(p Parser) (bool, string, string, string) {
	if TypeRegexp != p.Type {
		return false, "", p.Str, s.data
	}

	if p.Str != s.data {
		return false, "", p.Str, s.data
	}

	return true, p.Str, "", ""
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

func (s *RegexpMatcher) Label() string {
	return s.label
}
