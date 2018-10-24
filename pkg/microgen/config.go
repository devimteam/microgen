package microgen

type config struct {
	Plugins  []string
	Generate []PluginConfig
}

type PluginConfig struct {
	Plugin string `yaml:"plugin"`
	Args   []byte `yaml:",inline"`
}
