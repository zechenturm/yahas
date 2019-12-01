package main

import (
	"encoding/json"
	"fmt"
	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
	"github.com/zechenturm/yahas/yahasplugin"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/structs"

	"github.com/gorilla/mux"
)

type RPlugin struct {
}

var logger *logging.Logger

var items *map[string]*item.Item

var Plugin RPlugin

func (RPlugin) Init(args yahasplugin.Provider, l *logging.Logger, configFile *os.File) error {
	router, err := args.RequestRouter()
	if err != nil {
		return err
	}
	items, err = args.Items()
	if err != nil {
		return err
	}
	logger = l
	router.HandleFunc("/items/{item}", getItemHandler).Methods("GET")
	router.HandleFunc("/items/{item}/{property}", getItemPropertyHandler).Methods("GET")
	router.HandleFunc("/items/{item}/state", setItemStateHandlerPut).Methods("PUT")
	router.HandleFunc("/items/{item}/state", setItemStateHandlerPost).Methods("POST")
	router.HandleFunc("/items", getItemsHandler).Methods("GET")
	return nil
}

func (RPlugin) DeInit() error {
	// nothing to do
	return nil
}

func getItemsHandler(w http.ResponseWriter, r *http.Request) {
	logger.InfoLn("received items request")
	w.Header().Set("Content-Type", "application/json")
	itemDataArray := make([]item.ItemData, len(*items))
	index := 0
	for _, item := range *items {
		itemDataArray[index] = item.Data()
		index++
	}
	json.NewEncoder(w).Encode(itemDataArray)
	logger.DebugLn("returned", itemDataArray)
	r.Body.Close()

}

func getItemHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger.InfoLn("received requerst for item:" + params["item"])
	json.NewEncoder(w).Encode((*items)[params["item"]].Data())
	r.Body.Close()
}

func getItemPropertyHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	itemName := params["item"]
	itemProperty := strings.Title(params["property"])
	logger.InfoLn("received request for", itemProperty, "of", itemName)
	m := structs.Map((*items)[itemName].Data())
	logger.DebugLn("created item map", m)
	prop := m[itemProperty]
	logger.DebugLn("returning", prop)
	fmt.Fprintf(w, "%s", prop)
	r.Body.Close()
}

func setItemStateHandlerPut(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	itemName := params["item"]
	stateBytes, _ := ioutil.ReadAll(r.Body)
	state := string(stateBytes)
	defer r.Body.Close()
	logger.DebugLn("received request to update "+itemName+"'s state to", state)
	it := (*items)[itemName]
	it.Send(state)
}

func setItemStateHandlerPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	itemName := params["item"]
	stateBytes, _ := ioutil.ReadAll(r.Body)
	state := string(stateBytes)
	defer r.Body.Close()
	logger.DebugLn("received request to update "+itemName+"'s state to", state)
	it := (*items)[itemName]
	it.Send(state)
	json.NewEncoder(w).Encode(it.Data())
}
