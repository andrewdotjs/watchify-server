package series

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
	var orderedBy string = r.URL.Query().Get("orderedBy")
	var orderedByQuery string
	var series types.Series

	// Return all series if no ID.
	if id == "" {
		var seriesArray []types.Series

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
						series
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
			var series types.Series

			if err := rows.Scan(
				&series.Id,
				&series.Title,
				&series.Description,
				&series.EpisodeCount,
				&series.UploadDate,
				&series.LastModified,
			); err != nil {
				defer database.Close()
				log.Fatalf("ERR : %v", err)
			}

			series.Episodes = map[string]any{
				"count": series.EpisodeCount,
				"url":   ("/api/v1/series/" + series.Id + "/episodes"),
			}

			series.Cover = map[string]any{
				"exists": true,
				"url":    ("/api/v1/series/" + series.Id + "/cover"),
			}

			seriesArray = append(seriesArray, series)
		}

		responses.Status{
			Status: 200,
			Data:   seriesArray,
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(
		`
		SELECT
			id, title, description, episode_count, upload_date, last_modified
		FROM
			series
		WHERE
			id=?
		`,
		id,
	).Scan(
		&series.Id,
		&series.Title,
		&series.Description,
		&series.EpisodeCount,
		&series.UploadDate,
		&series.LastModified,
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
	series.Episodes = map[string]any{
		"count": series.EpisodeCount,
		"url":   ("/api/v1/series/" + series.Id + "/episodes"),
	}

	series.Cover = map[string]any{
		"exists": true,
		"url":    ("/api/v1/series/" + series.Id + "/cover"),
	}

	responses.Status{
		Status: 200,
		Data:   series,
	}.ToClient(w)
}
