package microgen

import "github.com/vetcher/go-astra/types"

type Context struct {
	Interface           *types.Interface
	Source              string
	SourcePackageImport string
	Files               []File
}

type File struct {
	Content []byte
	Path    string
	// Unique name of file, that other plugins can easily find it
	Name string
}
