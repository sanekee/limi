package limi

import (
	"fmt"
	"regexp"
	"strings"
)

type MatcherType int

const (
	TypeUnknown MatcherType = iota
	TypeString
	TypeRegexp
	TypeLabel
)

type Matcher interface {
	Parse(Parser) (bool, string, string, string)
	Match(string) (bool, string, string)
	Type() MatcherType
	Data() string
	Label() string
}

type Parser struct {
	Str  string
	Type MatcherType
}

func SplitParsers(str string) ([]Parser, error) {
	var parsers []Parser
	parse := str
	next := ""
	for {
		if parse == "" {
			break
		}

		if parse != "" && parse[0] == '{' {
			idx := strings.IndexByte(parse, '}')
			if idx < 0 {
				return nil, fmt.Errorf("missing closing } in label %w", ErrInvalidInput)
			}

			parse, next = parse[:idx+1], parse[idx+1:]
			if idx := strings.Index(parse, ":"); idx > 0 {
				expr := parse[idx:]
				_, err := regexp.Compile(expr)
				if err != nil {
					return nil, fmt.Errorf("invalid regular expression %s %w", expr, ErrInvalidInput)
				}
				parsers = append(parsers, Parser{Str: parse, Type: TypeRegexp})
			} else {
				parsers = append(parsers, Parser{Str: parse, Type: TypeLabel})
			}
			parse = next
		}

		idx := strings.IndexByte(parse, '{')
		if idx < 0 {
			break
		}

		if idx > 0 {
			parse, next = parse[0:idx], parse[idx:]
			parsers = append(parsers, Parser{Str: parse, Type: TypeString})
		}
		parse = next
	}

	if parse != "" {
		parsers = append(parsers, Parser{Str: parse, Type: TypeString})
	}

	return parsers, nil
}

func NewMatcher(p Parser) Matcher {
	switch p.Type {
	case TypeLabel:
		return NewLabelMatcher(p.Str)
	case TypeRegexp:
		return NewRegexpMatcher(p.Str)
	}
	return NewStringMatcher(p.Str)
}
