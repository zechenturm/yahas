package item

import (
	"github.com/zechenturm/yahas/logging"
	"math/rand"
	"sync"
	"time"
)

type Item struct {
	data        ItemData
	binding     itemBinding
	subscribers map[int64]chan ItemData
	mutex       sync.Mutex
	logger      *logging.Logger
}

type ItemData struct {
	Namespace    string
	Name         string            `json:"name"`
	Label        string            `json:"label"`
	State        string            `json:"state"`
	LastUpdated  string            `json:"updated"`
	Prefix       string            `json:"prefix"`
	Suffix       string            `json:"suffix"`
	Binding      itemBindingConfig `json:"Binding"`        // name of the Binding to be loaded should be stored under the "name" key
	UpdateOnSend bool              `json:"update_on_send"` // update state in Send() and Receive() if true and only in Receive() if false
}

type Loader struct {
	logger         *logging.Logger
	logLevels      *map[string]string
	BindingManager *bManager
}

func NewLoader(logger *logging.Logger, bindingLogLevels *map[string]string) *Loader {
	return &Loader{
		logger:         logger,
		logLevels:      bindingLogLevels,
		BindingManager: NewBManager(logger, bindingLogLevels),
	}
}

func (l *Loader) LoadItems(dirPath string) (NamespaceMap, error) {
	nsMap := make(NamespaceMap)
	err := nsMap.LoadAllNamespaces(dirPath, l.BindingManager, l.logger)
	return nsMap, err
}

func (it *Item) New(id ItemData, manager *bManager) error {

	it.data = id
	it.logger = manager.logger
	it.subscribers = make(map[int64]chan ItemData)
	if id.Binding.Name != "" {
		bind, err := manager.Load(id.Binding.Name)
		if err != nil {
			manager.logger.ErrorLn("Error loading Binding", id.Binding.Name, "for item", id.Name)
			return err
		}
		it.binding.Send, it.binding.Receive, err = bind.RegisterItem(id.Name, id.Binding.Settings)
		if err != nil {
			manager.logger.ErrorLn("Error registering item", id.Name)
			return err
		}
		go it.receive()
	}
	return nil
}

func (it *Item) Subscribe() (chan ItemData, int64) {

	id := rand.Int63()
	ch := make(chan ItemData)
	it.mutex.Lock()
	it.subscribers[id] = ch
	it.mutex.Unlock()

	return ch, id
}

func (it *Item) Unsubscribe(id int64) {
	it.mutex.Lock()
	delete(it.subscribers, id)
	it.mutex.Unlock()
}

func (it *Item) broadcast() {
	it.mutex.Lock()
	for _, channel := range it.subscribers {
		channel <- it.data
	}
	it.mutex.Unlock()
}

func (id *ItemData) update(state string) {
	id.LastUpdated = time.Now().Format("15:04:05 02.01.2006")
	id.State = state
}

func (it *Item) receive() {
	for state := range *it.binding.Receive {
		it.data.update(state)
		it.logger.InfoLn(it.data.Name, "received", state)
		go it.broadcast()
	}
}

func (it *Item) Send(state string) {
	if it.data.UpdateOnSend {
		it.data.update(state)
		it.logger.DebugLn(it.data.Name, "updating on send")
		go it.broadcast()
	}
	if it.binding.Send != nil {
		*it.binding.Send <- state
	}

	it.logger.InfoLn(it.data.Name, "sending", state)
}

func (it *Item) Data() ItemData {
	return it.data
}
