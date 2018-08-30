package gen

import (
	"bytes"
)

type File struct {
	imports map[string]string
	b       bytes.Buffer
}

func NewFile() *File {
	return &File{}
}

// Write
func (f *File) W(ss ...interface{}) *File {
	for i := range ss {
		switch s := ss[i].(type) {
		case string:
			f.b.WriteString(s)
		case func() string:
			f.b.WriteString(s())
		case []byte:
			f.b.Write(s)
		}
	}
	return f
}

/*
// Import
func (f *File) I() *File {

}
*/
func (f *File) Read(b []byte) (int, error) {
	return f.b.Read(b)
}

func (f *File) String() string {
	return string(f.Bytes())
}

func (f *File) Bytes() []byte {
	return f.b.Bytes()
}
