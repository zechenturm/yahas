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
	Loglevels   logConfig                  `json:"logging"`
	Permissions map[string]map[string]bool `json:"plugin-permissions"`
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
	coreconf = loadConfig()
	logging.InitLogging(logging.StrToLvl(coreconf.Loglevels.Default))
	coreLogger = logging.New("core", logging.StrToLvl(coreconf.Loglevels.Core))
	loader = item.NewLoader(coreLogger, &coreconf.Loglevels.Bindings)
	var err error
	Items, err = loader.LoadItems(itemFileDir)
	if err != nil {
		coreLogger.ErrorLn("Error loading items:", err)
	}
	mainRouter = mux.NewRouter()
	mainRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/"))))
	loadPlugins()
	http.ListenAndServe(":8000", mainRouter)
}

func loadConfig() coreConfig {
	bytes, err := ioutil.ReadFile("config/core.json")
	if err != nil {
		panic(err)
	}
	config := coreConfig{}
	if err != nil {
		panic(err)
	}
	json.Unmarshal(bytes, &config)
	fmt.Println("config:", config)
	return config
}
