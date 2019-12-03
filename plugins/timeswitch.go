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
	TimeItem       string `json:"time-item"`
	ControlledITem string `json:"controlled-item"`
}

type TimeSwitch struct {
	items   []timeItem
	itemMap *map[string]*item.Item
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
					timeItem := (*tm.itemMap)[itm.TimeItem].Data()
					strs := strings.Split(timeItem.State, ";")
					if len(strs) != 2 {
						l.ErrorLn("Failed to parse time item:", timeItem.Name, ":", timeItem.State)
						continue
					}
					if timeString == strs[0] {
						l.DebugLn(timeItem.Name, "set to ", strs[1])
						(*tm.itemMap)[itm.ControlledITem].Send(strs[1])
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
