package yahasplugin

import (
	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
	"os"

	"github.com/gorilla/mux"
)

type Plugin interface {
	Init(Provider, *logging.Logger, *os.File) error
	DeInit() error
}
type Provider interface {
	RequestRouter() (*mux.Router, error)
	Items() (*item.NamespaceMap, error)
	RequestPlugins() (Manager, error)
	BindingManager() (item.BindingManager, error)
}

type Manager interface {
	Load(name string) error
	Unload(name string) error
	Map() *map[string]*Plugin
}
