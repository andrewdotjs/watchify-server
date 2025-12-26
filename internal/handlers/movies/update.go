package movies

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/andrewdotjs/watchify-server/internal/functions"
	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
	"github.com/google/uuid"
)

func Update(
  w http.ResponseWriter,
  r *http.Request,
  db *sql.DB,
  appDirectory *string,
  log *logger.Logger,
) {
	var id string = r.PathValue("id")
	var movie types.Movie = types.Movie{}
	var oldMovie types.Movie = types.Movie{}
	var functionId string = uuid.NewString()

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


	if len(r.FormValue("title")) > 50 {
    log.Error(functionId, "The provided description was larger than 1000 bytes")
    responses.Error{
     	Type:     "null",
     	Title:    "Invalid Request",
     	Status:   400,
     	Detail:   "The title value was larger than 50 bytes. In UTF-8 encoding, English characters are 1 byte each.",
     	Instance: r.URL.Path,
    }.ToClient(w)
    return
	}

	// Check if description is larger than 1000 bytes
	if len(r.FormValue("description")) > 1000 {
	  log.Error(functionId, "The provided description was larger than 1000 bytes")
		responses.Error{
			Type:     "null",
			Title:    "Invalid Request",
			Status:   400,
			Detail:   "The description value was larger than 1000 bytes. In UTF-8 encoding, English characters are 1 byte each.",
			Instance: r.URL.Path,
		}.ToClient(w)
		return
	}

	// Check if hidden is larger than 5 bytes
	if len(r.FormValue("hidden")) > 5 {
	  log.Error(functionId, "The provided description was larger than 1000 bytes")
		responses.Error{
			Type:     "null",
			Title:    "Invalid Request",
			Status:   400,
			Detail:   "The hidden value was larger than 5 bytes. In UTF-8 encoding, English characters are 1 byte each.",
			Instance: r.URL.Path,
		}.ToClient(w)
		return
	}

	movie = types.Movie{
		Id:           id,
		Title:        functions.Sanitize(r.FormValue("title")),
		Hidden:       (r.FormValue("hidden") == "true"),
		Description:  functions.Sanitize(r.FormValue("description")),
		LastModified: time.Now().Format("01-02-2006 15:04:05"),
	}

	if err := db.QueryRow(
		`
		SELECT
			title, description, hidden
		FROM
			movies
		WHERE
			id = ?
		`,
		id,
	).Scan(
		&oldMovie.Title,
		&oldMovie.Description,
		&oldMovie.Hidden,
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

	if movie.Title == "" { movie.Title = oldMovie.Title }
	if movie.Description == "" { movie.Description = oldMovie.Description }

	if _, err := db.Exec(
  	`
   	  UPDATE
        movies
      SET
    		title = ?, description = ?, hidden = ?, last_modified = ?
     	WHERE
    		id = ?
  	`,
  	movie.Title,
  	movie.Description,
    movie.Hidden,
  	movie.LastModified,
  	movie.Id,
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
