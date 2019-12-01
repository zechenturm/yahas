package main

import (
	"encoding/json"
	"fmt"
	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
	"io/ioutil"
	"net/http"
	"os"

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

var itemFileName = "config/items.json"

var coreLogger *logging.Logger

var itemNames []string

var ItemArray []item.Item
var Items = make(map[string]*item.Item)

var coreconf coreConfig

var mainRouter *mux.Router

var loader *item.Loader

func main() {
	coreconf = loadConfig()
	logging.InitLogging(logging.StrToLvl(coreconf.Loglevels.Default))
	coreLogger = logging.New("core", logging.StrToLvl(coreconf.Loglevels.Core))
	loader = item.NewLoader(coreLogger, &coreconf.Loglevels.Bindings)
	configFile, err := os.Open(itemFileName)
	if err != nil {
		coreLogger.ErrorLn("error opening config file:", err)
		return
	}
	Items = loader.LoadItems(configFile)
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
