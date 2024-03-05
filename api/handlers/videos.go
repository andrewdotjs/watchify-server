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
	"github.com/gorilla/mux"
)

// Allows the client to retrieve the details of a specific uploaded video via passed in id.
//
// # Specifications:
//   - Method      : GET
//   - Endpoint    : /videos/{id}
//   - Auth?       : False
//
// # HTTP request path parameters:
//   - id          : REQUIRED. Video id.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - error_code  : If error, gives in-house error code for debugging. (not implemented yet)
//   - message     : If error, message detailing the error.
//   - data        : id, series_id, title (if empty, json data is empty)
func GetVideoHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	parameters := mux.Vars(r)
	id := parameters["id"]
	video := types.Video{Id: id}

	err := database.QueryRow("SELECT series_id, title FROM videos WHERE id=?;", video.Id).Scan(&video.SeriesId, &video.Title)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 200,
			Data:       nil,
		}.ToClient(w)
		return
	}

	responses.Status{
		StatusCode: 200,
		Data:       video,
	}.ToClient(w)
}

// Allows the client to retrieve the details of a specific uploaded video via passed in id.
//
// # Specifications:
//   - Method      : GET
//   - Endpoint    : /videos
//   - Auth?       : False
//
// # HTTP query parameters:
//   - limit       : OPTIONAL. Limit of how many videos to return at once.
//   - pagination  : OPTIONAL. Offset of query to allow for pages in client. offset = limit(page - 1).
//   - sort        : OPTIONAL. Sorts it either ascending it descending.
//   - search      : OPTIONAL. Hard searches for video by title.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - error_code  : If error, gives in-house error code for debugging. (not implemented yet)
//   - message     : If error, message detailing the error.
//   - data        : []{id, series_id, title} (if empty, json data is empty)
func GetAllVideosHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var queryLimit int
	var videoArray []types.Video

	if r.URL.Query().Get("limit") != "" {
		queryLimit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	}

	if (queryLimit < 1) || (queryLimit > 20) {
		queryLimit = 20
	}

	rows, err := database.Query(`SELECT id, title FROM videos WHERE series_id='' LIMIT ?;`, queryLimit)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 200,
			Data:       videoArray,
		}.ToClient(w)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var video types.Video

		if err := rows.Scan(&video.Id, &video.Title); err != nil {
			log.Fatalf("ERR : %v", err)
		}

		videoArray = append(videoArray, video)
	}

	responses.Status{
		StatusCode: 200,
		Data:       videoArray,
	}.ToClient(w)
	return
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
//   - error_code  : If error, gives in-house error code for debugging. (not implemented yet)
//   - message     : If error, message detailing the error.
func PostVideoHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var video types.Video

	// Error handling if form data exceeds 1GB
	if err := r.ParseMultipartForm(1 << 30); err != nil {
		responses.Status{
			StatusCode: 400,
			Message:    "Did the file exceed 1GB?",
		}.ToClient(w)
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("video")
	if err != nil {
		responses.Status{
			StatusCode: 400,
			Message:    "Unable to get file from form. Was fileName set to video?",
		}.ToClient(w)
		return
	}

	defer file.Close()

	video.SeriesId = r.FormValue("series-id")
	video.Title = r.FormValue("title")

	uploadDirectory := path.Join(*appDirectory, "storage", "videos")
	utilities.HandleVideoUpload(handler, &video, database, &uploadDirectory)

	responses.Status{
		StatusCode: 201,
	}.ToClient(w)
	return
}

// Deletes a video from the database and file system.
//
// # Specifications:
//   - Method   : DELETE
//   - Endpoint : /videos/{id}
//   - Auth?    : False
//
// # HTTP request query parameters:
//   - id : REQUIRED. Video id
func DeleteVideoHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var fileName string
	parameters := mux.Vars(r)

	id, ok := parameters["id"]
	if !ok {
		responses.Status{
			StatusCode: 400,
			Message:    "id is missing in path parameters",
		}.ToClient(w)
	}

	if _, err := database.Exec(`DELETE FROM videos WHERE id=?;`, id); err != nil {
		responses.Status{
			StatusCode: 500,
			Message:    "Error deleting video information from the database.",
		}.ToClient(w)
		return
	}

	if err := os.Remove(path.Join(*appDirectory, "storage", "videos", fileName)); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 500,
			Message:    "Error removing video file.",
		}.ToClient(w)
		return
	}

	responses.Status{
		StatusCode: 200,
	}.ToClient(w)
}
