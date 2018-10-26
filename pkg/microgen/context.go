package microgen

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
	PackageName string
	Name        string
	Value       reflect.Value
	Docs        []string
	Methods     []Method
}

func (iface Interface) String() string {
	b := strings.Builder{}
	b.WriteString("microgen.Interface{\n")
	fmt.Fprintf(&b, "Name:%s,\n", strconv.Quote(iface.Name))
	fmt.Fprintf(&b, "Docs:[]string{%s},\n", strings.Join(stringpipe(strconv.Quote)(iface.Docs), ","))
	fmt.Fprintf(&b, "Methods:[]microgen.Method{\n")
	for i := range iface.Methods {
		fmt.Fprintf(&b, "%s,\n", iface.Methods[i].String())
	}
	b.WriteString("},")
	b.WriteByte('}')
	return b.String()
}

type Method struct {
	Docs    []string
	Name    string
	Args    []string
	Results []string
}

func (m Method) String() string {
	b := strings.Builder{}
	b.WriteString("microgen.Method{\n")
	fmt.Fprintf(&b, "Name:%s,\n", strconv.Quote(m.Name))
	fmt.Fprintf(&b, "Docs:[]string{%s},\n", strings.Join(stringpipe(strconv.Quote)(m.Docs), ","))
	fmt.Fprintf(&b, "Args:[]string{%s},\n", strings.Join(stringpipe(strconv.Quote)(m.Args), ","))
	fmt.Fprintf(&b, "Results:[]string{%s},\n", strings.Join(stringpipe(strconv.Quote)(m.Results), ","))
	b.WriteByte('}')
	return b.String()
}

func stringpipe(ff ...func(string) string) func([]string) []string {
	return func(ww []string) []string {
		ss := make([]string, len(ww))
		copy(ss, ww)
		for i := range ff {
			for j := range ss {
				ss[j] = ff[i](ss[j])
			}
		}
		return ss
	}
}
