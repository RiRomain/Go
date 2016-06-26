package main

import ("fmt"
	"time"
	"container/list"
	"database/sql"
)
import _ "github.com/go-sql-driver/mysql"

type tempRecord struct {
	tempEntry list.List
}

type TempEntry struct {
	temp float32
	date time.Time
	channel string
}

func main() {

	db, err := sql.Open("mysql", "temperature:local/temperature")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	stmtIns, err := db.Prepare("INSERT INTO TEMP_ENTRY VALUES(0, ?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}

	defer stmtIns.Close()

	stmtOut, err := db.Prepare("SELECT TEMPERATURE FROM TEMP_ENTRY ORDER BY ID DESC LIMIT 1")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()
	var datetime = time.Now()
	datetime.Format(time.RFC3339)
	_, err = stmtIns.Exec("BIG_THANK", datetime, 23.2) // Insert tuples (i, i^2)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	var outTemp float32
	err = stmtOut.QueryRow().Scan(&outTemp) // WHERE number = 13
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	fmt.Println("The square number of 13 is: ", outTemp)

    	fmt.Println("hello, world\n")
}

