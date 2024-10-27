package movies

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/internal/functions"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
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
func ReadMovie(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var id string = r.PathValue("id")
	var hidden bool = (r.URL.Query().Get("hidden") == "true")
	var movieStruct types.Movie

	// Return all series if no ID.
	if id == "" {
		var movieArray []types.Movie

		rows, err := database.Query(`
			SELECT
				id, title, description, file_extension, file_name, upload_date, last_modified
			FROM
				movies
			WHERE
			  hidden = ?
			LIMIT
				30
    	`,
			hidden,
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
			var movie types.Movie

			if err := rows.Scan(
				&movie.Id,
				&movie.Title,
				&movie.Description,
				&movie.FileExtension,
				&movie.FileName,
				&movie.UploadDate,
				&movie.LastModified,
			); err != nil {
				defer database.Close()
				log.Fatalf("ERR : %v", err)
			}

			movie.Cover = map[string]any{
				"exists": true,
				"url":    ("/api/v1/movies/" + movie.Id + "/cover"),
			}

			movieArray = append(movieArray, movie)
		}

		responses.Status{
			Status: 200,
			Data:   movieArray,
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(
		`
		SELECT
			id, title, description, hidden, upload_date, last_modified
		FROM
			movies
		WHERE
			id=?
		`,
		id,
	).Scan(
		&movieStruct.Id,
		&movieStruct.Title,
		&movieStruct.Description,
		&movieStruct.Hidden,
		&movieStruct.UploadDate,
		&movieStruct.LastModified,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			responses.Error{
				Type:     "null",
				Title:    "Data not found",
				Status:   400,
				Detail:   "No movie could be found with the given id.",
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

	movieStruct.Cover = map[string]any{
		"exists": true,
		"url":    ("/api/v1/movie/" + movieStruct.Id + "/cover"),
	}

	responses.Status{
		Status: 200,
		Data:   movieStruct,
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
func CreateMovie(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var uploadDirectory string = path.Join(*appDirectory, "storage", "videos")
	var uploadedVideo, uploadedCover []*multipart.FileHeader
	// var coverId string = uuid.NewString()
	var movieStruct types.Movie

	if err := r.ParseMultipartForm(1 << 40); err != nil { // Error handling if form data exceeds 1TB
		log.Printf("%v", err)
		responses.Error{
			Type:   "null",
			Title:  "Incomplete request",
			Status: 400,
			Detail: "The upload form exceeded 1TB.",
		}.ToClient(w)
		return
	}

	uploadedCover = r.MultipartForm.File["cover"]
	if len(uploadedCover) == 0 {
		log.Print("Received no cover in request.")
		responses.Error{
			Type:   "null",
			Title:  "Bad request",
			Status: 400,
			Detail: "No uploaded cover present in form.",
		}.ToClient(w)
		return
	}

	uploadedVideo = r.MultipartForm.File["video"]
	if len(uploadedVideo) == 0 {
		log.Print("Received no videos in request.")
		responses.Error{
			Type:   "null",
			Title:  "Bad request",
			Status: 400,
			Detail: "No uploaded videos present in form.",
		}.ToClient(w)
		return
	}

	if len(uploadedVideo) > 1 {
		log.Print("Received too many videos in request.")
		responses.Error{
			Type:   "null",
			Title:  "Bad request",
			Status: 400,
			Detail: "Too many uploaded videos present in form, limit is 1.",
		}.ToClient(w)
		return
	}

	movieStruct = types.Movie{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Hidden:      (r.FormValue("hidden") == "true"),
	}

	movieId, _ := functions.UploadMovie(uploadedVideo[0], &movieStruct, database, &uploadDirectory)

	uploadDirectory = path.Join(*appDirectory, "storage", "covers")
	cover := types.MovieCover{
		MovieId: *movieId,
		UserId:  "",
	}

	if err := functions.UploadMovieCover(
		uploadedCover[0],
		&cover,
		database,
		&uploadDirectory,
	); err != nil {
		fmt.Printf("%v", err)
	}

	responses.Status{
		Status: 201,
		Data:   movieStruct,
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
func DeleteMovie(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var id string = r.PathValue("id")
	var coverFileName string
	var movieFileName string

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

	// Stage 1, find the video that is being used by the to-be-deleted movie and delete it.
	if err := database.QueryRow(`
  	SELECT
			file_name
  	FROM
			movies
  	WHERE
			id=?
    `,
		id,
	).Scan(&movieFileName); err != nil && !errors.Is(err, sql.ErrNoRows) {
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

	if err := os.Remove(path.Join(*appDirectory, "storage", "videos", movieFileName)); err != nil {
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

	if _, err := database.Exec(`
	  DELETE FROM
			movies
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

	// Stage 2, delete the cover from the database.
	if _, err := database.Exec(`
	  DELETE FROM
			movie_covers
	  WHERE
			movie_id=?
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

	// Stage 4, delete the movie itself from the database.
	if _, err := database.Exec(`
  	DELETE FROM
			movie
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

func UpdateMovie(w http.ResponseWriter, r *http.Request, db *sql.DB, appDirectory *string) {
	var id string = r.PathValue("id")
	var updatedMovie types.Movie
	var changedValues []interface{}
	var query string

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

	updatedMovie = types.Movie{
		Id:           id,
		Title:        r.FormValue("title"),
		Description:  r.FormValue("description"),
		LastModified: time.Now().Format("01-02-2006 15:04:05"),
	}

	// Check if description is less than 1000 bytes
	if len(updatedMovie.Description) > 1000 {
		responses.Error{
			Type:     "null",
			Title:    "Invalid Request",
			Status:   400,
			Detail:   "The description was larger than 1000 bytes. In UTF-8 encoding, the usual English characters are 1 byte each.",
			Instance: r.URL.Path,
		}.ToClient(w)
		return
	}

	updatedMovie.Description = strings.Replace(updatedMovie.Description, `\\"`, `"`, -1) // Cleanse 1: Remove \"
	query, changedValues = functions.BuildUpdateQuery("movies", updatedMovie)

	if _, err := db.Exec(query, changedValues...); err != nil {
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
