package template

import (
	"bytes"
	"fmt"
	"io"
)

type BufferAdapter struct {
	b bytes.Buffer
}

func (b BufferAdapter) Render(w io.Writer) error {
	_, err := w.Write(b.b.Bytes())
	return err
}

func (b *BufferAdapter) Raw(data []byte) {
	b.b.Write(data)
}

func (b *BufferAdapter) Printf(format string, a ...interface{}) {
	fmt.Fprintf(&b.b, format, a...)
}

func (b *BufferAdapter) Println(a ...interface{}) {
	fmt.Fprintln(&b.b, a...)
}

func (b *BufferAdapter) Ln(a ...interface{}) {
	b.Println(a...)
}

func (b *BufferAdapter) Lnf(format string, a ...interface{}) {
	b.Printf(format+"\n", a...)
}

func (b *BufferAdapter) Hold() *DelayBuffer {
	return &DelayBuffer{t: b}
}

type DelayBuffer struct {
	t *BufferAdapter
	BufferAdapter
}

func (b *DelayBuffer) Release() {
	b.t.Raw(b.b.Bytes())
}
