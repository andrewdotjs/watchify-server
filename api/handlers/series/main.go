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

	"github.com/andrewdotjs/watchify-server/api/database"
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
	var id string = r.PathValue("id")
	var series types.Series

	// Return all series if no ID.
	if id == "" {
		var seriesArray []types.Series

		rows, err := database.Query(
			`
			SELECT
				id, title, description, episodes
			FROM
				series
			LIMIT
				30
    	`,
		)

		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				responses.Status{
					Status: 200,
					Data:   nil,
				}.ToClient(w)
			default:
				responses.Error{
					Type:     "null",
					Title:    "Unknown Error",
					Status:   500,
					Detail:   fmt.Sprintf("%v", err),
					Instance: r.URL.Path,
				}.ToClient(w)
			}

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

			series.Episodes = map[string]any{
				"count":    series.EpisodeCount,
				"endpoint": ("/api/v1/series/" + series.Id + "/episodes"),
			}

			series.Cover = map[string]any{
				"exists":   true,
				"endpoint": ("/api/v1/series/" + series.Id + "/cover"),
			}

			series.Splash = map[string]any{
				"exists":   false,
				"endpoint": ("/api/v1/series/" + series.Id + "/splash"),
			}

			seriesArray = append(seriesArray, series)
		}

		responses.Status{
			Status: 200,
			Data:   seriesArray,
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(
		`
		SELECT
			id, title, description, episodes
		FROM
			series
		WHERE
			id=?
		`,
		id,
	).Scan(
		&series.Id,
		&series.Title,
		&series.Description,
		&series.EpisodeCount,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			responses.Error{
				Type:     "null",
				Title:    "Data not found",
				Status:   400,
				Detail:   "No series could be found with the given id.",
				Instance: r.URL.Path,
			}.ToClient(w)
		default:
			responses.Error{
				Type:     "null",
				Title:    "Unknown Error",
				Status:   500,
				Detail:   fmt.Sprintf("%v", err),
				Instance: r.URL.Path,
			}.ToClient(w)
		}

		return
	}

	// Assemble
	series.Episodes = map[string]any{
		"count":    series.EpisodeCount,
		"endpoint": ("/api/v1/series/" + series.Id + "/episodes"),
	}

	series.Cover = map[string]any{
		"exists":   true,
		"endpoint": ("/api/v1/series/" + series.Id + "/cover"),
	}

	series.Splash = map[string]any{
		"exists":   false,
		"endpoint": ("/api/v1/series/" + series.Id + "/splash"),
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

	if id == "" {
		responses.Error{
			Type:     "null",
			Title:    "Invalid Request",
			Status:   400,
			Detail:   "Id was not present in the URL.",
			Instance: r.URL.Path,
		}.ToClient(w)
		return
	}

	// Stage 1, find all videos that are in the to-be-deleted series and delete them.
	rows, err := database.Query(
		`
  	SELECT file_name
  	FROM videos
  	WHERE series_id=?
    `,
		id,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		var response responses.Error

		switch {
		default:
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
		}

		response.ToClient(w)
		return
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
			var response responses.Error

			switch {
			case errors.Is(err, os.ErrNotExist):
				response = responses.Error{
					Type:     "null",
					Title:    "File system and database out-of-sync.",
					Status:   500,
					Detail:   "Attempted to delete a non-existant video file that exists in the database.",
					Instance: r.URL.Path,
				}
			default:
				response = responses.Error{
					Type:     "null",
					Title:    "An unknown error has occurred.",
					Status:   500,
					Detail:   "Sorry, but this error hasn't been properly logged yet.",
					Instance: r.URL.Path,
				}
				log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
			}

			response.ToClient(w)
			return
		}
	}

	if _, err := database.Exec(
		`
	  DELETE FROM videos
		WHERE series_id=?
		`,
		id,
	); err != nil {
		var response responses.Error

		switch {
		default:
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
		}

		response.ToClient(w)
		return
	}

	// Stage 2, delete the cover from the database.
	if _, err := database.Exec(
		`
	  DELETE FROM covers
	  WHERE series_id=?
  	`,
		id,
	); err != nil {
		var response responses.Error

		switch {
		default:
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
		}

		response.ToClient(w)
		return
	}

	// Stage 3, delete the cover from the storage.
	if err := os.Remove(
		path.Join(*appDirectory, "storage", "covers", coverFileName),
	); err != nil && !errors.Is(err, os.ErrNotExist) {
		var response responses.Error

		switch {
		default:
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
		}

		response.ToClient(w)
		return
	}

	// Stage 4, delete the series itself from the database.
	if _, err := database.Exec(
		`
  	DELETE FROM series
  	WHERE id=?
  	`,
		id,
	); err != nil {
		var response responses.Error

		switch {
		default:
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
		}

		response.ToClient(w)
		return
	}

	responses.Status{
		Status: 200,
	}.ToClient(w)
}

func UpdateSeries(w http.ResponseWriter, r *http.Request, db *sql.DB, appDirectory *string) {
	id := r.PathValue("id")
	var episodeCount int

	if id == "" {
		responses.Error{
			Type:     "null",
			Title:    "Invalid Request",
			Status:   400,
			Detail:   "Id was not present in the URL.",
			Instance: r.URL.Path,
		}.ToClient(w)
		return
	}

	// update episode count
	if err := db.QueryRow(
		`
		SELECT
			COUNT(*) as count
		FROM
			videos
		WHERE
			series_id=?
		`,
		id,
	).Scan(&episodeCount); err != nil {
		var response responses.Error

		switch {
		default:
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
		}

		response.ToClient(w)
		return
	}

	updatedSeries := types.Series{
		Id:           id,
		Title:        r.FormValue("title"),
		Description:  r.FormValue("description"),
		EpisodeCount: episodeCount,
		LastModified: time.Now().Format("01-02-2006 15:04:05"),
	}

	statement, values := database.UpdateQueryBuilder("series", updatedSeries)
	if _, err := db.Exec(statement, values...); err != nil {
		var response responses.Error

		switch {
		default:
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
		}

		response.ToClient(w)
	} else {
		responses.Status{
			Status: 200,
		}.ToClient(w)
	}
}
