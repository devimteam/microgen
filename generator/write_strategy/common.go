package write_strategy

import "io"

type Renderer interface {
	Render(io.Writer) error
}

type Strategy interface {
	Write(Renderer) error
}
