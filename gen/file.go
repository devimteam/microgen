package gen

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/devimteam/microgen/internal"
	lg "github.com/devimteam/microgen/logger"
)

var importsCache map[string]string

type File struct {
	usedImports map[string]bool
	b           *bytes.Buffer
}

func NewFile() *File {
	return &File{b: &bytes.Buffer{}, usedImports: make(map[string]bool)}
}

// Write
func (f *File) W(ss ...interface{}) *File {
	for i := range ss {
		switch s := ss[i].(type) {
		case []interface{}:
			f.W(s...)
		case Imp:
			f.writeImport(string(s))
		case string:
			f.b.WriteString(s)
		case int:
			f.b.WriteString(strconv.Itoa(s))
		case func() string:
			f.b.WriteString(s())
		case []byte:
			f.b.Write(s)
		}
	}
	return f
}

func (f *File) Wln(ss ...interface{}) *File {
	for i := range ss {
		switch s := ss[i].(type) {
		case []interface{}:
			f.W(s...)
		case Imp:
			f.writeImport(string(s))
		case string:
			f.b.WriteString(s)
		case func() string:
			f.b.WriteString(s())
		case []byte:
			f.b.Write(s)
		case int:
			f.b.WriteString(strconv.Itoa(s))
		case byte:
			f.b.WriteByte(s)
		}
	}
	f.b.WriteByte('\n')
	return f
}

// Write formatted
func (f *File) Wf(format string, ss ...interface{}) *File {
	prepared := make([]interface{}, len(ss))
	for i := range ss {
		switch s := ss[i].(type) {
		case func() string:
			prepared[i] = s()
		default:
			prepared[i] = s
		}
	}
	return f.wf(format, prepared...)
}

func (f *File) wf(format string, a ...interface{}) *File {
	fmt.Fprintf(f.b, format, a...)
	return f
}

func (f *File) writeImport(i string) {
	alias, ok := importsCache[i]
	if ok {
		f.b.WriteString(alias)
		f.usedImports[i] = true
		return
	}
	defer func() {
		importsCache[i] = alias
	}()
	path, err := internal.GetRelatedFilePath(i)
	if err != nil {
		lg.Logger.Logln(2, "resolve import alias", i, "get related package path", "error", err)
		alias = constructFullPackageAlias(i)
		return
	}
	// 100% match way: parse directory and take one of the package names from.
	pkgs, err := parser.ParseDir(token.NewFileSet(), path, nil, parser.PackageClauseOnly)
	if err != nil {
		lg.Logger.Logln(2, "resolve import alias", i, "parse package", path, "error", err)
		alias = constructFullPackageAlias(i)
		return
	}
	for k := range pkgs {
		if k != "" {
			alias = k
			break
		}
	}
	return
}

func (f *File) writeError(err error) {
	f.b.WriteString("__\n" + err.Error() + "\n__")
}

func constructFullPackageAlias(pkg string) string {
	pkg = path.Clean(pkg)
	dirs := strings.Split(pkg, string(filepath.Separator))
	return strings.Join(dirs, "_")
}

// Register alias for import
func (f *File) I(alias, path string) *File {
	importsCache[path] = alias
	return f
}

func (f *File) Read(b []byte) (int, error) {
	return f.b.Read(b)
}

func (f *File) String() string {
	return string(f.Bytes())
}

func (f *File) Bytes() []byte {
	return f.b.Bytes()
}

func (f *File) Render(w io.Writer) error {
	_, err := w.Write(f.Bytes())
	return err
}

type (
	Imp string
)

func Dot(ss ...interface{}) []interface{} {
	return ss
}
