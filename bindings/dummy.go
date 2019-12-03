package main

import (
	"github.com/zechenturm/yahas/logging"
	"os"
	"strconv"
	"time"
)

type dummy struct {
	Items    map[string]dummyItem
	Messages []string
}

var Binding dummy

var logger *logging.Logger

type dummyItem struct {
	Send chan string
	Rec  chan string
}

func (*dummy) Init(l *logging.Logger, configFile *os.File) error {
	logger = l
	logger.InfoLn("init dummy binding!")
	if configFile != nil {
		logger.WarnLn("Found config file but none needed")
	} else {
		logger.DebugLn("No config file and none needed")
	}
	for i := 0; i < 10; i++ {
		Binding.Messages = append(Binding.Messages, "msg "+strconv.Itoa(i))
	}

	go func() {
		time.Sleep(10 * time.Second)
		for _, msg := range Binding.Messages {
			for _, item := range Binding.Items {
				logger.InfoLn("dummy sending")
				item.Rec <- msg
			}
			time.Sleep(time.Second)
		}
	}()
	return nil
}

func (d *dummy) RegisterItem(name string, params map[string]string) (*chan string, *chan string, error) {
	s := make(chan string)
	r := make(chan string)
	di := dummyItem{s, r}
	Binding.Items[name] = di
	go func() {
		for msg := range s {
			logger.InfoLn(name, "received", msg)
		}
	}()
	return &s, &r, nil
}

func (d *dummy) UnregisterItem(name string) {
	delete(d.Items, name)
}

func (d *dummy) DeInit() error {
	for name := range d.Items {
		d.UnregisterItem(name)
	}
	return nil
}
