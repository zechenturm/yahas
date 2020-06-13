package yahasplugin

import (
	"github.com/gorilla/mux"
	"github.com/zechenturm/yahas/item"
)

type yPlugin struct {
	Name           string
	router         *mux.Router
	bindingManager item.BindingManager
	plugin         Plugin
	items          *item.NamespaceMap
}

func (yp *yPlugin) Items() (*item.NamespaceMap, error) {
	return yp.items, nil
}

func (yp *yPlugin) RequestPlugins() (Manager, error) {
	return nil, nil
}

func (yp *yPlugin) BindingManager() (item.BindingManager, error) {
	return yp.bindingManager, nil
}
