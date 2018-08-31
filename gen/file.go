package gen

import (
	"bytes"
	"fmt"
	"io"
)

type File struct {
	imports map[string]string
	b       *bytes.Buffer
}

func NewFile() *File {
	return &File{b: &bytes.Buffer{}}
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

func (f *File) Render(w io.Writer) error {
	_, err := w.Write(f.Bytes())
	return err
}
