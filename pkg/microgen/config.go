package microgen

type config struct {
	Plugins []pluginConfig `toml:"plugin"`
}

type pluginConfig struct {
	Name string `toml:"name"`
}
