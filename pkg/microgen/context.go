package microgen

import "github.com/vetcher/go-astra/types"

type Context struct {
	Interface           *types.Interface
	Source              string
	SourcePackageImport string
	Dst                 string
	DstPackageImport    string
	Files               []File
}

type File struct {
	Content []byte
	Path    string

	sources []string
}
