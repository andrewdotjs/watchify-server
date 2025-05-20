package movies

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

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
func Read(w http.ResponseWriter, r *http.Request, database *sql.DB) {
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
