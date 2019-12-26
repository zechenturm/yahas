package main

import (
	"encoding/json"
	"fmt"
	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

type coreConfig struct {
	Loglevels     logConfig                  `json:"logging"`
	Permissions   map[string]map[string]bool `json:"plugin-permissions"`
	IgnorePlugins []string                   `json:"ignore-plugins"`
}

type logConfig struct {
	Core     string            `json:"core"`
	Default  string            `json:"default"`
	Bindings map[string]string `json:"bindings"`
	Plugins  map[string]string `json:"plugins"`
}

var itemFileDir = "config/items"

var coreLogger *logging.Logger

var Items = make(item.NamespaceMap)

var coreconf coreConfig

var mainRouter *mux.Router

var loader *item.Loader

func main() {
	coreconf, err := loadConfig()
	if err != nil {
		logging.InitLogging(logging.DEFAULT)
		coreLogger = logging.New("core", logging.DEFAULT)
		coreLogger.ErrorLn("Error loading config:", err)
		return
	}

	logging.InitLogging(logging.StrToLvl(coreconf.Loglevels.Default))
	coreLogger = logging.New("core", logging.StrToLvl(coreconf.Loglevels.Core))
	loader = item.NewLoader(coreLogger, &coreconf.Loglevels.Bindings)
	Items, err = loader.LoadItems(itemFileDir)
	if err != nil {
		coreLogger.ErrorLn("Error loading items:", err)
		return
	}
	mainRouter = mux.NewRouter()
	mainRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/"))))
	loadPlugins(coreconf.IgnorePlugins)
	err = http.ListenAndServe(":8000", mainRouter)
	if err != nil {
		coreLogger.ErrorLn("HTTP server error:", err)
	}
}

func loadConfig() (coreConfig, error) {
	bytes, err := ioutil.ReadFile("config/core.json")
	if err != nil {
		return coreConfig{}, err
	}
	config := coreConfig{}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return coreConfig{}, err
	}
	fmt.Println("config:", config)
	return config, nil
}
