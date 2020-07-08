package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/zechenturm/yahas/item"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/zechenturm/yahas/logging"
	"github.com/zechenturm/yahas/yahasplugin"
)

var logger *logging.Logger

type Telegram struct {
	subscriptions map[string]int64
	items         *item.NamespaceMap
	stop          chan struct{}
	bot           *tgbotapi.BotAPI
	updates       tgbotapi.UpdatesChannel
}

type config struct {
	APIToken string `json:"api-token"`
}

func (tg *Telegram) Init(args yahasplugin.Provider, l *logging.Logger, configFile *os.File) error {
	tg.subscriptions = make(map[string]int64)

	logger = l
	logger.DebugLn("init telegram binding!")
	if configFile == nil {
		return errors.New("no config file found")
	}

	var err error
	tg.items, err = args.Items()
	if err != nil {
		return err
	}

	conf := config{}
	err = json.NewDecoder(configFile).Decode(&conf)
	if err != nil {
		return err
	}

	tg.bot, err = tgbotapi.NewBotAPI(conf.APIToken)
	if err != nil {
		return err
	}

	tg.bot.Debug = false

	logger.DebugLn("Authorized on account", tg.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	tg.updates, err = tg.bot.GetUpdatesChan(u)

	go tg.updateLoop()

	return nil

	return nil
}

func (tg *Telegram) DeInit() error {
	logger.DebugLn("DeInit")
	close(tg.stop)

	return nil
}

func (tg *Telegram) updateLoop() {
	for update := range tg.updates {
		if update.Message == nil {
			continue
		}

		logger.DebugLn("[", update.Message.Chat.ID, ",", update.Message.From.ID, "]", update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		if update.Message.IsCommand() {
			tg.handleCommand(update.Message)
		} else {
			tg.bot.Send(msg)
		}
	}
}

func (tg *Telegram) handleCommand(message *tgbotapi.Message) {

	switch message.Command() {
	case "get":
		tg.handleGet(message)
	case "watch":
		tg.handleWatch(message)
	case "unwatch":
		tg.handleUnWatch(message)
	}
}

func (tg *Telegram) handleGet(message *tgbotapi.Message) {

	itm, err := tg.itemFromArgs(message)
	if err != nil {
		tg.sendString(message.Chat.ID, err.Error())
		return
	}

	tg.sendString(message.Chat.ID, itm.Data().Label+": "+itm.Data().State+itm.Data().Suffix)
	return
}

func (tg *Telegram) handleWatch(message *tgbotapi.Message) {
	itm, err := tg.itemFromArgs(message)
	if err != nil {
		tg.sendString(message.Chat.ID, err.Error())
	}

	channel, id := itm.Subscribe()

	tg.subscriptions[itm.Data().Name] = id

	go func() {
		for {
			select {
			case <-tg.stop:
				return
			case update := <-channel:
				tg.sendString(message.Chat.ID, update.Label+": "+update.State+update.Suffix)
			}
		}
	}()
}

func (tg *Telegram) handleUnWatch(message *tgbotapi.Message) {
	itm, err := tg.itemFromArgs(message)
	if err != nil {
		tg.sendString(message.Chat.ID, err.Error())
	}
	itm.Unsubscribe(tg.subscriptions[itm.Data().Name])
}

func (tg *Telegram) itemFromArgs(message *tgbotapi.Message) (*item.Item, error) {

	itemName := message.CommandArguments()
	if len(itemName) == 0 {
		return nil, errors.New("please specify an item")
	}

	itm, err := tg.items.GetItem("items", itemName)
	if err != nil {
		return nil, err
	}
	return itm, nil
}

func (tg *Telegram) sendString(id int64, text string) {
	msg := tgbotapi.NewMessage(id, text)
	tg.bot.Send(msg)
}

var Plugin Telegram
