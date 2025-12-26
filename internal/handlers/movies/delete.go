package movies

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/google/uuid"
)

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
func Delete(
  w http.ResponseWriter,
  r *http.Request,
  database *sql.DB,
  appDirectory *string,
  log *logger.Logger,
) {
	var id string = r.PathValue("id")
	var functionId string = uuid.NewString()
	var coverFileName string = ""
	var movieFileName string = ""


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
		  log.Error(functionId, fmt.Sprintf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err))
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
			log.Error(functionId, fmt.Sprintf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err))
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
		log.Error(functionId, fmt.Sprintf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err))
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

	// Stage 2, find the cover that is being used by the to-be-deleted movie and delete it.
	if err := database.QueryRow(`
  	SELECT
			file_name
  	FROM
      covers
  	WHERE
			parent_id=?
    `,
		id,
	).Scan(&coverFileName); err != nil && !errors.Is(err, sql.ErrNoRows) {
		var response responses.Error

		switch {
		default:
		  log.Error(functionId, fmt.Sprintf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err))
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
			covers
	  WHERE
			parent_id=?
  	`,
		id,
	); err != nil {
		var response responses.Error

		switch {
		default:
		  log.Error(functionId, fmt.Sprintf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err))
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

	// Stage 4, delete the cover from the storage.
	if err := os.Remove(
		path.Join(*appDirectory, "storage", "covers", coverFileName),
	); err != nil && !errors.Is(err, os.ErrNotExist) {
		var response responses.Error

		switch {
		default:
		  log.Error(functionId, fmt.Sprintf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err))
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

	// Stage 5, delete the movie itself from the database.
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
		  log.Error(functionId, fmt.Sprintf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err))
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
