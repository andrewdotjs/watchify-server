package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
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
// Specifications:
//   - Method        : GET
//   - Endpoint      : api/v1/series
//   - Authorization : False
//
// HTTP request query parameters:
//   - id            : Optional. Will match provided id with series with same id, fails if not exists.
//   - limit         : Optional. Overrides the default limit 30 for returned rows.
//   - search        : Optional. Does a hard search for a specific series.
//
// HTTP response JSON contents:
//   - status_code   : HTTP status code.
//   - error_code    : If error, gives in-house error code for debugging. (not implemented yet)
//   - message       : If error, Message detailing the error.
//   - data          : Series contents, each returning id, episode count, title, description.
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

// Uploads a series, its episodes, and its cover to the database and stores them within the
// storage folder.
//
// Specifications:
//   - Method        : POST
//   - Endpoint      : api/v1/series/upload
//   - Authorization : False
//
// HTTP request multipart form:
//   - video-files   : Required. Uploaded video files.
//   - name          : Required. Name of the soon-to-be uploaded series.
//   - description   : Required. Description of the soon-to-be series.
//
// HTTP response JSON contents:
//   - status_code   : HTTP status code.
//   - error_code    : If error, gives in-house error code for debugging. (not implemented yet)
//   - message       : If error, Message detailing the error.
//   - data          : Series id, title
func PostSeriesHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var series types.Series

	// Error handling if form data exceeds 1TB
	if err := r.ParseMultipartForm(1 << 40); err != nil {
		responses.Status{
			StatusCode: 400,
			Message:    "Did the file exceed 1TB?",
		}.ToClient(w)
		return
	}

	currentTime := time.Now().Format("01-02-2006 15:04:05")
	series.Id = uuid.New().String()
	series.Title = r.FormValue("series-title")
	series.Description = r.FormValue("series-description")
	series.UploadDate = currentTime
	series.LastModified = currentTime

	uploadedFiles := r.MultipartForm.File["videos"]
	uploadDirectory := path.Join(*appDirectory, "storage", "videos")

	for index, uploadedFile := range uploadedFiles {
		video := types.Video{SeriesId: series.Id}
		utilities.HandleVideoUpload(uploadedFile, &video, database, &uploadDirectory)
		series.Episodes = index + 1
	}

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
// Specifications:
//   - Method        : DELETE
//   - Endpoint      : api/v1/series/delete
//   - Authorization : False
//
// Possible query parameters:
//   - id            : required, deletes series, videos, and covers that match the id from both db and storage.
//
// HTTP response JSON contents:
//   - status_code   : HTTP status code.
//   - error_code    : If error, gives in-house error code for debugging. (not implemented yet)
//   - message       : If error, Message detailing the error.
func DeleteSeriesHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	id := mux.Vars(r)["id"]

	_, err := database.Exec("DELETE FROM series WHERE id=?;", id)
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	responses.Status{
		StatusCode: 200,
	}.ToClient(w)
}
