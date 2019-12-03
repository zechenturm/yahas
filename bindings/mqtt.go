package main

import (
	"encoding/json"
	"errors"
	"github.com/zechenturm/yahas/logging"
	"os"

	"github.com/yosssi/gmq/mqtt"

	"github.com/yosssi/gmq/mqtt/client"
)

type mqttBinding struct {
	Items map[string]*mqttItem
}

var Binding mqttBinding

var logger *logging.Logger

var clnt = client.New(&client.Options{
	ErrorHandler: func(err error) {
		logger.ErrorLn(err)
	},
})

type mqttItem struct {
	Name      string
	RecTopic  string
	SendTopic string
	Send      chan string
	Receive   chan string
}

type initError struct {
	err string
}

func (s initError) Error() string {
	return s.err
}

type config struct {
	ClientID string `json:"client-id"`
	Server   string `jsson:"server"`
}

func (mi *mqttBinding) Init(l *logging.Logger, configFile *os.File) error {
	logger = l
	mi.Items = make(map[string]*mqttItem)
	defer clnt.Terminate()
	if configFile == nil {
		return initError{"Can not open config file!"}
	}
	conf := config{}
	json.NewDecoder(configFile).Decode(&conf)
	err := clnt.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  conf.Server,
		ClientID: []byte(conf.ClientID),
	})
	return err

}

func (mi *mqttBinding) RegisterItem(name string, params map[string]string) (*chan string, *chan string, error) {
	s := make(chan string)
	r := make(chan string)
	// check if specific send & receive channels were specified
	sendTopic, ok_send := params["topic_send"]
	recTopic, ok_rec := params["topic_rec"]
	// check if topic for both (use the same topic for send and receive) was specified
	bothTopic, ok_both := params["topic"]
	// if neither of the 3 is present, abort! MQTT needs a topic to publish and subscribe to
	if !ok_send && !ok_rec && !ok_both {
		return nil, nil, errors.New("No topic(s) specified!")
	}
	di := mqttItem{Name: name}
	// set send and receive topics, if send or receive is sspecified, use that, otherwise use topic_both a fallback
	if ok_send {
		di.SendTopic = sendTopic
	} else {
		di.SendTopic = bothTopic
	}
	if ok_rec {
		di.RecTopic = recTopic
	} else {
		di.RecTopic = bothTopic
	}

	logger.DebugLn("creating", di)
	// only send things if there is a topic to publish to
	if ok_send || ok_both {
		logger.DebugLn("initialising send")
		go func() {
			for msg := range s {
				logger.DebugLn("MQTT", name, "sending", msg)
				clnt.Publish(&client.PublishOptions{
					QoS:       mqtt.QoS0,
					TopicName: []byte(di.SendTopic),
					Message:   []byte(msg),
				})
			}
			logger.DebugLn("closing send for", name)
		}()
	} else {
		logger.DebugLn("no topic to send / publish to")
		go func() {
			for msg := range s {
				logger.ErrorLn(name, "received message to send but no topic to send it to! message:", msg)
			}
			logger.DebugLn("closing receive for", name)
		}()
	}

	if ok_rec || ok_both {
		logger.DebugLn("initialising receive")
		err := clnt.Subscribe(&client.SubscribeOptions{
			SubReqs: []*client.SubReq{
				&client.SubReq{
					TopicFilter: []byte(di.RecTopic),
					QoS:         mqtt.QoS0,
					Handler: func(topicName, message []byte) {
						logger.DebugLn(string(topicName), ":", string(message))
						r <- string(message)
					},
				},
			},
		})
		if err != nil {
			return nil, nil, err
		}
	} else {
		logger.DebugLn("no topic to receive from / subscribe to")
	}
	di.Send = s
	di.Receive = r
	mi.Items[name] = &di
	return &s, &r, nil
}

func (mb *mqttBinding) UnregisterItem(name string) {
	logger.DebugLn("unregistering", name)
	item, ok := mb.Items[name]
	if !ok {
		logger.DebugLn("item not registered")
		return
	}
	logger.DebugLn("unsubscribing", name)
	topics := make([][]byte, 1)
	topics[0] = []byte(item.RecTopic)
	clnt.Unsubscribe(&client.UnsubscribeOptions{TopicFilters: topics})
	close(item.Send)
	close(item.Receive)
}

func (mb *mqttBinding) DeInit() error {
	logger.DebugLn("DeInit()")
	logger.DebugLn(len(mb.Items))
	for name := range mb.Items {
		mb.UnregisterItem(name)
	}
	logger.DebugLn("DeInit() done")
	return nil
}
