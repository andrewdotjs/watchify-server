package shows

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
	var orderedBy string = r.URL.Query().Get("orderedBy")
	var orderedByQuery string = ""
	var show types.Show = types.Show{}
	var functionId string = uuid.NewString()

	// Return all series if no ID.
	if id == "" {
		var shows []types.Show

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
					  shows
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
			var show types.Show

			if err := rows.Scan(
				&show.Id,
				&show.Title,
				&show.Description,
				&show.EpisodeCount,
				&show.UploadDate,
				&show.LastModified,
			); err != nil {
				defer database.Close()
				log.Error(functionId, fmt.Sprintf("%v", err))
			}

			show.Episodes = map[string]any{
				"count": show.EpisodeCount,
				"url":   ("/api/v1/series/" + show.Id + "/episodes"),
			}

			show.Cover = map[string]any{
				"exists": true,
				"url":    ("/api/v1/series/" + show.Id + "/cover"),
			}

			shows = append(shows, show)
		}

		responses.Status{
			Status: 200,
			Data:   shows,
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(
		`
		SELECT
			id, title, description, episode_count, upload_date, last_modified
		FROM
		  shows
		WHERE
			id=?
		`,
		id,
	).Scan(
		&show.Id,
		&show.Title,
		&show.Description,
		&show.EpisodeCount,
		&show.UploadDate,
		&show.LastModified,
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
	show.Episodes = map[string]any{
		"count": show.EpisodeCount,
		"url":   ("/api/v1/series/" + show.Id + "/episodes"),
	}

	show.Cover = map[string]any{
		"exists": true,
		"url":    ("/api/v1/series/" + show.Id + "/cover"),
	}

	responses.Status{
		Status: 200,
		Data:   show,
	}.ToClient(w)
}
