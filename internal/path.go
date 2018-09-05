package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	lg "github.com/devimteam/microgen/logger"
)

func ResolvePackagePath(outPath string) (string, error) {
	lg.Logger.Logln(3, "Try to resolve path for", outPath, "package...")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is empty")
	}
	lg.Logger.Logln(4, "\tGOPATH:", gopath)

	absOutPath, err := filepath.Abs(outPath)
	if err != nil {
		return "", err
	}
	lg.Logger.Logln(4, "\tResolving path:", absOutPath)

	lg.Logger.Logln(5, "\tSearch in paths:")
	for _, path := range splitPaths(gopath) {
		gopathSrc := filepath.Join(path, "src")
		lg.Logger.Logln(5, "\t\t", gopathSrc)
		if strings.HasPrefix(absOutPath, gopathSrc) {
			return formatPackage(absOutPath[len(gopathSrc)+1:]), nil
		}
	}
	return "", fmt.Errorf("path(%s) not in GOPATH(%s)", absOutPath, gopath)
}
