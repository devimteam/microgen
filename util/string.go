package util

import (
	"strings"
	"unicode"
)

func ToUpperFirst(s string) string {
	return strings.ToUpper(string(s[0])) + s[1:]
}

func ToSnakeCase(s string) string {
	for i := 0; i < len(s); i++ {
		if unicode.IsUpper(rune(s[i])) {
			s = strings.Join([]string{s[:i], ToLowerFirst(s[i:])}, "_")
		}
	}
	return s
}

func ToLowerFirst(s string) string {
	return strings.ToLower(string(s[0])) + s[1:]
}

func FirstLowerChar(s string) string {
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