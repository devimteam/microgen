package internal

import "strings"

func splitPaths(path string) []string {
	return strings.Split(path, ";")
}

func formatPackage(s string) string {
	return strings.Replace(s, "\\", "/", -1)
}
