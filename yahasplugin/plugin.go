package yahasplugin

import (
	"os"

	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
)

type Plugin interface {
	Init(Provider, *logging.Logger, *os.File) error
	DeInit() error
}

type Service interface {
	Name() string
	ProvidedBy() string
}

type Provider interface {
	//RequestRouter() (*mux.Router, error)
	Items() (*item.NamespaceMap, error)
	RequestPlugins() (Manager, error)
	BindingManager() (item.BindingManager, error)
	Register(string, Service) error
	Unregister(string) error
	Get(string) (Service, error)
}

type Manager interface {
	Load(name string) error
	Unload(name string) error
	Map() *map[string]*Plugin
}
