package item

import (
	"errors"
	"github.com/zechenturm/yahas/logging"
	"os"
	"plugin"
)

type Binding interface {
	Init(*logging.Logger, *os.File) error // initialise Binding
	DeInit() error
	RegisterItem(string, map[string]string) (*chan string, *chan string, error) // RegisterItem(name) (send, receive, error)
	UnregisterItem(string)
}

type itemBinding struct {
	config  itemBindingConfig
	Receive *chan string
	Send    *chan string
}

type itemBindingConfig struct {
	Name     string            `json:"name"`
	Settings map[string]string `json:"settings"`
}

type BindingLoader interface {
	Load(name string, to string) (Binding, error)
}

type BindingManager interface {
	LoadBinding(name string) error
	UnloadBinding(name string) error
}

type bManager struct {
	logger    *logging.Logger
	logLevels *map[string]string
	bindings  map[string]Binding
}

func NewBManager(logger *logging.Logger, logLevels *map[string]string) *bManager {
	return &bManager{
		logger:    logger,
		logLevels: logLevels,
		bindings:  make(map[string]Binding),
	}
}

func (bm *bManager) Load(name string) (Binding, error) {
	b, ok := bm.bindings[name]
	if !ok {
		return bm.loadFromDisk(name)
	}
	return b, nil
}

func (bm *bManager) UnloadBinding(name string) error {
	bm.logger.DebugLn("unloading", name)
	b, ok := bm.bindings[name]
	if !ok {
		return errors.New("Binding not loaded")
	}
	delete(bm.bindings, name)
	return b.DeInit()
}

func (bm *bManager) LoadBinding(name string) error {
	_, ok := bm.bindings[name]
	if !ok {
		_, err := bm.loadFromDisk(name)
		return err
	}
	return nil
}

func (bm *bManager) loadFromDisk(name string) (Binding, error) {
	var b Binding
	p, err := plugin.Open("./bindings/" + name + ".so")
	if err != nil {
		return b, err
	}

	sb, err := p.Lookup("Binding")
	if err != nil {
		return b, err
	}
	b, ok := sb.(Binding)
	if !ok {
		return b, errors.New("loaded symbol of wrong type, Binding interface not implemented fully?")
	}

	bm.bindings[name] = b
	configFile, err := os.Open("./config/bindings/" + name + ".json")
	if err != nil {
		bm.logger.DebugLn("Could not open config file for "+name+":", err)
	} else {
		bm.logger.DebugLn("Opened config file for", name)
	}
	logLevel := logging.StrToLvl((*bm.logLevels)[name])
	err = b.Init(logging.New(name, logLevel), configFile)
	return b, err
}
