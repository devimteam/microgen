package template

import "github.com/devimteam/microgen/generator/write_strategy"

type Template interface {
	// Do all preparing actions, e.g. scan file.
	// Should be called first.
	Prepare() error
	// Default relative path for template (=file)
	DefaultPath() string
	// Template chooses generation strategy, e.g. appends to file or create new.
	ChooseStrategy() (write_strategy.Strategy, error)
	// Main render function, where template produce code.
	Render() write_strategy.Renderer
}

type EmptyTemplate struct {
}

func (*EmptyTemplate) Prepare() error      { return nil }
func (*EmptyTemplate) DefaultPath() string { return "" }
func (*EmptyTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	return write_strategy.NewNopStrategy("", ""), nil
}
func (*EmptyTemplate) Render() write_strategy.Renderer { return nil }
