package microgen

import (
	toml "github.com/pelletier/go-toml"
)

type PluginConfig struct {
	Plugin string    `toml:"plugin"`
	Params toml.Tree `toml:"params"`
}
