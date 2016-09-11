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

type GoogleChart struct {
	Cols[] Cols `json:"cols"`
	Rows[] Rows `json:"rows"`
}

type Cols struct {
	Id string `json:"id"`
	Label string `json:"label"`
	Type string `json:"type"`
}
type Rows struct {
	C[] RowEntry `json:"c"`
}
type RowEntry struct{
	V string `json:"v"`
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

func convertToGoogleChart(tempEntry[] TempEntry)(GoogleChart) {


	googleChart := GoogleChart{}
	googleChart.Cols = []Cols{}
	googleChart.Rows = []Rows{}
	col1 := Cols {
		Id:"",
		Label:"date",
		Type:"string",
	}
	col2 := Cols {
		Id:"",
		Label:"°C",
		Type:"number",
	}
	/*col3 := Cols {
		Id:"",
		Label:"channel",
		Type:"string",
	}*/
	googleChart.Cols = append(googleChart.Cols, col1)
	googleChart.Cols = append(googleChart.Cols, col2)
	//googleChart.Cols = append(googleChart.Cols, col3)
	fmt.Println(len(googleChart.Cols))
	for _,element := range tempEntry {
		var row Rows
		row.C = []RowEntry{}
		rowEntry1 := RowEntry{}

		rowEntry1.V = element.Date.Format("02/01/06 15:04")

		row.C = append(row.C, rowEntry1)
		rowEntry2 := RowEntry{}
		rowEntry2.V = strconv.FormatFloat(float64(element.Temp), 'f', 6, 32)
		row.C = append(row.C, rowEntry2)
		googleChart.Rows = append(googleChart.Rows, row)
	}
	googleChart.Rows = reverse(googleChart.Rows)
	return googleChart
}

func reverse(numbers []Rows) []Rows {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
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
		utc, err := time.LoadLocation("Europe/Berlin")
		if err != nil {
			fmt.Println("err: ", err.Error())
			return nil, err
		}
		actualEntry.Date = actualEntry.Date.In(utc)
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
	googleChart := convertToGoogleChart(tempEntry)
	fmt.Println(googleChart)
	jsonOut, err := json.Marshal(googleChart)
	fmt.Println(jsonOut)


	http.HandleFunc("/", handler)
	http.HandleFunc("/v1/temp", handleRequest)
	err = http.ListenAndServe(":12004", nil)
	if err != nil {
		panic(err.Error())
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, x-requested-with")
	if r.Method == "OPTIONS" {
		return
	}
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
	format := r.URL.Query().Get("format")
	maxEntryAsInt, err := strconv.Atoi(maxEntry)
	if err != nil {
		logAndHandleError(w, "Error while parsing max entry for channel %s, received max entry: %s\n Error: %s", channel, maxEntry, err.Error())
		return
	}
	tempEntry, err := readEntries(maxEntryAsInt, channel, db)
	gChartEntry := convertToGoogleChart(tempEntry)
	if err != nil {
		logAndHandleError(w, "Error while reading entry for channel %s, max entry %d\n Error: %s", channel, maxEntryAsInt, err.Error())
		return
	}
	var jsonOut[] byte
	if len(format) == 0 || "gChart" != format {
		jsonOut, err = json.Marshal(tempEntry)
	} else {
		jsonOut, err = json.Marshal(gChartEntry)
		fmt.Print("Selecting gChart format")
	}
	if err != nil {
		logAndHandleError(w, "Error while marshaling entry for channel %s, max entry %d: %s\n Error: %s", tempEntry, channel, maxEntryAsInt, err.Error())
		return
	}
	fmt.Printf("Requested entry for channel %s with a maximum of %d entry\n", channel, maxEntryAsInt)
	fmt.Print("Going to return: ")
	fmt.Println(tempEntry)
	fmt.Println(jsonOut)
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, successInfo)
}

func logAndHandleError(w http.ResponseWriter, format string, a ...interface{}) {
	errorInfo := fmt.Sprintf(format, a...)
	fmt.Println(errorInfo)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.Error(w, errorInfo, http.StatusInternalServerError)
}

func handler(w http.ResponseWriter, r *http.Request) {
	logAndHandleError(w, "read temperature entry usage: GET temp?channel=xxxxx&maxEntry=xx\nadd  temperature entry usage: POST temp?channel=xxxxx&temp=xx.xx")
}