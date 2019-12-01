package main

import (
	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
	"github.com/zechenturm/yahas/yahasplugin"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type ReloadPlugin struct {
	logger *logging.Logger
	pm     yahasplugin.Manager
	bm     item.BindingManager
}

var Plugin ReloadPlugin

func (rp *ReloadPlugin) Init(provider yahasplugin.Provider, logger *logging.Logger, config *os.File) error {
	man, err := provider.RequestPlugins()
	if err != nil {
		return err
	}
	rp.pm = man

	bm, err := provider.BindingManager()
	if err != nil {
		return err
	}
	rp.bm = bm

	rp.logger = logger
	logger.DebugLn("plugins: ", rp.pm.Map())

	router, err := provider.RequestRouter()
	if err != nil {
		return err
	}

	router.HandleFunc("/plugin/{plugin}/unload", unloadModuleHandler(rp.logger, rp.pm))
	router.HandleFunc("/plugin/{plugin}/load", loadModuleHandler(rp.logger, rp.pm))
	router.HandleFunc("/binding/{binding}/unload", unloadBindingHandler(rp.logger, rp.bm))
	router.HandleFunc("/binding/{binding}/load", loadBindingHandler(rp.logger, rp.bm))
	return nil
}

func unloadModuleHandler(logger *logging.Logger, pm yahasplugin.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		logger.DebugLn("unloading", params["plugin"])
		err := pm.Unload(params["plugin"])
		if err != nil {
			logger.ErrorLn("Error unloading "+params["plugin"]+":", err)
		}
	}
}

func loadModuleHandler(logger *logging.Logger, pm yahasplugin.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		logger.DebugLn("loading", params["plugin"])
		err := pm.Load(params["plugin"])
		if err != nil {
			logger.ErrorLn("Error loading "+params["plugin"]+":", err)
		}
	}
}

func (*ReloadPlugin) DeInit() error {
	return nil
}

func unloadBindingHandler(logger *logging.Logger, bm item.BindingManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		name := params["binding"]
		logger.DebugLn("unloading binding", name)
		err := bm.UnloadBinding(name)
		if err != nil {
			logger.ErrorLn("error loading binding", name, err)
		}
	}
}

func loadBindingHandler(logger *logging.Logger, bm item.BindingManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		name := params["binding"]
		logger.DebugLn("unloading binding", name)
		err := bm.LoadBinding(name)
		if err != nil {
			logger.ErrorLn("error loading binding", name, err)
		}
	}
}
