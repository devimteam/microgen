package util

import (
	"strings"
)

func ToUpperFirst(s string) string {
	return strings.ToUpper(string(s[0])) + s[1:]
}