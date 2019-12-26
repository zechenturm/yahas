package item

import (
	"encoding/json"
	"github.com/zechenturm/yahas/logging"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Namespace map[string]*Item

type NamespaceMap map[string]*Namespace

type NotFoundError struct {
	itemName string
}

type AlreadyLoadedError struct {
	name string
}

func (nfe NotFoundError) Error() string {
	return "item " + nfe.itemName + "not found"
}

func (ale AlreadyLoadedError) Error() string {
	return ale.name + " already loaded"
}

func (ns *Namespace) Get(name string) (*Item, error) {
	itm, ok := (*ns)[name]
	if !ok {
		return nil, NotFoundError{name}
	}
	return itm, nil

}

func (ns *Namespace) ForEachItem(function func(name string, itm *Item)) {
	for itmName, itm := range *ns {
		function(itmName, itm)
	}
}

func (nm *NamespaceMap) GetNamespace(name string) (*Namespace, error) {
	ns, ok := (*nm)[name]
	if !ok {
		return nil, NotFoundError{name}
	}
	return ns, nil
}

func (nm *NamespaceMap) GetItem(namespace, name string) (*Item, error) {
	ns, err := nm.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.Get(name)
}

func (nm *NamespaceMap) loadNamespace(dirPath, name string, bm *bManager) error {
	if _, ok := (*nm)[name]; ok {
		return AlreadyLoadedError{name}
	}
	var itemDataArray []ItemData
	nsFile, err := os.Open(filepath.Join(dirPath, name+".json"))
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(nsFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, &itemDataArray); err != nil {
		return err
	}
	itemArray := make([]Item, len(itemDataArray))
	namespace := make(Namespace)
	for index, data := range itemDataArray {
		err = itemArray[index].New(data, bm)
		if err != nil {
			return err
		}
		namespace[data.Name] = &itemArray[index]
	}
	(*nm)[name] = &namespace
	return nil
}

func readNamespaces(dirPath string) (*[]string, error) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, file := range files {
		name := file.Name()
		if name[len(name)-len(".json"):] != ".json" {
			continue
		}
		names = append(names, name[:len(name)-len(".json")])
	}
	return &names, nil
}

func (nm *NamespaceMap) LoadAllNamespaces(dirPath string, bm *bManager, l *logging.Logger) error {
	l.DebugLn("looking for items in:", dirPath)
	names, err := readNamespaces(dirPath)
	if err != nil {
		return err
	}
	l.DebugLn("read namespace names:", names)
	for _, name := range *names {
		err = nm.loadNamespace(dirPath, name, bm)
		if err != nil {
			break
		}
	}
	return err
}

func (nm *NamespaceMap) ForEachNamespace(function func(name string, ns *Namespace)) {
	for name, ns := range *nm {
		function(name, ns)
	}
}

func (nm *NamespaceMap) ForEachItem(function func(ns, name string, itm *Item)) {
	for nsName, ns := range *nm {
		for itmName, itm := range *ns {
			function(nsName, itmName, itm)
		}
	}
}
