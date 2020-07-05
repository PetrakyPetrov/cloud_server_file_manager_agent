package main

import (
	"log"
	"net/http"
	"os"
	"strings"
)

func apiResponse(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	allowedEndpoints := map[string]bool{
		"files": true,
	}

	apiEndPointURL := r.URL.Path
	apiEndPoints := strings.Split(apiEndPointURL, "/")
	endPoint := apiEndPoints[1]

	if !allowedEndpoints[endPoint] {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Method not found"}`))
	} else {
		switch r.Method {
		case "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "GET file"}`))
		case "POST":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"message": "POST file"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Can't find file"}`))
		}
	}
}

func main() {

	file, err := os.OpenFile("/var/log/cloud_server_file_manager_agent.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
	log.Println("Agent started")

	http.HandleFunc("/", apiResponse)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
