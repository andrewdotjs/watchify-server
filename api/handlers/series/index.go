package series

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/api/types"
	"github.com/andrewdotjs/watchify-server/api/utilities"
	"github.com/google/uuid"
)

// Gets and returns an array of series stored in the database.
//
// # Specifications:
//   - Method      : GET
//   - Endpoint    : /series/{id}
//   - Auth?       : False
//
// # HTTP request query parameters:
//   - series_id   : OPTIONAL. UUID of the series.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - message     : If error, Message detailing the error.
//   - data        : Series contents, each returning id, episode count, title, description.
func ReadSeries(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var series types.Series
	id := r.PathValue("id")

	if id == "" {
		var seriesArray []types.Series

		rows, err := database.Query(`
		SELECT id, title, description, episodes
		FROM series
		LIMIT 30
    `)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				defer database.Close()
				log.Fatalf("ERR : %v", err)
			}

			responses.Status{
				Status: 200,
				Data:   nil,
			}.ToClient(w)
			return
		}

		defer rows.Close()

		for rows.Next() {
			var series types.Series

			if err := rows.Scan(
				&series.Id,
				&series.Title,
				&series.Description,
				&series.EpisodeCount,
			); err != nil {
				defer database.Close()
				log.Fatalf("ERR : %v", err)
			}

			seriesArray = append(seriesArray, series)
		}

		responses.Status{
			Status: 200,
			Data:   seriesArray,
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(`
		SELECT id, title, description, episodes
		FROM series
		WHERE id=?
		`,
		id,
	).Scan(
		&series.Id,
		&series.Title,
		&series.Description,
		&series.EpisodeCount,
	); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			Status:  400,
			Message: "No series found with given id.",
		}.ToClient(w)
		return
	}

	responses.Status{
		Status: 200,
		Data:   series,
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
//   - message     : If error, Message detailing the error.
//   - data        : Series id, title
func CreateSeries(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	if err := r.ParseMultipartForm(1 << 40); err != nil { // Error handling if form data exceeds 1TB
		log.Printf("%v", err)
		responses.Error{
			Type:   "null",
			Title:  "Incomplete request",
			Status: 400,
			Detail: "The upload form exceeded 1TB",
		}.ToClient(w)
		return
	}

	uploadDirectory := path.Join(*appDirectory, "storage", "videos")
	currentTime := time.Now().Format("01-02-2006 15:04:05")
	uploadedVideos := r.MultipartForm.File["videos"]

	uploadedCover := r.MultipartForm.File["cover"]
	if len(uploadedCover) == 0 {
		log.Print("Received no cover in request.")
		responses.Error{
			Type:   "null",
			Title:  "Bad request",
			Status: 400,
			Detail: "No uploaded cover present in form",
		}.ToClient(w)
		return
	}
	//if err != nil {
	//	var response responses.Error
	//
	//	switch {
	//	case errors.Is(err, http.ErrMissingFile):
	//		log.Print("Received no cover in request.")
	//		response = responses.Error{
	//			Type:   "null",
	//			Title:  "Bad request",
	//			Status: 400,
	//			Detail: "No uploaded cover present in form",
	//		}
	//	default:
	//		log.Print(err)
	//		response = responses.Error{
	//			Type:   "null",
	//			Title:  "Unaccounted Error",
	//			Status: 500,
	//			Detail: fmt.Sprintf("%v", err),
	//		}
	//	}
	//
	//	response.ToClient(w)
	//	return
	//}

	if len(uploadedVideos) == 0 {
		log.Print("Received no videos in request.")
		responses.Error{
			Type:   "null",
			Title:  "Bad request",
			Status: 400,
			Detail: "No uploaded videos present in form",
		}.ToClient(w)
		return
	}

	series := types.Series{
		Id:           uuid.New().String(),
		Title:        r.FormValue("title"),
		Description:  r.FormValue("description"),
		UploadDate:   currentTime,
		LastModified: currentTime,
	}

	// Handle upload for every file that was passed in the form.
	for index, uploadedFile := range uploadedVideos {
		video := types.Video{SeriesId: series.Id}
		utilities.HandleVideoUpload(uploadedFile, &video, database, &uploadDirectory)
		series.EpisodeCount = index + 1
	}

	uploadDirectory = path.Join(*appDirectory, "storage", "covers")
	cover := types.Cover{SeriesId: series.Id}

	fmt.Print(uploadedCover)

	utilities.HandleCoverUpload(
		uploadedCover[0],
		&cover,
		database,
		&uploadDirectory,
	)

	if _, err := database.Exec(`
   	INSERT INTO series
   	VALUES (?, ?, ?, ?, ?, ?)
    `,
		series.Id,
		series.Title,
		series.Description,
		series.EpisodeCount,
		series.UploadDate,
		series.LastModified,
	); err != nil {
		log.Fatalf("ERR : %v", err)
	}

	responses.Status{
		Status: 201,
		Data:   series,
	}.ToClient(w)
}

// Deletes a series, its episodes, and its cover from the database and storage folders.
//
// # Specifications:
//   - Method      : DELETE
//   - Endpoint    : /series/{id}
//   - Auth?       : False
//
// # HTTP request path parameters:
//   - id          : REQUIRED. UUID of the series.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - message     : If error, Message detailing the error.
func DeleteSeries(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var coverFileName string
	id := r.PathValue("id")

	// Stage 1, find all videos that are in the to-be-deleted series and delete them.
	rows, err := database.Query(`
  	SELECT file_name
  	FROM videos
  	WHERE series_id=?
    `,
		id,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var videoFileName string

		if err := rows.Scan(&videoFileName); err != nil {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		if err := os.Remove(
			path.Join(*appDirectory, "storage", "videos", videoFileName),
		); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				defer database.Close()
				log.Fatalf("ERR : %v", err)
			}

			responses.Status{
				Status:  500,
				Message: "Error removing video file.",
			}.ToClient(w)
			return
		}
	}

	if _, err := database.Exec(`
	  DELETE FROM videos
		WHERE series_id=?
		`,
		id,
	); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	// Stage 2, delete the cover from the database.
	if _, err := database.Exec(`
	  DELETE FROM covers
	  WHERE series_id=?
  	`,
		id,
	); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	// Stage 3, delete the cover from the storage.
	if err := os.Remove(
		path.Join(*appDirectory, "storage", "covers", coverFileName),
	); err != nil && !errors.Is(err, os.ErrNotExist) {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	// Stage 4, delete the series itself from the database.
	if _, err := database.Exec(`
  	DELETE FROM series
  	WHERE id=?
  	`,
		id,
	); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	responses.Status{
		Status: 200,
	}.ToClient(w)
}
