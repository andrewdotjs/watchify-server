package movies

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/internal/functions"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
)

func Update(w http.ResponseWriter, r *http.Request, db *sql.DB, appDirectory *string) {
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
			fmt.Printf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err)
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
