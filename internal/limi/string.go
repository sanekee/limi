package limi

type StringMatcher struct {
	data string
}

func NewStringMatcher(str string) *StringMatcher {
	return &StringMatcher{data: str}
}

func (s *StringMatcher) Match(str string) (bool, string, string) {
	if s.data == str {
		return true, str, ""
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
		return false, "", str
	}

	//partial match
	trail1 := str[len(matched):]
	return true, string(matched), trail1

}

func (s *StringMatcher) Parse(p Parser) (bool, string, string, string) {
	if TypeString != p.Type {
		return false, "", p.Str, s.data
	}

	str := p.Str
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
