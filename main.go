package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// CloudWorkDir system folder
const CloudWorkDir = "/cloud_storage"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func apiResponse(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	allowedEndpoints := map[string]bool{
		"files":  true,
		"folder": true,
	}

	apiEndPointURL := r.URL.Path

	apiEndPoint := strings.Split(apiEndPointURL, "/")
	accountID := apiEndPoint[2]
	endPoint := apiEndPoint[3]

	if _, err := strconv.Atoi(accountID); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Method not found"}`))
		return
	}

	workDir := CloudWorkDir + "/" + accountID + "/"

	if !allowedEndpoints[endPoint] {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Method not found"}`))
	} else {
		switch r.Method {
		case "GET":

			fileName := r.URL.Query().Get("name")
			if fileName == "" {
				//Get not set, send a 400 bad request
				http.Error(w, "Get 'file' not specified in url.", 400)
				return
			}

			//Check if file exists and open
			path := workDir + "/" + fileName
			Openfile, err := os.Open(path)
			defer Openfile.Close() //Close after function return
			if err != nil {
				//File not found, send 404
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "File not found"}`))
				return
			}

			FileHeader := make([]byte, 512)
			//Copy the headers into the FileHeader buffer
			Openfile.Read(FileHeader)
			//Get content type of file
			FileContentType := http.DetectContentType(FileHeader)

			//Get the file size
			FileStat, _ := Openfile.Stat()                     //Get info from file
			FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

			//Send the headers
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
			w.Header().Set("Content-Type", FileContentType)
			w.Header().Set("Content-Length", FileSize)

			//Send the file
			//We read 512 bytes from the file already, so we reset the offset back to 0
			Openfile.Seek(0, 0)
			io.Copy(w, Openfile)
			return

		case "POST":
			fmt.Println(workDir)
			_, err := os.Stat(workDir)
			if os.IsNotExist(err) {
				errDir := os.MkdirAll(workDir, 0755)
				if errDir != nil {
					log.Fatal(err)
				}
			}

			file, handler, err := r.FormFile("file")
			if err != nil {
				panic(err)
			}
			defer file.Close()
			body, err := ioutil.ReadAll(file)
			if err != nil {
				panic(err)
			}

			d1 := []byte(body)
			err = ioutil.WriteFile(workDir+"/"+handler.Filename, d1, 0644)
			check(err)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"message": "File uplouded successfully"}`))

		case "DELETE":

			fileName := r.URL.Query().Get("name")
			if fileName == "" {
				//Get not set, send a 400 bad request
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"message": "Get 'file' not specified in url."}`))
				return
			}

			//Check if file exists and open
			path := workDir + "/" + fileName
			if _, err := os.Stat(path); os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "File not found"}`))
				return
			}

			err := os.Remove(path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message": "Cannot delete file"}`))
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "File deleted successfully"}`))

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

	http.HandleFunc("/account/", apiResponse)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
