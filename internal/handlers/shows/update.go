package shows

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	var episodeCount int = 0
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

	// update episode count
	if err := db.QueryRow(`
			SELECT
				COUNT(*) as count
			FROM
				episodes
			WHERE
				parent_id = ?
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
			log.Error(functionId, fmt.Sprintf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err))
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

	description = strings.ReplaceAll(description, "\n", "&#13;") // Cleanse 1
	description = strings.ReplaceAll(description, "\"", `\\"`)   // Cleanse 2

	updatedShow := types.Show{
		Id:           id,
		Title:        r.FormValue("title"),
		Description:  description,
		EpisodeCount: episodeCount,
		LastModified: time.Now().Format("01-02-2006 15:04:05"),
	}

	if _, err := db.Exec(`
			UPDATE
			  shows
			SET
				title = ?, description = ?, episodes = ?, last_modified = ?
			WHERE
				id = ?
		`,
		updatedShow.Title,
		updatedShow.Description,
		updatedShow.EpisodeCount,
		updatedShow.LastModified,
		updatedShow.Id,
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
			log.Error(functionId, fmt.Sprintf("Failed to give an accurate error response as it was not logged yet. Please log immediately. %v", err))
		}

		response.ToClient(w)
	}

	responses.Status{
		Status: 200,
	}.ToClient(w)
}
