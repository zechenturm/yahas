package main

import (
	"encoding/json"
	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
	"github.com/zechenturm/yahas/yahasplugin"
	"os"
	"strings"
	"time"
)

type timeItem struct {
	TimeItem        itemData   `json:"time-item"`
	ControlledItems []itemData `json:"controlled-item"`
}

type itemData struct {
	Namespace string `json:"namespace"`
	Name string `json:"name"`
}

type TimeSwitch struct {
	items   []timeItem
	itemMap *item.NamespaceMap
	stop    chan struct{}
	timer   chan struct{}
}

var Plugin TimeSwitch

func (tm *TimeSwitch) Init(provider yahasplugin.Provider, l *logging.Logger, configFile *os.File) error {
	err := json.NewDecoder(configFile).Decode(&tm.items)
	if err != nil {
		return err
	}
	l.DebugLn("read items:", tm.items)

	tm.itemMap, err = provider.Items()
	if err != nil {
		return err
	}

	go func() {
		timer := time.Tick(1 * time.Minute)
		for {
			select {
			case t := <-timer:
				l.DebugLn("check", t)
				timeString := t.Format("15:04")
				for _, itm := range tm.items {
					ptr, err := tm.itemMap.GetItem(itm.TimeItem.Namespace, itm.TimeItem.Name)
					if err != nil {
						l.ErrorLn(err)
						return
					}
					timeItem := ptr.Data()
					strs := strings.Split(timeItem.State, ";")
					if len(strs) != 2 {
						l.ErrorLn("Failed to parse time item:", timeItem.Name, ":", timeItem.State)
						continue
					}
					if timeString == strs[0] {
						l.DebugLn(timeItem.Name, "set to ", strs[1])
						for _, it := range itm.ControlledItems {
							i, err := tm.itemMap.GetItem(it.Namespace, it.Name)
							if err != nil {
								l.ErrorLn(it.Namespace, it.Name, err)
								continue
							}
							i.Send(strs[1])
						}
					}
				}
			case <-tm.stop:
				l.DebugLn("stopping")
				return
			}
		}
	}()

	return err
}

func (tm *TimeSwitch) DeInit() error {
	tm.stop <- struct{}{}
	return nil
}
