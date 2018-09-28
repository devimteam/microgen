package microgen

type config struct {
	Plugins   []string       `toml:"plugins"`
	Interface string         `toml:"interface"`
	Generate  []PluginConfig `toml:"generate"`
}

type PluginConfig struct {
	Name string   `toml:"name"`
	Args []string `toml:"args"`
	//Path string   `toml:"path"`
}
