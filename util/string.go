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
			s = strings.Join([]string{s[:i], strings.ToLower(string(s[i])) + s[i+1:]}, "_")
		}
	}
	return s
}