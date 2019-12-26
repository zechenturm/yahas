package main

import (
	"errors"
	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
	"github.com/zechenturm/yahas/yahasplugin"
	"io/ioutil"
	"os"
	"path"
	"plugin"

	"github.com/gorilla/mux"
)

var plugins map[string]*yahasplugin.Plugin
var yPlugins map[string]*yPlugin
var pManager pluginManager

type pluginManager struct {
}

func (pm *pluginManager) Load(name string) error {
	if _, ok := plugins[name]; ok {
		return errors.New("Plugin alreeady loaded")
	}
	return loadPlugin(name + ".so")
}

func (pm *pluginManager) Unload(name string) error {
	plug, ok := yPlugins[name]
	if !ok {
		return errors.New("Plugin not loaded")
	}
	coreLogger.DebugLn("unlading " + name)
	*(plug.router) = mux.Router{}
	mainRouter.PathPrefix("/" + name).Subrouter()
	pl := plugins[name]
	delete(yPlugins, name)
	delete(plugins, name)
	return (*pl).DeInit()
}

func (pm *pluginManager) Map() *map[string]*yahasplugin.Plugin {
	return &plugins
}

func checkPermission(plugin string, name string, defaultAllow bool) bool {
	perm := false
	permMap, ok := coreconf.Permissions[plugin]
	if ok {
		perm, ok = permMap[name]
	}
	if defaultAllow && !ok {
		return true
	}
	return perm && ok
}

type yPlugin struct {
	Name   string
	router *mux.Router
}

func (yp *yPlugin) RequestRouter() (*mux.Router, error) {
	if !checkPermission(yp.Name, "router", true) {
		return nil, errors.New("Permission denied")
	}
	if yp.router == nil {
		yp.router = mainRouter.PathPrefix("/" + yp.Name).Subrouter()
	}
	return yp.router, nil
}

func (yp *yPlugin) Items() (*item.NamespaceMap, error) {
	if !checkPermission(yp.Name, "items", true) {
		return nil, errors.New("Permission denied")
	}
	return &Items, nil
}

func (yp *yPlugin) RequestPlugins() (yahasplugin.Manager, error) {
	if !checkPermission(yp.Name, "plugins", false) {
		return nil, errors.New("Permission denied")
	}
	return &pManager, nil
}

func (yp *yPlugin) BindingManager() (item.BindingManager, error) {
	if !checkPermission(yp.Name, "bindings", false) {
		return nil, errors.New("Permission denied")
	}
	return loader.BindingManager, nil
}

func loadPlugins(ignore []string) error {
	plugins = make(map[string]*yahasplugin.Plugin)
	yPlugins = make(map[string]*yPlugin)
	files, err := ioutil.ReadDir("./plugins")
	if err != nil {
		return err
	}

	coreLogger.DebugLn("plugin ignore list:", ignore)

	shouldIgnore := func(name string) bool {
		for _, elem := range ignore {
			if name == elem {
				return true
			}
		}
		return false
	}

	for _, file := range files {
		baseName := file.Name()[:len(file.Name())-len(".so")]
		if path.Ext(file.Name()) == ".so" && !shouldIgnore(baseName) {
			err := loadPlugin(file.Name())
			if err != nil {
				coreLogger.ErrorLn("Error loading plugin", file.Name(), ":", err)
			}
		}
	}
	coreLogger.DebugLn("plugins:", plugins)
	coreLogger.DebugLn("yplugins:", yPlugins)
	return nil
}

func loadPlugin(name string) error {
	coreLogger.InfoLn("loading plugin", name)
	var ypi yahasplugin.Plugin
	p, err := plugin.Open("./plugins/" + name)
	if err != nil {
		return err
	}

	sb, err := p.Lookup("Plugin")
	if err != nil {
		return err
	}
	ypi, ok := sb.(yahasplugin.Plugin)
	if !ok {
		return errors.New("loaded symbol of wrong type, yahasplugin.Plugin interface not implemented fully?")
	}

	yp := yPlugin{Name: name[:len(name)-3]}
	plugins[yp.Name] = &ypi
	yPlugins[yp.Name] = &yp
	coreLogger.InfoLn("loaded plugin", yp.Name)
	configFile, err := os.Open("./config/plugins/" + yp.Name + ".json")
	if err != nil {
		coreLogger.DebugLn("Could not open config file for "+yp.Name+":", err)
	} else {
		coreLogger.DebugLn("Opened config file for", name)
	}
	coreLogger.DebugLn("plugin permissions: ", coreconf.Permissions[name])
	return ypi.Init(&yp, logging.New(yp.Name, logging.StrToLvl(coreconf.Loglevels.Plugins[yp.Name])), configFile)
}
