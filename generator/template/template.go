package template

import (
	"context"

	"github.com/devimteam/microgen/generator/write_strategy"
)

type Template interface {
	// Do all preparing actions, e.g. scan file.
	// Should be called first.
	Prepare(ctx context.Context) error
	// Default relative path for template (=file)
	DefaultPath() string
	// Template chooses generation strategy, e.g. appends to file or create new.
	ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error)
	// Main render function, where template produce code.
	Render(ctx context.Context) write_strategy.Renderer
}

// Template for tags, that not produce any files.
type EmptyTemplate struct{}

func (EmptyTemplate) Prepare(context.Context) error { return nil }
func (EmptyTemplate) DefaultPath() string           { return "" }
func (EmptyTemplate) ChooseStrategy(context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewNopStrategy("", ""), nil
}
func (EmptyTemplate) Render(context.Context) write_strategy.Renderer { return nil }
