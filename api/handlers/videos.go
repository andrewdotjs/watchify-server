package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/api/types"
	"github.com/andrewdotjs/watchify-server/api/utilities"
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
	video := types.Video{Id: id}

	if id == "" {
		var videoArray []types.Video
		var queryLimit int

		queryLimitParam := r.URL.Query().Get("limit")
		if queryLimitParam != "" {
			if queryLimitTemp, err := strconv.Atoi(queryLimitParam); err != nil {
				queryLimit = 20
			} else {
				queryLimit = queryLimitTemp
			}
		}

		rows, err := database.Query(`
			SELECT id, title
			FROM videos
			WHERE series_id=''
			LIMIT ?
			`,
			queryLimit,
		)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				defer database.Close()
				log.Fatalf("ERR : %v", err)
			}

			responses.Status{
				Status: 200,
				Data:   videoArray,
			}.ToClient(w)
			return
		}

		defer rows.Close()

		for rows.Next() {
			var video types.Video

			if err := rows.Scan(
				&video.Id,
				&video.Title,
			); err != nil {
				log.Fatalf("ERR : %v", err)
			}

			videoArray = append(videoArray, video)
		}

		responses.Status{
			Status: 200,
			Data:   videoArray,
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(`
  	SELECT series_id, title, episode
  	FROM videos
  	WHERE id=?
  	`,
		video.Id,
	).Scan(
		&video.SeriesId,
		&video.Title,
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
	  SELECT id, episode
		FROM videos
		WHERE series_id=?
		AND (episode=?
		OR episode=?)
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
	var video types.Video

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
	video.Title = r.FormValue("title")

	uploadDirectory := path.Join(*appDirectory, "storage", "videos")
	utilities.HandleVideoUpload(handler, &video, database, &uploadDirectory)

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
