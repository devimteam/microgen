package microgen

import (
	"github.com/vetcher/go-astra/types"
)

type Context struct {
	Interface           *types.Interface
	Source              string
	SourcePackageName   string
	SourcePackageImport string
	FileHeader          string
	Files               []File
}

type File struct {
	Content []byte
	Path    string
	// Unique name of file, that other plugins can easily find it
	Name string
}
