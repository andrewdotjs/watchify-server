package stream

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/andrewdotjs/watchify-server/internal/functions"
	"github.com/andrewdotjs/watchify-server/internal/responses"
)

// Returns a video stream to the client using the id.
//
// # Specifications:
//   - Method   : GET
//   - Endpoint : /stream/{id}
//   - Auth?    : False
//
// # HTTP request path parameters:
//   - id       : REQUIRED. UUID of the video.
func Read(
  w http.ResponseWriter,
  r *http.Request,
  database *sql.DB,
  appDirectory *string,
) {
	var id string = r.PathValue("id")
	var streamType string = "movies"
	var fileName string = ""

	if len(r.URL.Query().Get("type")) > 10 {
  	responses.Status{
  		Type:     "null",
  		Title:    "Unknown Error",
  		Status:   400,
  		Detail:   "Value in type query too large, limit is 10",
  		Instance: r.URL.Path,
  	}.ToClient(w)
  	return
	}

	if r.URL.Query().Get("type") == "show" {
    streamType = "episodes"
	}

	if err := database.QueryRow(
	  fmt.Sprintf(
  	  `
  			SELECT
  			  file_name
  			FROM
  			  %s
  			WHERE
  			  id=?
  		`,
      streamType,
		),
		id,
	).Scan(&fileName); err != nil {
	  responses.Status{
  		Type:     "null",
  		Title:    "Unknown Error",
  		Status:   500,
  		Detail:   fmt.Sprintf("%v", err),
  		Instance: r.URL.Path,
  	}.ToClient(w)
  	return
	}

	filePath := path.Join(*appDirectory, "storage", "videos", fileName)
	videoFile, err := os.Open(filePath)
	if err != nil {
	  responses.Status{
  		Type:     "null",
  		Title:    "Unknown Error",
  		Status:   500,
  		Detail:   fmt.Sprintf("%v", err),
  		Instance: r.URL.Path,
  	}.ToClient(w)
  	return
	}

	defer videoFile.Close()

	// Get file information
	fileInfo, err := videoFile.Stat()
	if err != nil {
	  responses.Status{
  		Type:     "null",
  		Title:    "Unknown Error",
  		Status:   500,
  		Detail:   fmt.Sprintf("%v", err),
  		Instance: r.URL.Path,
  	}.ToClient(w)
  	return
	}

	w.Header().Set("Content-Type", "video/mp4")

	// Parse the Range header to determine the requested byte range
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		ranges := strings.SplitN(rangeHeader[6:], "-", 2)
		start, _ := strconv.ParseInt(ranges[0], 10, 64)
		CHUNK_SIZE := math.Pow(10, 6)

		// Set the appropriate headers for partial content
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()-start))
		w.Header().Set(
			"Content-Range",
			fmt.Sprintf(
				"bytes %d-%d/%d",
				start,
				functions.Minimum((int(start)+int(CHUNK_SIZE)), int(fileInfo.Size())-1),
				fileInfo.Size(),
			),
		)

		// Seek to the specified position and stream the partial content
		videoFile.Seek(start, 0)
		http.ServeContent(w, r, filePath, fileInfo.ModTime(), videoFile)
		return
	}

	// If no Range header is present, serve the entire file
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	http.ServeContent(w, r, filePath, fileInfo.ModTime(), videoFile)
}
