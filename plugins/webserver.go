package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/zechenturm/yahas/logging"
	"github.com/zechenturm/yahas/yahasplugin"
)

var logger *logging.Logger

type Webserver struct {
	mainRouter *mux.Router
	logger     *logging.Logger
}

func (ws *Webserver) Init(provider yahasplugin.Provider, l *logging.Logger, configFile *os.File) error {
	ws.logger = l
	l.DebugLn("initialising")
	ws.mainRouter = mux.NewRouter()
	ws.mainRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/"))))
	l.DebugLn("registering service")
	err := provider.Register("webserver", ws)
	l.DebugLn("init done")
	go func() {
		err = http.ListenAndServe(":8000", ws.mainRouter)
		if err != nil {
			l.ErrorLn("HTTP server error:", err)
		}
	}()
	return err
}

func (ws *Webserver) Name() string {
	return "webserver"
}

func (ws *Webserver) ProvidedBy() string {
	return "webserver"
}

func (ws *Webserver) SubRouter(prefix string) {
	ws.mainRouter.PathPrefix(prefix).Subrouter()
}

func (ws *Webserver) DeInit() error {
	return nil
}

var Plugin Webserver
