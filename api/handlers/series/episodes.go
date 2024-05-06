package series

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/api/types"
)

func CreateEpisodes() {}

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
func ReadEpisodes(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var videos []types.Video
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
			id, title, episode
   	FROM
			videos
   	WHERE
			series_id=?
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
		var video types.Video

		if err := rows.Scan(
			&video.Id,
			&video.Title,
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

func UpdateEpisodes() {}
func DeleteEpisodes() {}
