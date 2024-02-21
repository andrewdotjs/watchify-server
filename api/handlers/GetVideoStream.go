package handlers

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/andrewdotjs/watchify-server/api/utilities"
)

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	videoIdentifier := r.URL.Query().Get("v")
	w.Header().Set("Content-Type", "video/mp4")

	if videoIdentifier == "" {
		http.Error(w, "No video identifier passed in. Please use v={id}.", 500)
		return
	}

	var videoFilePath string = fmt.Sprintf("./storage/videos/%v.mp4", videoIdentifier)
	videoFile, err := os.Open(videoFilePath)
	if err != nil {
		http.Error(w, "Error opening video file", 500)
		return
	}
	defer videoFile.Close()

	// Get file information
	fileInfo, err := videoFile.Stat()
	if err != nil {
		http.Error(w, "Error getting file information", 500)
		return
	}

	// Get total file size
	fileSize := fileInfo.Size()

	// Parse the Range header to determine the requested byte range
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		ranges := strings.SplitN(rangeHeader[6:], "-", 2)
		start, _ := strconv.ParseInt(ranges[0], 10, 64)
		CHUNK_SIZE := math.Pow(10, 6)

		// Set the appropriate headers for partial content
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, utilities.Minimum((int(start)+int(CHUNK_SIZE)), (int(fileSize)-1)), fileSize))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize-start))
		w.WriteHeader(http.StatusPartialContent)

		// Seek to the specified position and stream the partial content
		videoFile.Seek(start, 0)
		http.ServeContent(w, r, videoFilePath, fileInfo.ModTime(), videoFile)
	} else {
		// If no Range header is present, serve the entire file
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
		http.ServeContent(w, r, videoFilePath, fileInfo.ModTime(), videoFile)
	}
}
