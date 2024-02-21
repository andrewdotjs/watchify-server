package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/types"
	"github.com/google/uuid"
)

func UploadCoverHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	w.Header().Set("Content-Type", "application/json")
	var uploadDirectory string = "./storage/covers"
	var cover types.Cover

	// Error handling if form data exceeds 1MB
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			ErrorCode:  "40013",
			Message:    "Did the file exceed 1MB?",
		}.ToJSON())
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("cover")

	// Error handling if form data exceeds 1GB
	if err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			ErrorCode:  "4001341",
			Message:    "Unable to get file from form. Was fileName set to cover?",
		}.ToJSON())
		return
	}

	defer file.Close()

	cover.Id = fmt.Sprint(uuid.New())
	cover.SeriesId = r.FormValue("series-identifier")
	cover.FileName = fmt.Sprintf("%s.%s", cover.Id, strings.Split(handler.Filename, ".")[1])
	cover.UploadDate = time.Now().Format("2006-01-02 15:04:05")

	// Create the upload directory if it doesn't exist
	if _, err := os.Stat(uploadDirectory); os.IsNotExist(err) {
		log.Printf("SYS : Could not find upload directory at '%s'. Creating one.", uploadDirectory)
		os.Mkdir(uploadDirectory, os.ModePerm)
	}

	if _, err = database.Exec(
		`INSERT INTO covers VALUES (?, ?, ?, ?);`,
		cover.Id,
		cover.SeriesId,
		cover.FileName,
		cover.UploadDate,
	); err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			ErrorCode:  "40203",
			Message:    "Unable to insert data into the database.",
		}.ToJSON())
		log.Fatalf("ERR : %v", err)
		return
	}

	// Create the file in the upload directory
	uploadPath := filepath.Join(uploadDirectory, cover.FileName)
	out, err := os.Create(uploadPath)

	// Error handling if the file cannot be created
	if err != nil {
		w.WriteHeader(500)
		w.Write(responses.Error{
			StatusCode: 500,
			ErrorCode:  "5032",
			Message:    "Unable to create file.",
		}.ToJSON())
		return
	}

	defer out.Close()

	// Copy the file to the destination and error handle.
	if _, err := io.Copy(out, file); err != nil {
		w.WriteHeader(500)
		w.Write(responses.Error{
			StatusCode: 500,
			ErrorCode:  "500",
			Message:    "Unable to copy file.",
		}.ToJSON())
		return
	}

	w.WriteHeader(200)
	w.Write(responses.Cover{
		StatusCode: 200,
		Data:       cover,
	}.ToJSON())
}
