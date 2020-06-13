package main

import (
	"io/ioutil"
	"path"

	"github.com/zechenturm/yahas/yahasplugin"

	"github.com/zechenturm/yahas/service"
)

var serviceManager service.ServiceManager

func loadPlugins(ignore []string) error {
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

	pm := yahasplugin.PluginManager{BindingMan: loader.BindingManager, Logger: coreLogger, Permissions: coreconf.Permissions, LogLevels: coreconf.Loglevels.Plugins, Items: &Items}

	for _, file := range files {
		baseName := file.Name()[:len(file.Name())-len(".so")]
		if path.Ext(file.Name()) == ".so" && !shouldIgnore(baseName) {
			err := pm.LoadPlugin(file.Name())
			if err != nil {
				coreLogger.ErrorLn("Error loading plugin", file.Name(), ":", err)
			}
		}
	}
	return nil
}
