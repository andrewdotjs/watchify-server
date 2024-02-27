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
	"github.com/google/uuid"
)

// Handles requests at the "/api/v1/covers" endpoint. Requires the
// c query parameter to be passed in for the client to recieve the
// image.
func GetCoverHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var coverIdentifier string = r.URL.Query().Get("c")
	var uploadDirectory string
	var coverFileName string

	if coverIdentifier == "" {
		responses.Status{
			StatusCode: 400,
			Message:    "c query param not passed in.",
		}.ToClient(w)
		return
	}

	if err := database.QueryRow("SELECT file_name FROM covers WHERE id=?;", coverIdentifier).Scan(&coverFileName); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 500,
			Message:    "No database match for provided id.",
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
		responses.Status{
			StatusCode: 500,
			Message:    "Could not find folder in file system. Folder has been created.",
		}.ToClient(w)
		return
	}

	// Read file at upload directory and
	buffer, err := os.ReadFile(path.Join(uploadDirectory, coverFileName))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 500,
			Message:    "Could not find file in file system.",
		}.ToClient(w)
		return
	}

	responses.File{
		FileBuffer: buffer,
	}.ToClient(w)
}

// Handles requests at the "/api/v1/covers/upload" endpoint. Requires
// a mulitpart form with the field names cover (jpg), series-id (string),
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
