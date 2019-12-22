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

type Namespace struct {
	Name string `json:"name"`
	Items []item.ItemData `json:"items"`
}

var logger *logging.Logger

var items *item.NamespaceMap

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
	router.HandleFunc("/items/{namespace}/{item}", getItemHandler).Methods("GET")
	router.HandleFunc("/items/{namespace}/{item}/{property}", getItemPropertyHandler).Methods("GET")
	router.HandleFunc("/items//{namespace}/{item}/state", setItemStateHandlerPut).Methods("PUT")
	router.HandleFunc("/items//{namespace}/{item}/state", setItemStateHandlerPost).Methods("POST")
	router.HandleFunc("/items", getItemsHandler).Methods("GET")
	router.HandleFunc("/items/{namespace}", getNamespaceHandler).Methods("GET")
	return nil
}

func (RPlugin) DeInit() error {
	// nothing to do
	return nil
}

func getNamespaceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	logger.InfoLn("received items request")
	w.Header().Set("Content-Type", "application/json")
	nsName := mux.Vars(r)["namespace"]
	ns, err := items.GetNamespace(nsName)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	itemDataArray := make([]item.ItemData, len(*ns))
	index := 0
	ns.ForEachItem(func(name string, itm *item.Item) {
			itemDataArray[index] = itm.Data()
			index++
	})
	json.NewEncoder(w).Encode(itemDataArray)
	logger.DebugLn("returned", itemDataArray)


}

func getItemsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	logger.InfoLn("received items request")
	w.Header().Set("Content-Type", "application/json")
	var nsArray []Namespace
	items.ForEachNamespace(func(name string, ns *item.Namespace) {
		namespace := Namespace{Name: name}
		ns.ForEachItem(func(name string, itm *item.Item) {
			namespace.Items = append(namespace.Items, itm.Data())
		})
		nsArray = append(nsArray, namespace)
	})
	json.NewEncoder(w).Encode(nsArray)
	logger.DebugLn("returned", nsArray)

}

func getItemHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	params := mux.Vars(r)
	ns := params["namespace"]
	name := params["item"]
	logger.InfoLn("received requerst for item:", ns + "/" + name)
	itm, err := items.GetItem(ns, name)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	json.NewEncoder(w).Encode(itm.Data())
}

func getItemPropertyHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	params := mux.Vars(r)
	ns := params["namespace"]
	name := params["item"]
	itemProperty := strings.Title(params["property"])
	logger.InfoLn("received request for", itemProperty, "of", name)
	logger.InfoLn("received requerst for item:", ns + "/" + name)
	itm, err := items.GetItem(ns, name)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	m := structs.Map(itm.Data())
	logger.DebugLn("created item map", m)
	prop := m[itemProperty]
	logger.DebugLn("returning", prop)
	fmt.Fprintf(w, "%s", prop)
}

func setItemStateHandlerPut(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	params := mux.Vars(r)
	ns := params["namespace"]
	name := params["item"]
	stateBytes, _ := ioutil.ReadAll(r.Body)
	state := string(stateBytes)
	defer r.Body.Close()
	logger.DebugLn("received request to update "+name+"'s state to", state)
	itm, err := items.GetItem(ns, name)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	itm.Send(state)
}

func setItemStateHandlerPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	ns := params["namespace"]
	name := params["item"]
	stateBytes, _ := ioutil.ReadAll(r.Body)
	state := string(stateBytes)
	defer r.Body.Close()
	logger.DebugLn("received request to update "+name+"'s state to", state)
	itm, err := items.GetItem(ns, name)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	itm.Send(state)
	json.NewEncoder(w).Encode(itm.Data())
}
