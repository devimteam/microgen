package strings

import (
	"strings"
	"unicode"
)

func LastWordFromName(name string) string {
	lastUpper := strings.LastIndexFunc(name, unicode.IsUpper)
	if lastUpper == -1 {
		lastUpper = 0
	}
	return strings.ToLower(name[lastUpper:])
}
