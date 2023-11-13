package limi

import "context"

type StringMatcher struct {
	data string
}

func NewStringMatcher(str string) *StringMatcher {
	return &StringMatcher{data: str}
}

func (s *StringMatcher) Match(ctx context.Context, str string) (bool, string) {
	if s.data == str {
		return true, ""
	}

	slen := min(len(str), len(s.data))
	var matched []byte
	for i := 0; i < slen; i++ {
		if str[i] != s.data[i] {
			break
		}
		matched = append(matched, str[i])
	}

	if string(matched) != s.data {
		return false, str
	}

	//partial match
	trail1 := str[len(matched):]
	return true, trail1

}

func (s *StringMatcher) Parse(str string) (bool, string, string, string) {
	slen := min(len(str), len(s.data))
	var matched []byte
	for i := 0; i < slen; i++ {
		if str[i] != s.data[i] {
			break
		}
		matched = append(matched, str[i])
	}

	trail1 := str[len(matched):]
	trail2 := s.data[len(matched):]
	return str == s.data, string(matched), trail1, trail2
}

func (s *StringMatcher) Data() string {
	return s.data
}

func (s *StringMatcher) Type() MatcherType {
	return TypeString
}

func min(i1, i2 int) int {
	if i1 < i2 {
		return i1
	}
	return i2
}

func (s *StringMatcher) Label() string {
	return ""
}
