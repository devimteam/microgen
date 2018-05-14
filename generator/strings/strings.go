package strings

import (
	"strings"
	"unicode"
)

// Fetch information from slice of comments (docs).
// Returns appendix of first comment which has tag as prefix.
func FetchMetaInfo(tag string, comments []string) string {
	for _, comment := range comments {
		if len(comment) > len(tag) && strings.HasPrefix(comment, tag) {
			return comment[len(tag)+1:]
		}
	}
	return ""
}

func ContainTag(strs []string, prefix string) bool {
	for _, comment := range strs {
		if strings.HasPrefix(comment, prefix) {
			return true
		}
	}
	return false
}

func LastWordFromName(name string) string {
	lastUpper := strings.LastIndexFunc(name, unicode.IsUpper)
	if lastUpper == -1 {
		lastUpper = 0
	}
	return strings.ToLower(name[lastUpper:])
}
