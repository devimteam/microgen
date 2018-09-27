package microgen

import (
	"sync"

	lg "github.com/devimteam/microgen/logger"
)

var (
	pluginsRegistry map[string]Plugin
	regLock         sync.Mutex
)

func RegisterPlugin(name string, plugin Plugin) {
	regLock.Lock()
	pluginsRegistry[name] = plugin
	lg.Logger.Logln(4, "register plugin", name)
	regLock.Unlock()
}
