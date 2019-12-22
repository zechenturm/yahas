package main

import (
	"database/sql"
	"encoding/json"
	"github.com/zechenturm/yahas/item"
	"github.com/zechenturm/yahas/logging"
	"github.com/zechenturm/yahas/yahasplugin"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
)

type DBPlugin struct {
	itemHandles map[*item.Item]int64
	done        chan struct{}
}

var logger *logging.Logger

func (dbp *DBPlugin) Init(args yahasplugin.Provider, l *logging.Logger, configFile *os.File) error {
	logger = l
	dbp.itemHandles = make(map[*item.Item]int64)
	router, err := args.RequestRouter()
	if err != nil {
		return err
	}
	router.HandleFunc("/api/items/{item}/history", getItemHistoryHandler).Methods("POST")
	initDatabase(configFile)

	items, err := args.Items()
	if err != nil {
		return err
	}
	logger.DebugLn("subscribing to items")
	items.ForEachItem(func(ns, name string, itm *item.Item) {
		updateChan, handle := itm.Subscribe()
		dbp.itemHandles[itm] = handle
		go func(updateChan chan item.ItemData) {

			for {
				select {
				case update := <-updateChan:
					logger.DebugLn("received update", update)
					insertItemIntoDatabase(update)
				case <-dbp.done:
					logger.DebugLn("stopping DB routine")
					return
				}
			}
		}(updateChan)
	})
	return nil
}

func (dbp *DBPlugin) DeInit() error {
	logger.DebugLn("DeInit database plugin")
	for item, handle := range dbp.itemHandles {
		item.Unsubscribe(handle)
	}
	closeDB()
	return nil
}

func getItemHistoryHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	itemName := params["item"]
	bytes, _ := ioutil.ReadAll(r.Body)
	time := string(bytes)
	data := getItemHistoryFromDatabase(itemName, time)
	json.NewEncoder(w).Encode(data)
	r.Body.Close()
}

var Plugin DBPlugin

var db *sql.DB

var dbFileName = "./config/db.json"
var dbName = ""

type dbSettingsType struct {
	Host     string `json:"host"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type dataPointType struct {
	Time  string `json:"time"`
	Value string `json:"value"`
}

func loadDatabaseSettings(configFile *os.File) dbSettingsType {
	logger.DebugLn("reading db settings file")
	var settings dbSettingsType
	if err := json.NewDecoder(configFile).Decode(&settings); err != nil {
		logger.ErrorLn("error unmarshalling settings json:", err)
		return dbSettingsType{}
	}
	dbName = settings.Database

	return settings
}

func initDatabase(configFile *os.File) {
	settings := loadDatabaseSettings(configFile)
	var err error
	db, err = sql.Open("mysql", settings.User+":"+settings.Password+"@tcp("+settings.Host+")/"+settings.Database)
	if err != nil {
		logger.ErrorLn("error connecting to database:", err)
	}
	logger.DebugLn("opened database", db)
}

func checkIfTableExists(tableName string) bool {
	if db == nil {
		logger.WarnLn("database is nil!")
	}
	rows, err := db.Query("SHOW TABLES LIKE '" + tableName + "'")
	if err != nil {
		logger.ErrorLn("error querying database:", err)
		return false
	}
	if !rows.Next() {
		rows.Close()
		return false
	}
	rows.Close()
	return true
}

func insertItemIntoDatabase(it item.ItemData) {
	if db == nil {
		logger.WarnLn("database is nil, nothing written to db")
		return
	}
	if !checkIfTableExists(it.Name) {
		logger.DebugLn("table for " + it.Name + " not found, creating it")
		createTableForItem(it.Name)
	}
	if _, err := db.Exec("INSERT INTO " + it.Name + "(time, value) VALUES ('" + stringToDatetime(it.LastUpdated) + "', '" + it.State + "')"); err != nil {
		logger.ErrorLn("error inserting item "+it.Name+" into database:", err)
	}
}

func createTableForItem(itemName string) {
	if _, err := db.Exec("CREATE TABLE `" + dbName + "`.`" + itemName + "` ( `time` DATETIME NOT NULL , `value` VARCHAR(50) NOT NULL , PRIMARY KEY (`time`)) ENGINE = MyISAM;"); err != nil {
		logger.ErrorLn("error creating table for item "+itemName+":", err)
	}
}

func getItemHistoryFromDatabase(itemName, fromWhen string) []dataPointType {
	var data []dataPointType
	rows, err := db.Query("SELECT * FROM " + itemName + " WHERE time > '" + fromWhen + "' ORDER BY time")
	if err != nil {
		logger.ErrorLn("error getting history for item "+itemName+":", err)
		return data
	}
	defer rows.Close()
	for rows.Next() {
		var dataPoint dataPointType
		if err := rows.Scan(&dataPoint.Time, &dataPoint.Value); err != nil {
			logger.ErrorLn("error reading row", err)
			continue
		}
		data = append(data, dataPoint)
	}
	return data
}

func stringToDatetime(t string) string {
	tim, _ := time.Parse("15:04:05 02.01.2006", t)
	return tim.Format("2006-01-02 15:04:05")
}

func closeDB() {
	db.Close()
}
