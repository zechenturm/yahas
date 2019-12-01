package main

import (
	"encoding/json"
	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
	"github.com/zechenturm/yahas/yahasplugin"
	"net/http"
	"os"
	"sort"

	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var connections []*websocket.Conn
var updates = make(chan item.ItemData, 10)
var logger *logging.Logger

type WSPlugin struct {
	handles map[*item.Item]int64
	done    chan struct{}
}

func (wp *WSPlugin) Init(args yahasplugin.Provider, l *logging.Logger, configFile *os.File) error {
	logger = l
	logger.InfoLn("websocket init!")
	if configFile != nil {
		logger.WarnLn("Found config file but none needed")
	}

	wp.handles = make(map[*item.Item]int64)
	wp.done = make(chan struct{})

	router, err := args.RequestRouter()
	if err != nil {
		return err
	}
	router.HandleFunc("", websocketHandler)

	items, err := args.Items()
	if err != nil {
		return err
	}

	for _, itm := range *items {
		updateChan, handle := itm.Subscribe()
		wp.handles[itm] = handle
		go func(updateChan chan item.ItemData) {
			for update := range updateChan {
				logger.DebugLn("sending update for", update.Name)
				updates <- update
			}
		}(updateChan)
	}
	go websocketManager(wp.done)

	return nil
}

func (wp *WSPlugin) DeInit() error {
	close(wp.done)
	for _, conn := range connections {
		err := conn.Close()
		if err != nil {
			// while this is an error, it probably doesn't concern the end use
			// probably means the connection was already dead
			// because the browser window was closed
			logger.DebugLn("error closing conection:", err)
		}
	}
	for itm, handle := range wp.handles {
		itm.Unsubscribe(handle)
	}
	// TODO: Should we return some sort of error if there were errors closing a connection?
	return nil
}

var Plugin WSPlugin

func sendWebSocketUpdateMessage(c *websocket.Conn, update item.ItemData) error {
	logger.DebugLn("sending update:", update)
	message, err := json.Marshal(update)
	if err != nil {
		logger.ErrorLn("error marschalling item", update)
	}
	err = c.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		logger.ErrorLn("error sending update over websocket:", err)
		c.Close()
		return err
	}
	return nil
}

func websocketManager(done chan struct{}) {
	for {
		select {
		case update := <-updates:
			var indToRemove []int
			for cindex, c := range connections {
				err := sendWebSocketUpdateMessage(c, update)
				if err != nil {
					indToRemove = append(indToRemove, cindex)
				}
			}
			// sort descending so that the highest index is removed first
			// this way the indices still pending remmmoval are unaffected by the changing indices
			sort.Sort(sort.Reverse(sort.IntSlice(indToRemove)))
			logger.DebugLn("indices to remove: ", indToRemove)
			for _, index := range indToRemove {
				logger.DebugLn("removing connection")
				if index == len(connections) {
					// remove last connection
					connections = connections[0 : index-1]
				} else {
					// only try appending index+1... if this/these actually exist
					connections = append(connections[:index], connections[index+1:]...)
				}
			}
		case <-done:
			logger.DebugLn("stopping websocket manager")
			return
		}
	}
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	logger.DebugLn("Websocket upgrade request")
	c, err := websocketUpgrader.Upgrade(w, r, nil)
	defer r.Body.Close()
	if err != nil {
		logger.ErrorLn("Error upgrading connection to websocket:", err)
		return
	}
	connections = append(connections, c)
}
