package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/api/types"
	"github.com/andrewdotjs/watchify-server/api/utilities"
)

func GetVideoHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var videoIdentifier string = r.URL.Query().Get("v")
	var video types.Video

	if videoIdentifier == "" {
		var queryLimit int
		var videoArray []types.Video

		if r.URL.Query().Get("limit") != "" {
			queryLimit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
		}

		if (queryLimit < 1) || (queryLimit > 20) {
			queryLimit = 20
		}

		rows, err := database.Query(`SELECT id, title FROM video WHERE series_id='' LIMIT ?;`, queryLimit)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				responses.Status{
					StatusCode: 200,
					Data:       videoArray,
				}.ToClient(w)
				return
			} else {
				log.Fatalf("ERR : %v", err)
			}
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

	err := database.QueryRow(
		"SELECT id, series_id, title FROM videos WHERE id=?;",
		videoIdentifier,
	).Scan(
		&video.Id,
		&video.SeriesId,
		&video.Episode,
		&video.Title,
		&video.FileName,
		&video.UploadDate,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			responses.Status{
				StatusCode: 400,
				Message:    "No database matche with the provided id.",
			}.ToClient(w)
			return
		}

		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	responses.Status{
		StatusCode: 200,
		Data:       video,
	}.ToClient(w)
}

func StreamHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var videoIdentifier string = r.URL.Query().Get("v")
	var fileName string
	var filePath string
	var err error

	if videoIdentifier == "" {
		responses.Status{
			StatusCode: 400,
			Message:    "Invalid query parameters.",
		}.ToClient(w)
		return
	}

	err = database.QueryRow("SELECT file_name FROM videos WHERE id=?", videoIdentifier).Scan(&fileName)

	filePath = path.Join(*appDirectory, "storage", "videos", fileName)
	videoFile, err := os.Open(filePath)
	if err != nil {
		responses.Status{
			StatusCode: 500,
			Message:    "Error opening video file.",
		}.ToClient(w)
		return
	}

	defer videoFile.Close()

	// Get file information
	fileInfo, err := videoFile.Stat()
	if err != nil {
		responses.Status{
			StatusCode: 500,
			Message:    "Error getting video file information.",
		}.ToClient(w)
		return
	}

	// Get total file size
	fileSize := fileInfo.Size()

	w.Header().Set("Content-Type", "video/mp4")

	// Parse the Range header to determine the requested byte range
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		ranges := strings.SplitN(rangeHeader[6:], "-", 2)
		start, _ := strconv.ParseInt(ranges[0], 10, 64)
		CHUNK_SIZE := math.Pow(10, 6)

		// Set the appropriate headers for partial content
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, utilities.Minimum((int(start)+int(CHUNK_SIZE)), (int(fileSize)-1)), fileSize))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize-start))
		w.WriteHeader(206)

		// Seek to the specified position and stream the partial content
		videoFile.Seek(start, 0)
		http.ServeContent(w, r, filePath, fileInfo.ModTime(), videoFile)
	} else {
		// If no Range header is present, serve the entire file
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
		http.ServeContent(w, r, filePath, fileInfo.ModTime(), videoFile)
	}
}

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

func DeleteVideoHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var videoIdentifer string = r.URL.Query().Get("v")
	var fileName string

	if videoIdentifer == "" {
		responses.Status{
			StatusCode: 400,
			Message:    "Invalid query parameters.",
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(`SELECT file_name FROM videos WHERE id=?;`, videoIdentifer).Scan(&fileName); err != nil {
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

	if _, err := database.Exec(`DELETE FROM videos WHERE id=?;`, videoIdentifer); err != nil {
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
