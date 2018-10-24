package microgen

import (
	"reflect"
)

type Context struct {
	Interface           *Interface
	Source              string
	SourcePackageName   string
	SourcePackageImport string
	FileHeader          string
	AllowedMethods      map[string]bool
	Files               []File
}

type File struct {
	Content []byte
	Path    string
	// Unique name of file, that other plugins can easily find it
	Name string
}

type Interface struct {
	Name    string
	Value   reflect.Value
	Docs    []string
	Methods []Method
}

type Method struct {
	Docs    []string
	Name    string
	Args    []string
	Results []string
}
