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
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/types"
	"github.com/google/uuid"
)

func UploadVideoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uploadDir := "./storage/videos"

	err := r.ParseMultipartForm(1 << 30) // 1 GB limit
	if err != nil {
		response, _ := json.Marshal(types.Message{
			StatusCode: http.StatusBadRequest,
			Message:    "Could not parse form data. Is it beyond 1GB file limit? Is encoding set to multipart/form-data?",
		})

		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("video")
	if err != nil {
		response, _ := json.Marshal(types.Message{
			StatusCode: http.StatusBadRequest,
			Message:    "Unable to get file from form. Was fileName set to video?",
		})
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	defer file.Close()

	// Create a unique ID
	fileIdentifier := fmt.Sprint(uuid.New())
	splitFileName := strings.Split(handler.Filename, ".")
	fileName := fmt.Sprintf("%s.%s", fileIdentifier, splitFileName[1])

	// Create the upload directory if it doesn't exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		log.Printf("SYS : Could not find upload directory '%s'. Creating one.", uploadDir)
		os.Mkdir(uploadDir, os.ModePerm)
	}

	// Create the file in the upload directory
	uploadPath := filepath.Join(uploadDir, fileName)
	out, err := os.Create(uploadPath)

	if err != nil {
		response, _ := json.Marshal(types.Message{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to create file",
		})
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(response)
		return
	}

	defer out.Close()

	// Copy the file to the destination

	_, err = io.Copy(out, file)
	if err != nil {
		response, _ := json.Marshal(types.Message{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to copy file",
		})
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
	uploadDate := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.Exec(`INSERT INTO videos VALUES (?, ?, ?, ?, ?)`, fileIdentifier, "", 0, "Test Title", uploadDate)

	if err != nil {
		response, _ := json.Marshal(types.Message{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to insert data into database",
		})
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	response, _ := json.Marshal(types.Message{
		StatusCode: http.StatusOK,
		Message:    "File uploaded successfully and data inserted into the database.",
		Video: types.Video{
			Id:         fileIdentifier,
			UploadDate: uploadDate,
		},
	})

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
