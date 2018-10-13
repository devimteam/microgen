package microgen

import (
	"encoding/json"
)

//type config struct {
//	Plugins   []string       `json:"plugins"`
//	Interface string         `json:"interface"`
//	Generate  []PluginConfig `json:"generate"`
//}

type config struct {
	Plugins    []string
	Interfaces map[string][]pluginConfig
}

type pluginConfig struct {
	Name string
	Args []byte `yaml:",inline"`
}

type PluginConfig struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}
