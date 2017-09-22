package util

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

func ToSnakeCase(s string) string {
	for i := range s {
		if unicode.IsUpper(rune(s[i])) {
			if i != 0 {
				s = strings.Join([]string{s[:i], ToLowerFirst(s[i:])}, "_")
			} else {
				s = ToLowerFirst(s)
			}
		}
	}
	return s
}

func ToLowerFirst(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(string(s[0])) + s[1:]
}

func FirstLowerChar(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(string(s[0]))
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
