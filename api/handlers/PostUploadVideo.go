package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/types"
	"github.com/andrewdotjs/watchify-server/utilities"
	"github.com/google/uuid"
)

func UploadVideoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var uploadDirectory string = "./storage/videos"
	var videoIdentifier string
	var seriesIdentifier string
	var episodeNumber int = 0
	var videoTitle string
	var fileName string
	var uploadDate string

	// Error handling if form data exceeds 1GB
	if err := r.ParseMultipartForm(1 << 30); err != nil {
		response := utilities.ErrorMessage(http.StatusBadRequest, "Did the file exceed 1GB?")
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("video")

	// Error handling if form data exceeds 1GB
	if err != nil {
		response := utilities.ErrorMessage(http.StatusBadRequest, "Unable to get file from form. Was fileName set to video?")
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	defer file.Close()

	seriesIdentifier = r.FormValue("series-identifier")
	videoTitle = r.FormValue("title")
	episodeNumber, err = strconv.Atoi(r.FormValue("episode-number"))

	if err != nil {
		log.Printf("ERR : %v. setting episode number to 0", err)
		episodeNumber = 0
	}

	// Create a unique ID
	videoIdentifier = fmt.Sprint(uuid.New())
	fileName = fmt.Sprintf("%s.%s", videoIdentifier, strings.Split(handler.Filename, ".")[1])

	// Create the upload directory if it doesn't exist
	if _, err := os.Stat(uploadDirectory); os.IsNotExist(err) {
		log.Printf("SYS : Could not find upload directory at '%s'. Creating one.", uploadDirectory)
		os.Mkdir(uploadDirectory, os.ModePerm)
	}

	// Create the file in the upload directory
	uploadPath := filepath.Join(uploadDirectory, fileName)
	out, err := os.Create(uploadPath)

	// Error handling if the file cannot be created
	if err != nil {
		response := utilities.ErrorMessage(http.StatusBadRequest, "Unable to create file.")
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(response)
		return
	}

	defer out.Close()

	// Copy the file to the destination and error handle.
	if _, err := io.Copy(out, file); err != nil {
		response := utilities.ErrorMessage(http.StatusBadRequest, "Unable to copy file.")
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	// Insert data into SQLite database
	database, err := sql.Open("sqlite3", "./db/videos.db")
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	defer database.Close()
	uploadDate = time.Now().Format("2006-01-02 15:04:05")

	if _, err = database.Exec(`INSERT INTO videos VALUES (?, ?, ?, ?, ?, ?)`, videoIdentifier, seriesIdentifier, episodeNumber, videoTitle, fileName, uploadDate); err != nil {
		response := utilities.ErrorMessage(http.StatusBadRequest, "Unable to insert data into the database.")
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	response, _ := json.Marshal(types.Message{
		StatusCode: http.StatusOK,
		Message:    "File uploaded successfully and data sent to database.",
		Video: types.Video{
			Id:         videoIdentifier,
			SeriesId:   seriesIdentifier,
			Episode:    episodeNumber,
			Title:      videoTitle,
			FileName:   fileName,
			UploadDate: uploadDate,
		},
	})

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
