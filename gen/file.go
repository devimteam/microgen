package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	lg "microgen/logger"
)

var importsCache = make(map[string]string)

type File struct {
	usedImports map[string]string
	b           *bytes.Buffer
}

func NewFile() *File {
	return &File{b: &bytes.Buffer{}, usedImports: make(map[string]string)}
}

// Write
func (f *File) W(ss ...interface{}) *File {
	for i := range ss {
		switch s := ss[i].(type) {
		case []interface{}:
			f.W(s...)
		case Imp:
			f.writeImport(string(s), "")
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
		lg.Logger.Logf(lg.Debug, "arg %d: type %T\n", i, ss[i])
		switch s := ss[i].(type) {
		case []interface{}:
			f.W(s...)
		case imp:
			f.writeImport(s.pkg, s.decl)
		case Imp:
			f.writeImport(string(s), "")
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

func (f *File) writeImport(path string, decl string) {
	lg.Logger.Logln(lg.Debug, "write import", path)
	alias, ok := importsCache[path]
	if ok {
		lg.Logger.Logln(lg.Debug, "take import", path, "from cache, alias:", alias)
		f.b.WriteString(alias)
		f.usedImports[path] = alias
		return
	}
	defer func() {
		lg.Logger.Logln(lg.Debug, "save import", path, "to cache, alias:", alias)
		importsCache[path] = alias
	}()
	lg.Logger.Logln(lg.Debug, "try to find import", path)
	path, err := GetRelatedFilePath(path)
	if err != nil {
		lg.Logger.Logln(2, "resolve import alias", path, "get related package path", "error", err)
		alias = constructFullPackageAlias(path)
		return
	}
	lg.Logger.Logln(lg.Debug, "parse", path, "to find import", path)
	// 100% match way: parse directory and take one of the package names from.
	pkgs, err := parser.ParseDir(token.NewFileSet(), path, nonTestFilter, parser.PackageClauseOnly)
	if err != nil {
		lg.Logger.Logln(2, "resolve import alias", path, "parse package", path, "error", err)
		alias = constructFullPackageAlias(path)
		return
	}
	for k, pkg := range pkgs {
		lg.Logger.Logln(lg.Debug, path, "has package", k)
		// Name of type was not provided
		if decl == "" {
			alias = k
			break
		}
		if !ast.PackageExports(pkg) {
			continue
		}
		if ast.FilterPackage(pkg, func(name string) bool { return name == decl }) {
			// filter returns true if package has declaration
			// make it to be sure, that we choose right alias
			alias = k
			break
		}
	}
	f.b.WriteString(alias)
	f.usedImports[path] = alias
	return
}

// filters all files with tests
func nonTestFilter(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
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
	imports := makeSlice(f.usedImports)

	return f.b.Bytes()
}

func makeSlice(m map[string]string) []string {
	imports := make([]string, len(m))
	i := 0
	for k := range m {
		imports[i] = k
		i++
	}
	sort.Strings(imports)
	for i := range imports {
		imports[i] = m[imports[i]] + " " + imports[i]
	}
	return imports
}

func (f *File) Render(w io.Writer) error {
	_, err := w.Write(f.Bytes())
	return err
}

type (
	Imp string
	imp struct {
		pkg  string
		decl string
	}
)

func Qual(Package, Type string) imp {
	return imp{pkg: Package, decl: Type}
}

func Dot(ss ...interface{}) []interface{} {
	return ss
}
