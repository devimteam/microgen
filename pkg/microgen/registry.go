package microgen

import (
	"reflect"
	"sync"

	"github.com/devimteam/microgen/logger"
	lg "github.com/devimteam/microgen/logger"
)

var (
	// holds all plugins by their names.
	pluginsRepository    = make(map[string]Plugin)
	interfacesRepository = make([]api, 0, 1)
	interfacesComments   = make([][]string, 0, 1)
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

func RegisterInterface(name string, value interface{}) {
	if name == "" {
		return
	}
	regLock.Lock()
	interfacesRepository = append(interfacesRepository, api{name: name, value: reflect.ValueOf(value)})
	interfacesComments = append(interfacesComments, nil)
	lg.Logger.Logln(logger.Detail, "register interface", name)
	regLock.Unlock()
}

/*
func AddComments(name string, comments []string) {
	if name == "" {
		return
	}
	regLock.Lock()
	for i := range interfacesRepository {
		if interfacesRepository[i] == name {
			interfacesComments[i] = append(interfacesComments[i], comments...)
			break
		}
	}
	regLock.Unlock()
}
*/
type api struct {
	name     string
	value    reflect.Value
	comments []string
}
