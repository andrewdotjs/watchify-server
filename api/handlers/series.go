package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/api/types"
	"github.com/andrewdotjs/watchify-server/api/utilities"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Gets and returns an array of series stored in the database.
//
// # Specifications:
//   - Method      : GET
//   - Endpoint    : /series/{id}
//   - Auth?       : False
//
// # HTTP request path parameters:
//   - id          : REQUIRED. Will match provided id with series with same id, fails if not exists.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - error_code  : If error, gives in-house error code for debugging. (not implemented yet)
//   - message     : If error, Message detailing the error.
//   - data        : Series contents, each returning id, episode count, title, description.
func GetSeriesHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var series types.Series
	id := mux.Vars(r)["id"]

	err := database.QueryRow("SELECT id, title, description, episodes FROM series WHERE id=?", id).Scan(&series.Id, &series.Title, &series.Description, &series.Episodes)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 200,
			Data:       nil,
		}.ToClient(w)
		return
	}

	responses.Status{
		StatusCode: 200,
		Data:       series,
	}.ToClient(w)
}

func GetAllSeriesHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var seriesArray []types.Series

	rows, err := database.Query(`SELECT id, title, description, episodes FROM series LIMIT 30`)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 200,
			Data:       nil,
		}.ToClient(w)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var series types.Series

		if err := rows.Scan(&series.Id, &series.Title, &series.Description, &series.Episodes); err != nil {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		seriesArray = append(seriesArray, series)
	}

	responses.Status{
		StatusCode: 200,
		Data:       seriesArray,
	}.ToClient(w)
	return
}

// Returns the covers stored in the database and file-system. If none are present,
// return a placeholder cover.
//
// # Specifications:
//   - Method   : GET
//   - Endpoint : /series/cover/{id}
//   - Auth?    : False
//
// # HTTP request path parameters (Required that user queries with one of these):
//   - id       : REQUIRED. Series id.
func GetSeriesCoverHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var coverFileName string

	id, ok := mux.Vars(r)["id"]
	if !ok {
		responses.File{
			FileBuffer: utilities.PlaceholderCover(),
		}.ToClient(w)
		return
	}

	if err := database.QueryRow("SELECT file_name FROM covers WHERE series_id=?;", id).Scan(&coverFileName); err != nil {
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

	uploadDirectory := path.Join(*appDirectory, "storage", "covers")

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

	responses.File{
		FileBuffer: buffer,
	}.ToClient(w)
}

// Uploads a series, its episodes, and its cover to the database and stores them within the
// storage folder.
//
// # Specifications:
//   - Method      : POST
//   - Endpoint    : /series
//   - Auth?       : False
//
// # HTTP request multipart form:
//   - video-files : REQUIRED. Uploaded video files.
//   - name        : REQUIRED. Name of the soon-to-be uploaded series.
//   - description : REQUIRED. Description of the series.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - error_code  : If error, gives in-house error code for debugging. (not implemented yet)
//   - message     : If error, Message detailing the error.
//   - data        : Series id, title
func PostSeriesHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	// Error handling if form data exceeds 1TB
	if err := r.ParseMultipartForm(1 << 40); err != nil {
		responses.Status{
			StatusCode: 400,
			Message:    "Did the file exceed 1TB?",
		}.ToClient(w)
		return
	}

	currentTime := time.Now().Format("01-02-2006 15:04:05")
	uploadedVideos := r.MultipartForm.File["videos"]
	uploadedCover := r.MultipartForm.File["cover"]

	series := types.Series{
		Id:           uuid.New().String(),
		Title:        r.FormValue("title"),
		Description:  r.FormValue("description"),
		UploadDate:   currentTime,
		LastModified: currentTime,
	}

	uploadDirectory := path.Join(*appDirectory, "storage", "videos")
	for index, uploadedFile := range uploadedVideos {
		video := types.Video{SeriesId: series.Id}
		utilities.HandleVideoUpload(uploadedFile, &video, database, &uploadDirectory)
		series.Episodes = index + 1
	}

	uploadDirectory = path.Join(*appDirectory, "storage", "covers")
	cover := types.Cover{SeriesId: series.Id}
	utilities.HandleCoverUpload(uploadedCover[0], &cover, database, &uploadDirectory)

	_, err := database.Exec(`
	INSERT INTO series
	VALUES (?, ?, ?, ?, ?, ?);
	`, series.Id, series.Title, series.Description, series.Episodes, series.UploadDate, series.LastModified)
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	responses.Status{
		StatusCode: 201,
		Data:       series,
	}.ToClient(w)
}

// Deletes a series, its episodes, and its cover from the database and storage folders.
//
// # Specifications:
//   - Method      : DELETE
//   - Endpoint    : /series/{id}
//   - Auth?       : False
//
// # Possible path parameters:
//   - id          : REQUIRED. Series id.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - error_code  : If error, gives in-house error code for debugging. (not implemented yet)
//   - message     : If error, Message detailing the error.
func DeleteSeriesHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var coverFileName string
	id := mux.Vars(r)["id"]

	// Stage 1, find all videos that are in to-be-deleted series and delete them.
	rows, err := database.Query("SELECT file_name FROM videos WHERE series_id=?", id)
	defer rows.Close()

	for rows.Next() {
		var videoFileName string
		rows.Scan(&videoFileName)

		if err := os.Remove(path.Join(*appDirectory, "storage", "videos", videoFileName)); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				defer database.Close()
				log.Fatalf("ERR : %v", err)
			}

			responses.Status{
				StatusCode: 500,
				Message:    "Error removing video file.",
			}.ToClient(w)
			return
		}
	}

	_, err = database.Exec("DELETE FROM videos WHERE series_id=?;", id)
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	// Stage 2, delete the cover for the series.
	err = database.QueryRow("SELECT file_name FROM covers WHERE series_id=?", id).Scan(&coverFileName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	} else if err == nil {
		if _, err = database.Exec("DELETE FROM covers WHERE series_id=?;", id); err != nil {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		if err := os.Remove(path.Join(*appDirectory, "storage", "covers", coverFileName)); err != nil && !errors.Is(err, os.ErrNotExist) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}
	}

	// Stage 3, delete the series itself.
	_, err = database.Exec("DELETE FROM series WHERE id=?;", id)
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	responses.Status{
		StatusCode: 200,
	}.ToClient(w)
}
