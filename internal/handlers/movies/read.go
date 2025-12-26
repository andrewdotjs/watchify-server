package movies

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/andrewdotjs/watchify-server/internal/logger"
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
func Read(
  w http.ResponseWriter,
  r *http.Request,
  database *sql.DB,
  log *logger.Logger,
) {
	var id string = r.PathValue("id")
	var hidden bool = (r.URL.Query().Get("hidden") == "true")
	var movieStruct types.Movie = types.Movie{}
	var functionId string = uuid.NewString()

	// Return all movies if no ID.
	if id == "" {
		var movieArray []types.Movie

		log.Info(functionId, "No ID given, attempting to return all movies")
		log.Info(functionId, "Querying database for all movies, limit 30")
		rows, err := database.Query(`
			SELECT
				id, title, description, upload_date, last_modified
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
		  if !errors.Is(err, sql.ErrNoRows) {
				log.Error(functionId, fmt.Sprintf("An unknown error occurred. %v", err))
				responses.Status{
					Type:     "null",
					Title:    "Unknown Error",
					Status:   500,
					Detail:   fmt.Sprintf("%v", err),
					Instance: r.URL.Path,
				}.ToClient(w)
				return
			}

			log.Info(functionId, "No movies present in database")
			responses.Status{
				Status: 200,
				Data:   nil,
			}.ToClient(w)
			return
		}

		defer rows.Close()
		for rows.Next() {
			var movie types.Movie

			if err := rows.Scan(
				&movie.Id,
				&movie.Title,
				&movie.Description,
				&movie.UploadDate,
				&movie.LastModified,
			); err != nil {
				defer database.Close()
				log.Error(functionId, fmt.Sprintf("An unknown error occurred. %v", err))
			}

			movie.Cover = map[string]any{
				"exists": true,
				"url":    ("/api/v1/movies/" + movie.Id + "/cover"),
			}

			movieArray = append(movieArray, movie)
		}

		log.Info(functionId, fmt.Sprintf("Returned %d movies", len(movieArray)))
		responses.Status{
			Status: 200,
			Data:   movieArray,
		}.ToClient(w)
		return
	}

	log.Info(functionId, fmt.Sprintf("ID %s given, attempting to return movie", id))

	if err := database.QueryRow(
		`
		SELECT
			id, title, description, hidden, file_extension, file_name, upload_date, last_modified
		FROM
			movies
		WHERE
			id = ?
		`,
		id,
	).Scan(
		&movieStruct.Id,
		&movieStruct.Title,
		&movieStruct.Description,
		&movieStruct.Hidden,
		&movieStruct.FileExtension,
		&movieStruct.FileName,
		&movieStruct.UploadDate,
		&movieStruct.LastModified,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		  log.Info(functionId, "No movie found with provided ID")
			responses.Error{
				Type:     "null",
				Title:    "Data not found",
				Status:   400,
				Detail:   "No movie could be found with the given id.",
				Instance: r.URL.Path,
			}.ToClient(w)
		default:
		  log.Error(functionId, fmt.Sprintf("An unknown error occurred. %v", err))
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

	log.Info(functionId, fmt.Sprintf("Successfully returned information on movie with ID %s", id))
	responses.Status{
		Status: 200,
		Data:   movieStruct,
	}.ToClient(w)
}
