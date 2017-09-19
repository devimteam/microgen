package template

import "github.com/devimteam/microgen/generator/write_strategy"

type Template interface {
	DefaultPath() string
	ChooseStrategy() (write_strategy.Strategy, error)
	Render() write_strategy.Renderer
}
