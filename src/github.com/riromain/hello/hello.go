package main

import ("fmt"
	"time"
	"container/list"
	"database/sql"
)
import (
	_ "github.com/go-sql-driver/mysql"
	"math/big"
)

type tempRecord struct {
	tempEntry list.List
}

type TempEntry struct {
	id big.Int
	channel string
	date time.Time
	temp float32
}

func addEntry(temperature float32, channel string, db sql.DB) {
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

func readEntries(maxEntry int, )

func main() {

	db, err := sql.Open("mysql", "user:password@tcp(xxx.xxx.xxx.xxx:XXXX)/dbName?parseTime=true")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	addEntry(23.2, "BIG_THANK", db)

	stmtOut, err := db.Prepare("SELECT TEMPERATURE FROM TEMP_ENTRY ORDER BY ID DESC LIMIT 1")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	var outTemp float32
	err = stmtOut.QueryRow().Scan(&outTemp) // WHERE number = 13
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	fmt.Println("The square number of 13 is: ", outTemp)

    	stmtOut, err = db.Prepare("SELECT CREATION_TIME, TEMPERATURE FROM TEMP_ENTRY WHERE CHANNEL = ? ORDER BY ID DESC LIMIT ?")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query("BIG_TANK", 10)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer rows.Close()

	var (
		creationTime time.Time
		temperature float32
	)

	for rows.Next() {
		err := rows.Scan(&creationTime, &temperature)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		fmt.Println(creationTime, temperature)
	}
	err = rows.Err()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
}

