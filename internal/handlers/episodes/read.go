package episodes

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
)

// Gets and returns an array of episodes of a series stored in the database.
//
// # Specifications:
//   - Method      : GET
//   - Endpoint    : /series/{id}/episodes
//   - Auth?       : False
//
// # HTTP request path parameters:
//   - id          : REQUIRED. UUID of the series.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - message     : If error, Message detailing the error.
//   - data        : Series episodes, each returning id, episode.
func Read(
  w http.ResponseWriter,
  r *http.Request,
  db *sql.DB,
  log *logger.Logger,
) {
	var videos []types.Episode
	id := r.PathValue("id")

	if id == "" {
		responses.Error{
			Type:     "null",
			Title:    "Invalid Request",
			Status:   400,
			Detail:   "Id not passed in as a path parameter.",
			Instance: r.URL.Path,
		}.ToClient(w)
		return
	}

	rows, err := db.Query(
		`
   	SELECT
			id, episode_number
   	FROM
			episodes
   	WHERE
			parent_id=?
    `,
		id,
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
		var video types.Episode

		if err := rows.Scan(
			&video.Id,
			&video.EpisodeNumber,
		); err != nil {
			switch {
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
		} else {
			videos = append(videos, video)
		}
	}

	responses.Status{
		Status: 200,
		Data:   videos,
	}.ToClient(w)
}
