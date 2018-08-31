package generator

import "strings"

func splitPaths(path string) []string {
	return strings.Split(path, ";")
}
