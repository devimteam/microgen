package write_method

import "io"

type Renderer interface {
	Render(io.Writer) error
}

type Method interface {
	Write(Renderer) error
}
