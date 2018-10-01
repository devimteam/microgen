package microgen

import (
	"sync"

	"github.com/devimteam/microgen/logger"
	lg "github.com/devimteam/microgen/logger"
)

var (
	// holds all plugins by their names.
	pluginsRepository = make(map[string]Plugin)
	// makes RegisterPlugin calls concurrency safe (just for fun and because can).
	regLock sync.Mutex
)

// RegisterPlugin registers plugin to global repository of plugins.
// When generation process begins, microgen takes plugins from this repository and executes them.
// Plugin with empty name will be ignored and not added to repository.
// Already registered plugins can be overwritten.
func RegisterPlugin(name string, plugin Plugin) {
	if name == "" {
		return
	}
	regLock.Lock()
	pluginsRepository[name] = plugin
	lg.Logger.Logln(logger.Detail, "register plugin", name)
	regLock.Unlock()
}
