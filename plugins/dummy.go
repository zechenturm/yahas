package main

import (
	"fmt"
	"github.com/zechenturm/yahas/logging"
	"github.com/zechenturm/yahas/yahasplugin"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var logger *logging.Logger

type DummyPlugin struct {
	stop   chan struct{}
	router *mux.Router
}

func (d *DummyPlugin) Init(args yahasplugin.Provider, l *logging.Logger, configFile *os.File) error {
	logger = l
	logger.InfoLn("dummy init!")
	if configFile != nil {
		logger.WarnLn("Found config file but none needed")
	} else {
		logger.DebugLn("No config file and none needed")
	}
	router, err := args.RequestRouter()
	if err != nil {
		logger.ErrorLn("Error requesting router:", err)
		return err
	}
	d.router = router
	d.router.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		logger.DebugLn("HTTP Request")
		fmt.Fprintln(w, "Hello World")
		r.Body.Close()
	})

	items, err := args.Items()
	if err != nil {
		logger.ErrorLn("Error requesting items:", err)
		return err
	}
	d.stop = make(chan struct{})
	go func() {
		itemChan, id := (*(*items)["humidity3"]).Subscribe()
		logger.DebugLn("Subscribed to humidity3 with id", id)
		for {
			select {
			case update := <-itemChan:
				logger.InfoLn("dummmy received:", update)
			case <-d.stop:
				logger.DebugLn("Stopping receive routine")
				(*(*items)["humidity3"]).Unsubscribe(id)
				return
			}
		}
	}()

	return nil
}

func (d *DummyPlugin) DeInit() error {
	logger.DebugLn("DeInit")
	close(d.stop)

	return nil
}

var Plugin DummyPlugin
