package main

import (
	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
	"github.com/zechenturm/yahas/yahasplugin"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
)

type SitePlugin struct {
}

var Plugin SitePlugin
var logger *logging.Logger
var items *item.NamespaceMap

type ItemsStuct struct {
	Title    string
	Sitemaps []string
}

func (SitePlugin) Init(args yahasplugin.Provider, l *logging.Logger, configFile *os.File) error {
	router, err := args.RequestRouter()
	if err != nil {
		return err
	}
	logger = l
	items, err = args.Items()
	if err != nil {
		return err
	}
	router.HandleFunc("/{site}", siteHandler)
	router.HandleFunc("/{site}/html", siteHTMLHandler)
	return nil
}

func (SitePlugin) DeInit() error {
	//nothing to do
	return nil
}

func siteHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger.DebugLn("received requerst for site:" + params["site"])
	t, err := template.New("main.html").Funcs(template.FuncMap{
		"getItem": func(namespace, name string) item.ItemData { itm, _ := items.GetItem(namespace, name); return itm.Data()},
		"split":   strings.Split,
	}).ParseGlob("templates/disp/*.html")
	t, err = t.ParseFiles("templates/sites/"+params["site"]+".html", "templates/main.html")
	if err != nil {
		logger.ErrorLn(err)
	}
	var sitemaps []string
	err = filepath.Walk("templates/sites", func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".html" {
			sitemaps = append(sitemaps, strings.Title(info.Name()[:len(info.Name())-len(".html")]))
		}
		return nil
	})
	err = t.Execute(w, ItemsStuct{Title: strings.Title(params["site"]), Sitemaps: sitemaps})
	if err != nil {
		logger.ErrorLn(err)
	}
}

func siteHTMLHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger.DebugLn("received html requerst for site:" + params["site"])
	t, err := template.New(params["site"] + ".html").Funcs(template.FuncMap{
		"getItem": func(namespace, name string) item.ItemData { itm, _ := items.GetItem(namespace, name); return itm.Data()},
		"split":   strings.Split,
	}).ParseGlob("templates/disp/*.html")
	t, err = t.ParseFiles("templates/sites/" + strings.ToLower(params["site"]) + ".html")
	if err != nil {
		logger.ErrorLn(err)
	}
	err = t.ExecuteTemplate(w, "sitemap", ItemsStuct{Title: strings.Title(params["site"]), Sitemaps:[]string{params["site"]}})
	if err != nil {
		logger.ErrorLn(err)
	}
}
