package main

import (
	"net/http"
	"fmt"
)

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/v1/string", handleRequest)
	err := http.ListenAndServe(":12004", nil)
	if err != nil {
		panic(err.Error())
	}
}

var tempString string

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleHTTPRead(w, r)
	case http.MethodPost:
		handleHTTPWrite(w, r)
	default:
		logAndHandleError(w, "read string usage: GET string\nadd string entry usage: POST string?string=xxxxx")
	}
}

func handleHTTPRead(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delivering stored string" + tempString)
	fmt.Fprintf(w, "%s", tempString)
}

func handleHTTPWrite(w http.ResponseWriter, r *http.Request) {
	stringReceived := r.URL.Query().Get("string")
	if len(stringReceived) == 0 {
		logAndHandleError(w, "Invalid insertion request:\nusage: POST string?string=xxxxx")
		return
	}
	tempString = stringReceived
	successInfo := fmt.Sprintf("Stored new string %s", tempString)
	fmt.Println(successInfo)
	fmt.Fprint(w, successInfo)
}


func handler(w http.ResponseWriter, r *http.Request) {
	logAndHandleError(w, "read last string usage: GET v1/string\nstore new string usage: POST v1/string?string=xxx")
}

func logAndHandleError(w http.ResponseWriter, format string, a ...interface{}) {
	errorInfo := fmt.Sprintf(format, a...)
	fmt.Println(errorInfo)
	http.Error(w, errorInfo, http.StatusInternalServerError)
}