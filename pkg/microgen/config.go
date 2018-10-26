package microgen

import "github.com/pelletier/go-toml"

type config struct {
	Plugins  []string
	Generate []PluginConfig
}

type PluginConfig struct {
	Plugin string    `toml:"plugin"`
	Params toml.Tree `toml:"params"`
}
