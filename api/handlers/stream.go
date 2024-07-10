package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/andrewdotjs/watchify-server/api/libs/server"
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
func StreamHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var fileName string
	id := r.PathValue("id")

	if err := database.QueryRow(`
	  SELECT file_name
		FROM videos
		WHERE id=?
		`,
		id,
	).Scan(
		&fileName,
	); err != nil {
		w.WriteHeader(500)
		return
	}

	filePath := path.Join(*appDirectory, "storage", "videos", fileName)
	videoFile, err := os.Open(filePath)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	defer videoFile.Close()

	// Get file information
	fileInfo, err := videoFile.Stat()
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")

	// Parse the Range header to determine the requested byte range
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		ranges := strings.SplitN(rangeHeader[6:], "-", 2)
		start, _ := strconv.ParseInt(ranges[0], 10, 64)
		CHUNK_SIZE := math.Pow(10, 6)

		// Set the appropriate headers for partial content
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, server.Minimum((int(start)+int(CHUNK_SIZE)), (int(fileInfo.Size())-1)), fileInfo.Size()))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()-start))

		// Seek to the specified position and stream the partial content
		videoFile.Seek(start, 0)
		http.ServeContent(w, r, filePath, fileInfo.ModTime(), videoFile)
		return
	}

	// If no Range header is present, serve the entire file
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	http.ServeContent(w, r, filePath, fileInfo.ModTime(), videoFile)
}
