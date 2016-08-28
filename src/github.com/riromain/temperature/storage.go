package main

import (
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
	"database/sql"
	"net/http"
	"strconv"
	"encoding/json"
)

type TempEntry struct {
	Channel string
	Date    time.Time
	Temp    float32
}

var db *sql.DB

func addEntry(temperature float32, channel string, db *sql.DB) (error) {
	stmtIns, err := db.Prepare("INSERT INTO TEMP_ENTRY VALUES(0, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmtIns.Close()
	var datetime = time.Now()
	datetime.Format(time.RFC3339)
	_, err = stmtIns.Exec(channel, datetime, temperature)
	if err != nil {
		return err
	}
	return nil
}

func readEntries(maxEntry int, channel string, db *sql.DB) ([]TempEntry, error) {
	stmtOut, err := db.Prepare("SELECT CREATION_TIME, TEMPERATURE FROM TEMP_ENTRY WHERE CHANNEL = ? ORDER BY ID DESC LIMIT ?")
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(channel, maxEntry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	temperature := []TempEntry{}
	for rows.Next() {
		var actualEntry TempEntry
		err = rows.Scan(&actualEntry.Date, &actualEntry.Temp)
		if err != nil {
			return nil, err
		}
		actualEntry.Channel = channel
		temperature = append(temperature, actualEntry)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return temperature, nil
}


func main() {
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(xxx.xxx.xxx.xxx:XXXX)/dbName?parseTime=true")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	tempEntry, _ := readEntries(5, "BIG_TANK", db)
	fmt.Println(tempEntry)

	http.HandleFunc("/", handler)
	http.HandleFunc("/v1/temp", handleRequest)
	err = http.ListenAndServe(":12004", nil)
	if err != nil {
		panic(err.Error())
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleHTTPRead(w, r)
	case http.MethodPost:
		handleHTTPWrite(w, r)
	default:
		logAndHandleError(w, "read temperature usage: GET temp?channel=xxxxx&maxEntry=xx\nadd temperature entry usage: POST temp?channel=xxxxx&temp=xx.xx")
	}
}

func handleHTTPRead(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	if (len(channel) == 0) {
		logAndHandleError(w, "usage: GET /v1/temp?channel=xxxxx&maxEntry=xx")
		return
	}
	maxEntry := r.URL.Query().Get("maxEntry")
	if len(maxEntry) == 0 {
		maxEntry = "10";
	}
	maxEntryAsInt, err := strconv.Atoi(maxEntry)
	if err != nil {
		logAndHandleError(w, "Error while parsing max entry for channel %s, received max entry: %s\n Error: %s", channel, maxEntry, err.Error())
		return
	}
	tempEntry, err := readEntries(maxEntryAsInt, channel, db)
	if err != nil {
		logAndHandleError(w, "Error while reading entry for channel %s, max entry %d\n Error: %s", channel, maxEntryAsInt, err.Error())
		return
	}
	jsonOut, err := json.Marshal(tempEntry)
	if err != nil {
		logAndHandleError(w, "Error while marshaling entry for channel %s, max entry %d: %s\n Error: %s", tempEntry, channel, maxEntryAsInt, err.Error())
		return
	}
	fmt.Printf("Requested entry for channel %s with a maximum of %d entry\n", channel, maxEntryAsInt)
	fmt.Print("Going to return: ")
	fmt.Println(tempEntry)
	w.Write(jsonOut)
}

func handleHTTPWrite(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	temperature := r.URL.Query().Get("temp")
	if len(channel) == 0 || len(temperature) == 0 {
		logAndHandleError(w, "Invalid insertion request: channel %s temperature %s\nusage: POST temp?channel=xxxxx&temp=xx.xx", channel, temperature)
		return
	}
	convertedTemp, err := strconv.ParseFloat(temperature, 32)
	if err != nil {
		logAndHandleError(w, "Error while parsing given temperature: %s", err.Error())
		return
	}
	err = addEntry(float32(convertedTemp), channel, db)
	if err != nil {
		logAndHandleError(w, "Problem while inserting entry for channel %s with temperature %f: %s", channel, convertedTemp, err.Error())
		return
	}
	successInfo := fmt.Sprintf("Added entry for channel %s with temperature %f", channel, convertedTemp)
	fmt.Println(successInfo)
	fmt.Fprint(w, successInfo)
}

func logAndHandleError(w http.ResponseWriter, format string, a ...interface{}) {
	errorInfo := fmt.Sprintf(format, a...)
	fmt.Println(errorInfo)
	http.Error(w, errorInfo, http.StatusInternalServerError)
}

func handler(w http.ResponseWriter, r *http.Request) {
	logAndHandleError(w, "read temperature entry usage: GET temp?channel=xxxxx&maxEntry=xx\nadd  temperature entry usage: POST temp?channel=xxxxx&temp=xx.xx")
}