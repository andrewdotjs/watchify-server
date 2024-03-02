package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/api/types"
	"github.com/andrewdotjs/watchify-server/api/utilities"
	"github.com/google/uuid"
)

// Returns the covers stored in the database and file-system.
//
// Specifications:
//   - Method        : GET
//   - Endpoint      : api/v1/covers
//   - Authorization : False
//
// HTTP request query parameters (Required that user queries with one of these):
//   - id            : Matches provided id with cover, falls back to placeholder if fails.
//   - s             : Matches provided series id with cover, falls back to placeholder if fails.
func GetCoverHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var coverIdentifier string = r.URL.Query().Get("c")
	var uploadDirectory string
	var coverFileName string

	if coverIdentifier == "" {
		responses.File{
			FileBuffer: utilities.PlaceholderCover(),
		}.ToClient(w)
		return
	}

	if err := database.QueryRow("SELECT file_name FROM covers WHERE id=?;", coverIdentifier).Scan(&coverFileName); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		// Log failure to file "no database match for provided id."
		responses.File{
			FileBuffer: utilities.PlaceholderCover(),
		}.ToClient(w)
		return
	}

	uploadDirectory = path.Join(*appDirectory, "storage", "covers")

	// Create the upload directory if it doesn't exist
	if _, err := os.Stat(uploadDirectory); err != nil {
		if !os.IsNotExist(err) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		log.Printf("SYS : Could not find upload directory at '%s'. Creating one.", uploadDirectory)
		os.Mkdir(uploadDirectory, os.ModePerm)

		// log to file "Could not find folder in file system. Folder has been created."
	}

	// Read file at upload directory and
	buffer, err := os.ReadFile(path.Join(uploadDirectory, coverFileName))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		// log to file "Could not find file in file system."

		responses.File{
			FileBuffer: utilities.PlaceholderCover(),
		}.ToClient(w)
		return
	}

	log.Print(buffer)
	responses.File{
		FileBuffer: buffer,
	}.ToClient(w)
}

// Allows the user to upload a file to the file system and store its information to
// the database.
//
// Specifications:
//   - Method        : POST
//   - Endpoint      : api/v1/covers
//   - Authorization : False
//
// HTTP request multipart form:
//   - series-id     : Id of the series that the user wants to attach the cover to.
//
// HTTP response JSON contents:
//   - status_code   : HTTP status code.
//   - error_code    : If error, gives in-house error code for debugging. (not implemented yet)
//   - message       : If error, message detailing the error.
//   - data          : id, series_id
func PostCoverHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var uploadDirectory string
	var uploadPath string
	var cover types.Cover
	var id string

	// Error handling if form data exceeds 1MB
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		responses.Status{
			StatusCode: 400,
			Message:    "Did the file exceed 1MB?",
		}.ToClient(w)
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("cover")

	// Error handling if form data exceeds 1GB
	if err != nil {
		responses.Status{
			StatusCode: 400,
			Message:    "Unable to get file from form. Was fileName set to cover?",
		}.ToClient(w)
		return
	}

	defer file.Close()

	id = uuid.New().String()
	cover = types.Cover{
		Id:         id,
		SeriesId:   r.FormValue("series-id"),
		FileName:   fmt.Sprintf("%s.%s", id, strings.Split(handler.Filename, ".")[1]),
		UploadDate: time.Now().Format("2006-01-02 15:04:05"),
	}

	uploadDirectory = path.Join(*appDirectory, "storage", "covers")

	// Create the upload directory if it doesn't exist
	if _, err := os.Stat(uploadDirectory); err != nil {
		if !os.IsNotExist(err) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		log.Printf("SYS : Could not find upload directory at '%s'. Creating one.", uploadDirectory)
		os.Mkdir(uploadDirectory, os.ModePerm)
	}

	_, err = database.Exec(`INSERT INTO covers VALUES (?, ?, ?, ?);`, cover.Id, cover.SeriesId, cover.FileName, cover.UploadDate)
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	uploadPath = filepath.Join(uploadDirectory, cover.FileName)
	out, err := os.Create(uploadPath)
	if err != nil {
		responses.Status{
			StatusCode: 500,
			Message:    "Unable to create file.",
		}.ToClient(w)
		return
	}

	defer out.Close()

	// Copy the file to the destination and error handle.
	if _, err := io.Copy(out, file); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	responses.Status{
		StatusCode: 200,
		Data:       cover,
	}.ToClient(w)
}

// Deletes a cover from the database and file system.
//
// Specifications:
//   - Method        : DELETE
//   - Endpoint      : api/v1/cover/delete
//   - Authorization : False
//
// HTTP request query parameters:
//   - id            :
//
// HTTP response JSON contents:
//   - status_code   : HTTP status code.
//   - error_code    : If error, gives in-house error code for debugging. (not implemented yet)
//   - message       : If error, message detailing the error.
func DeleteCoverHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var coverIdentifer string = r.URL.Query().Get("c")
	var fileName string

	if coverIdentifer == "" {
		responses.Status{
			StatusCode: 400,
			Message:    "c query param was not passed in.",
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(`SELECT file_name FROM covers WHERE id=?`, coverIdentifer).Scan(&fileName); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 400,
			Message:    "Unable to find cover identifier from c.",
		}.ToClient(w)
		return
	}

	if err := os.Remove(path.Join("./storage/covers", fileName)); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 400,
			Message:    "Could not delete cover from storage",
		}.ToClient(w)
		return
	}

	if _, err := database.Exec(`DELETE FROM videos WHERE id=?`, coverIdentifer); err != nil {
		responses.Status{
			StatusCode: 400,
			Message:    "Could not delete cover information from database.",
		}.ToClient(w)
		return
	}

	responses.Status{
		StatusCode: 200,
	}.ToClient(w)
}
