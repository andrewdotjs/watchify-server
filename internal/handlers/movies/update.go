package movies

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
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
	var updatedMovie types.Movie = types.Movie{}
	var changedValues []any = []any{}
	var query string = ""
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

	updatedMovie = types.Movie{
		Id:           id,
		Title:        r.FormValue("title"),
		Description:  r.FormValue("description"),
		LastModified: time.Now().Format("01-02-2006 15:04:05"),
	}

	// Check if description is larger than 1000 bytes
	if len(updatedMovie.Description) > 1000 {
	  log.Error(functionId, "The provided description was larger than 1000 bytes")
		responses.Error{
			Type:     "null",
			Title:    "Invalid Request",
			Status:   400,
			Detail:   "The description was larger than 1000 bytes. In UTF-8 encoding, the usual English characters are 1 byte each.",
			Instance: r.URL.Path,
		}.ToClient(w)
		return
	}

	updatedMovie.Description = strings.ReplaceAll(updatedMovie.Description, `\\"`, `"`) // Cleanse 1: Remove \"
	query, changedValues = functions.BuildUpdateQuery("movies", updatedMovie)

	if _, err := db.Exec(query, changedValues...); err != nil {
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
