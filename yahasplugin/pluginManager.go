package yahasplugin

import (
	"errors"
	"os"
	"plugin"

	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
)

type PluginManager struct {
	Logger      *logging.Logger
	BindingMan  item.BindingManager
	Permissions map[string]map[string]bool
	LogLevels   map[string]string
	Items       *item.NamespaceMap
	plugins     map[string]*yPlugin
}

func (pm *PluginManager) Load(name string) error {
	if _, ok := pm.plugins[name]; ok {
		return errors.New("Plugin alreeady loaded")
	}
	return pm.LoadPlugin(name + ".so")
}

func (pm *PluginManager) Unload(name string) error {
	pm.Logger.DebugLn("unlading " + name)
	pl := pm.plugins[name]
	delete(pm.plugins, name)
	return pl.plugin.DeInit()
}

func (pm *PluginManager) LoadPlugin(name string) error {
	if pm.plugins == nil {
		pm.plugins = make(map[string]*yPlugin)
	}
	pm.Logger.InfoLn("loading plugin", name)
	p, err := plugin.Open("./plugins/" + name)
	if err != nil {
		return err
	}

	sb, err := p.Lookup("Plugin")
	if err != nil {
		return err
	}
	ypi, ok := sb.(Plugin)
	if !ok {
		return errors.New("loaded symbol of wrong type, yahasplugin.Plugin interface not implemented fully?")
	}

	baseName := name[:len(name)-len(".so")]
	yp := yPlugin{Name: baseName, bindingManager: pm.BindingMan, items: pm.Items, plugin: ypi}
	pm.plugins[baseName] = &yp
	pm.Logger.InfoLn("loaded plugin", name)
	return nil
}

func (pm *PluginManager) initPlugin(name string) error {

	yp := pm.plugins[name]
	ypi := yp.plugin

	configFile, err := os.Open("./config/plugins/" + yp.Name + ".json")
	if err != nil {
		pm.Logger.DebugLn("Could not open config file for "+yp.Name+":", err)
	} else {
		pm.Logger.DebugLn("Opened config file for", name)
	}
	pm.Logger.DebugLn("plugin permissions: ", pm.Permissions[name])
	return ypi.Init(yp, logging.New(yp.Name, logging.StrToLvl(pm.LogLevels[yp.Name])), configFile)
}
