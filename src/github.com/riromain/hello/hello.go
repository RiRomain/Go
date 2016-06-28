package main

import ("fmt"
	"time"
	"database/sql"
)
import (
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
	"encoding/json"
)

type TempEntry struct {
	Channel string
	Date    time.Time
	Temp    float32
}

func addEntry(temperature float32, channel string, db *sql.DB) {
	stmtIns, err := db.Prepare("INSERT INTO TEMP_ENTRY VALUES(0, ?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer stmtIns.Close()
	var datetime = time.Now()
	datetime.Format(time.RFC3339)
	_, err = stmtIns.Exec(channel, datetime, temperature)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
}

func readEntries(maxEntry int, channel string, db *sql.DB) ([]TempEntry){
	stmtOut, err := db.Prepare("SELECT CREATION_TIME, TEMPERATURE FROM TEMP_ENTRY WHERE CHANNEL = ? ORDER BY ID DESC LIMIT ?")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(channel, maxEntry)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer rows.Close()
	temperature := []TempEntry{}
	for rows.Next() {
		var actualEntry TempEntry
		err = rows.Scan(&actualEntry.Date, &actualEntry.Temp)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		actualEntry.Channel = channel
		temperature = append(temperature, actualEntry)
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
	}
	return temperature
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(xxx.xxx.xxx.xxx:XXXX)/dbName?parseTime=true")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	addEntry(23.2, "BIG_TANK", db)
	tempEntry := readEntries(5, "BIG_TANK", db)
	fmt.Println(tempEntry)

	http.HandleFunc("/", handler)
	http.HandleFunc("/getTemp", handleHTTPRead)
	http.ListenAndServe(":8080", nil)
}

func handleHTTPRead(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	maxEntry := r.URL.Query().Get("maxEntry")
	if len(channel) == 0 {
		maxEntry = "10";
	}
	maxEntryAsInt , _ := strconv.Atoi(maxEntry)
	tempEntry := readEntries(maxEntryAsInt, channel, db)
	fmt.Println(tempEntry)
	jsonOut, err := json.Marshal(tempEntry)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(jsonOut)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Response %s", r.URL.Path[1:])
}