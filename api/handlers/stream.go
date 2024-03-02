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

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/api/utilities"
	"github.com/gorilla/mux"
)

// Returns a video stream to the client using the id.
//
// Specifications:
//   - Method        : GET
//   - Endpoint      : api/v1/videos/stream
//   - Authorization : False
//
// HTTP request query parameters:
//   - id            : Required.
func StreamHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var fileName string
	var err error

	parameters := mux.Vars(r)
	id, ok := parameters["id"]

	if !ok {
		responses.Status{
			StatusCode: 400,
			Message:    "id is missing from path parameters",
		}.ToClient(w)
		return
	}

	err = database.QueryRow("SELECT file_name FROM videos WHERE id=?", id).Scan(&fileName)

	filePath := path.Join(*appDirectory, "storage", "videos", fileName)
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
