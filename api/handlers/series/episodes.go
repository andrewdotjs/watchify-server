package series

import (
	"database/sql"
	"errors"
	"log"
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
func ReadEpisodes(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var videos []types.Video
	id := r.PathValue("id")

	if id == "" {
		responses.Status{
			Status:  400,
			Message: "Id not passed in as a path parameter.",
		}.ToClient(w)
		return
	}

	rows, err := database.Query(`
   	SELECT id, title, episode
   	FROM videos
   	WHERE series_id=?
    `,
		id,
	)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			Status: 200,
			Data:   nil,
		}.ToClient(w)
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
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		videos = append(videos, video)
	}

	responses.Status{
		Status: 200,
		Data:   videos,
	}.ToClient(w)
}

func UpdateEpisodes() {}
func DeleteEpisodes() {}
