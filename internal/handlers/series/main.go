package series

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/internal/functions"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
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
	var orderedBy string = r.URL.Query().Get("orderedBy")
	var orderedByQuery string
	var series types.Series

	// Return all series if no ID.
	if id == "" {
		var seriesArray []types.Series

		if orderedBy != "" {
			switch orderedBy {
			case "upload_date":
				orderedByQuery = "ORDER BY upload_date DESC"
			default:
				orderedByQuery = ""
			}
		}

		rows, err := database.Query(
			fmt.Sprintf(`
					SELECT
						id, title, description, episode_count, upload_date, last_modified
					FROM
						series
					%v
					LIMIT
						15
				`,
				orderedByQuery,
			),
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
				&series.UploadDate,
				&series.LastModified,
			); err != nil {
				defer database.Close()
				log.Fatalf("ERR : %v", err)
			}

			series.Episodes = map[string]any{
				"count": series.EpisodeCount,
				"url":   ("/api/v1/series/" + series.Id + "/episodes"),
			}

			series.Cover = map[string]any{
				"exists": true,
				"url":    ("/api/v1/series/" + series.Id + "/cover"),
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
			id, title, description, episode_count, upload_date, last_modified
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
		&series.UploadDate,
		&series.LastModified,
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
		"count": series.EpisodeCount,
		"url":   ("/api/v1/series/" + series.Id + "/episodes"),
	}

	series.Cover = map[string]any{
		"exists": true,
		"url":    ("/api/v1/series/" + series.Id + "/cover"),
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
		video := types.Episode{SeriesId: series.Id}
		functions.SeriesEpisode(uploadedFile, &video, database, &uploadDirectory)
		series.EpisodeCount = index + 1
	}

	uploadDirectory = path.Join(*appDirectory, "storage", "covers")
	cover := types.SeriesCover{SeriesId: series.Id}

	functions.SeriesCover(
		uploadedCover[0],
		&cover,
		database,
		&uploadDirectory,
	)

	if _, err := database.Exec(`
   	INSERT INTO
			series
   	VALUES
			(?, ?, ?, ?, ?, ?, ?)
    `,
		series.Id,
		series.Title,
		series.Description,
		series.EpisodeCount,
		series.Hidden,
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
	var videoStorageDirectory string = path.Join(*appDirectory, "storage", "videos")
	var id string = r.PathValue("id")
	var coverFileName string

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
  	SELECT
			file_name
  	FROM
			series_episodes
  	WHERE
			series_id=?
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

		if err := os.Remove(path.Join(videoStorageDirectory, videoFileName)); err != nil {
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

	if _, err := database.Exec(`
			DELETE FROM
				series_episodes
			WHERE
				series_id=?
		`,
		id,
	); err != nil {
		var response responses.Error

		switch {
		default:
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
		}

		response.ToClient(w)
		return
	}

	// Stage 2, delete the cover from the database.
	if _, err := database.Exec(
		`
	  DELETE FROM
			series_covers
	  WHERE
			series_id=?
  	`,
		id,
	); err != nil {
		var response responses.Error

		switch {
		default:
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
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
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
		}

		response.ToClient(w)
		return
	}

	// Stage 4, delete the series itself from the database.
	if _, err := database.Exec(
		`
  	DELETE FROM
			series
  	WHERE
			id=?
  	`,
		id,
	); err != nil {
		var response responses.Error

		switch {
		default:
			log.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
			response = responses.Error{
				Type:     "null",
				Title:    "An unknown error has occurred.",
				Status:   500,
				Detail:   "Sorry, but this error hasn't been properly logged yet.",
				Instance: r.URL.Path,
			}
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
	if err := db.QueryRow(`
			SELECT
				COUNT(*) as count
			FROM
				series_episodes
			WHERE
				series_id = ?
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

	description := r.FormValue("description")

	if len(description) > 1000 {
		responses.Error{
			Type:     "null",
			Title:    "Invalid Request",
			Status:   400,
			Detail:   "The description was larger than 1000 bytes. In UTF-8 encoding, the usual English characters are 1 byte each.",
			Instance: r.URL.Path,
		}.ToClient(w)
		return
	}

	description = strings.Replace(description, "\n", "&#13;", -1) // Cleanse 1
	description = strings.Replace(description, "\"", `\\"`, -1)   // Cleanse 2

	updatedSeries := types.Series{
		Id:           id,
		Title:        r.FormValue("title"),
		Description:  description,
		EpisodeCount: episodeCount,
		LastModified: time.Now().Format("01-02-2006 15:04:05"),
	}

	if _, err := db.Exec(`
			UPDATE
				series
			SET
				title = ?, description = ?, episodes = ?, last_modified = ?
			WHERE
				id = ?
		`,
		updatedSeries.Title,
		updatedSeries.Description,
		updatedSeries.EpisodeCount,
		updatedSeries.LastModified,
		updatedSeries.Id,
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
	}

	responses.Status{
		Status: 200,
	}.ToClient(w)
}
