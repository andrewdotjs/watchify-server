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

// Allows the user to upload a file to the file system and store its information to
// the database.
//
// Specifications:
//   - Method        : GET
//   - Endpoint      : api/v1/videos
//   - Authorization : False
//
// HTTP request path parameters:
//   - id
//
// HTTP response JSON contents:
//   - status_code   : HTTP status code.
//   - error_code    : If error, gives in-house error code for debugging. (not implemented yet)
//   - message       : If error, message detailing the error.
//   - data          : id, title
func GetVideoHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var video types.Video
	parameters := mux.Vars(r)

	id, _ := parameters["id"]

	log.Println(id)

	err := database.QueryRow("SELECT id, series_id, title FROM videos WHERE id=?;", id).Scan(&video.Id, &video.SeriesId, &video.Episode, &video.Title, &video.FileName, &video.UploadDate)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 400,
			Message:    "No database matches with provided id.",
		}.ToClient(w)
		return
	}

	responses.Status{
		StatusCode: 200,
		Data:       video,
	}.ToClient(w)
}

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

// Allows the user to upload a video to the file system and store its information to
// the database.
//
// Specifications:
//   - Method        : GET
//   - Endpoint      : api/v1/videos
//   - Authorization : False
//
// HTTP request path parameters:
//   - id
//
// HTTP response JSON contents:
//   - status_code   : HTTP status code.
//   - error_code    : If error, gives in-house error code for debugging. (not implemented yet)
//   - message       : If error, message detailing the error.
//   - data          : id, title
func PostVideoHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var uploadDirectory string = path.Join(*appDirectory, "storage", "videos")
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

	video.SeriesId = r.FormValue("series-identifier")
	video.Title = r.FormValue("title")

	utilities.HandleVideoUpload(handler, &video, database, &uploadDirectory)

	responses.Status{
		StatusCode: 200,
		Data:       video,
	}.ToClient(w)
	return
}

// Deletes a video from the database and file system.
//
// Specifications:
//   - Method        : DELETE
//   - Endpoint      : api/v1/videos
//   - Authorization : False
//
// HTTP request query parameters:
//   - id            : Required.
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

	if err := database.QueryRow(`SELECT file_name FROM videos WHERE id=?;`, id).Scan(&fileName); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.Status{
			StatusCode: 500,
			Message:    "No database match with the provided id.",
		}.ToClient(w)
		return
	}

	if err := os.Remove(path.Join("./storage/videos", fileName)); err != nil {
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

	if _, err := database.Exec(`DELETE FROM videos WHERE id=?;`, id); err != nil {
		responses.Status{
			StatusCode: 500,
			Message:    "Error deleting video information from the database.",
		}.ToClient(w)
		return
	}

	responses.Status{
		StatusCode: 200,
	}.ToClient(w)
}
