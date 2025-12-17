package shows

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/responses"
)

// Deletes a series, its episodes, and its cover from the database and storage folders.
//
// # Specifications:
//   - Method      : DELETE
//   - Endpoint    : /shows/{id}
//   - Auth?       : False
//
// # HTTP request path parameters:
//   - id          : REQUIRED. UUID of the series.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - message     : If error, Message detailing the error.
func Delete(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
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
			episodes
  	WHERE
			parent_id=?
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
