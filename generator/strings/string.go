package strings

import (
	"strings"
	"unicode"
)

func ToUpperFirst(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func ToSomeCaseWithSep(sep rune, runeConv func(rune) rune) func(string) string {
	return func(s string) string {
		in := []rune(s)
		n := len(in)
		var runes []rune
		for i, r := range in {
			if isExtendedSpace(r) {
				runes = append(runes, sep)
				continue
			}
			if unicode.IsUpper(r) {
				if i > 0 && sep != runes[i-1] && ((i+1 < n && unicode.IsLower(in[i+1])) || unicode.IsLower(in[i-1])) {
					runes = append(runes, sep)
				}
				r = runeConv(r)
			}
			runes = append(runes, r)
		}
		return string(runes)
	}
}

func isExtendedSpace(r rune) bool {
	return unicode.IsSpace(r) || r == '_' || r == '-' || r == '.'
}

var (
	ToSnakeCase    = ToSomeCaseWithSep('_', unicode.ToLower)
	ToURLSnakeCase = ToSomeCaseWithSep('-', unicode.ToLower)
)

func ToLowerFirst(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(string(s[0])) + s[1:]
}

func IsInStringSlice(what string, where []string) bool {
	for _, item := range where {
		if item == what {
			return true
		}
	}
	return false
}

func FetchTags(strs []string, prefix string) (tags []string) {
	for _, comment := range strs {
		if strings.HasPrefix(comment, prefix) {
			tags = append(tags, strings.Split(strings.Replace(comment[len(prefix):], " ", "", -1), ",")...)
		}
	}
	return
}

func HasTag(strs []string, prefix string) bool {
	return ContainTag(strs, prefix)
}

func ToLower(str string) string {
	if len(str) > 0 && unicode.IsLower(rune(str[0])) {
		return str
	}
	for i := range str {
		if unicode.IsLower(rune(str[i])) {
			// Case, when only first char is upper.
			if i == 1 {
				return strings.ToLower(str[:1]) + str[1:]
			}
			return strings.ToLower(str[:i-1]) + str[i-1:]
		}
	}
	return strings.ToLower(str)
}

// Return last upper char in string or first char if no upper characters founded.
func LastUpperOrFirst(str string) string {
	for i := len(str) - 1; i >= 0; i-- {
		if unicode.IsUpper(rune(str[i])) {
			return string(str[i])
		}
	}
	return string(str[0])
}
