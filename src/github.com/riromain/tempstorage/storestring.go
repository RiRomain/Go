package ipstore

import (
	"net/http"
	"fmt"
	"strconv"
	"encoding/json"
)

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/v1/string", handleRequest)

}

var tempString string

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
	fmt.Println(tempString)
	w.Write(tempString)
}

func handleHTTPWrite(w http.ResponseWriter, r *http.Request) {
	stringReceived := r.URL.Query().Get("string")
	if len(stringReceived) == 0 {
		logAndHandleError(w, "Invalid insertion request: channel %s temperature %s\nusage: POST temp?channel=xxxxx&temp=xx.xx", channel, temperature)
		return
	}
	tempString = stringReceived
	successInfo := fmt.Sprintf("Stored new string %s", tempString)
	fmt.Println(successInfo)
	fmt.Fprint(w, successInfo)
}


func handler(w http.ResponseWriter, r *http.Request) {
	logAndHandleError(w, "real last string usage: GET v1/string\nstore new string usage: POST v1/string?string=xxx")
}

func logAndHandleError(w http.ResponseWriter, format string, a ...interface{}) {
	errorInfo := fmt.Sprintf(format, a...)
	fmt.Println(errorInfo)
	http.Error(w, errorInfo, http.StatusInternalServerError)
}