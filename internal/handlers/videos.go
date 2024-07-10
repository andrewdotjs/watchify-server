package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/functions"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
)

// Allows the client to retrieve the details of a specific uploaded video via passed in id.
//
// # Specifications:
//   - Method      : GET
//   - Endpoint    : /videos/{id}
//   - Auth?       : False
//
// # HTTP request path parameters:
//   - id          : REQUIRED. UUID of the video.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - message     : If error, message detailing the error.
//   - data        : id, series_id, title (if empty, json data is empty)
func ReadVideo(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	id := r.PathValue("id")
	video := types.Episode{Id: id}

	if id == "" {
		responses.Error{
			Type:     "null",
			Title:    "Invalid API request",
			Status:   400,
			Detail:   "No id was provided in the url params.",
			Instance: r.URL.Path,
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(`
  	SELECT
			series_id, episode_number
  	FROM
			series_episodes
  	WHERE
			id=?
  	`,
		video.Id,
	).Scan(
		&video.SeriesId,
		&video.EpisodeNumber,
	); err != nil {
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

	if video.SeriesId != "" {
		rows, err := database.Query(`
	  SELECT
			id, episode_number
		FROM
			series_episodes
		WHERE
			series_id=?
		AND
			(episode_number=? OR episode_number=?)
		`,
			video.SeriesId,
			video.EpisodeNumber-1,
			video.EpisodeNumber+1,
		)

		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				defer database.Close()
				log.Fatalf("ERR : %v", err)
			}

			responses.Status{
				Status: 200,
				Data:   video,
			}.ToClient(w)
			return
		}

		for rows.Next() {
			var id string
			var episodeNumber int

			if err := rows.Scan(
				&id,
				&episodeNumber,
			); err != nil {
				defer database.Close()
				log.Fatalf("ERR : %v", err)
			}

			videoAdjacent := map[string]string{
				"id":  id,
				"url": "/videos/" + id,
			}

			if episodeNumber == video.EpisodeNumber+1 {
				video.NextEpisode = videoAdjacent
			}

			if episodeNumber == video.EpisodeNumber-1 {
				video.PreviousEpisode = videoAdjacent
			}
		}
	}

	responses.Status{
		Status: 200,
		Data:   video,
	}.ToClient(w)
}

// Allows the client to upload a video to the file system and store its information to
// the database.
//
// # Specifications:
//   - Method      : POST
//   - Endpoint    : /videos
//   - Auth?       : False
//
// # HTTP form data:
//   - series-id   : REQUIRED. Series id.
//   - title       : REQUIRED. Video title.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - message     : If error, message detailing the error.
func CreateVideo(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var video types.Episode

	// Error handling if form data exceeds 1GB
	if err := r.ParseMultipartForm(1 << 30); err != nil {
		responses.Status{
			Status:  400,
			Message: "Did the file exceed 1GB?",
		}.ToClient(w)
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("video")
	if err != nil {
		responses.Status{
			Status:  400,
			Message: "Unable to get file from form. Was fileName set to video?",
		}.ToClient(w)
		return
	}

	defer file.Close()

	video.SeriesId = r.FormValue("series-id")

	uploadDirectory := path.Join(*appDirectory, "storage", "videos")
	functions.SeriesEpisode(handler, &video, database, &uploadDirectory)

	responses.Status{
		Status: 201,
	}.ToClient(w)
}

// Deletes a video from the database and file system.
//
// # Specifications:
//   - Method   : DELETE
//   - Endpoint : /videos/{id}
//   - Auth?    : False
//
// # HTTP request path parameters:
//   - id : REQUIRED. UUID of the video.
func DeleteVideo(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var fileName string
	id := r.PathValue("id")

	if _, err := database.Exec(`
	  DELETE FROM videos
		WHERE id=?;
	  `,
		id); err != nil {
		responses.Status{
			Status:  500,
			Message: "Error deleting video information from the database.",
		}.ToClient(w)
		return
	}

	if err := os.Remove(path.Join(*appDirectory, "storage", "videos", fileName)); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			Status:  500,
			Message: "Error removing video file.",
		}.ToClient(w)
		return
	}

	responses.Status{
		Status: 200,
	}.ToClient(w)
}
