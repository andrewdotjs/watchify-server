package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/types"
	"github.com/google/uuid"
)

func PostVideoHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	w.Header().Set("Content-Type", "application/json")
	var uploadDirectory string = "./storage/videos"
	var video types.Video

	// Error handling if form data exceeds 1GB
	if err := r.ParseMultipartForm(1 << 30); err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			Message:    "Did the file exceed 1GB?",
		}.ToJSON())
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("video")
	if err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			Message:    "Unable to get file from form. Was fileName set to video?",
		}.ToJSON())
		return
	} else {
		defer file.Close()
	}

	video.SeriesId = r.FormValue("series-identifier")
	video.Title = r.FormValue("title")
	video.Episode, err = strconv.Atoi(r.FormValue("episode-number"))
	if err != nil {
		log.Printf("ERR : %v. setting episode number to 0", err)
		video.Episode = 0
	}

	video.Id = fmt.Sprint(uuid.New())
	video.FileName = fmt.Sprintf("%s.%s", video.Id, strings.Split(handler.Filename, ".")[1])
	video.UploadDate = time.Now().Format("2006-01-02 15:04:05")

	if _, err := os.Stat(uploadDirectory); os.IsNotExist(err) {
		log.Printf("SYS : Could not find upload directory at '%s'. Creating one.", uploadDirectory)
		os.Mkdir(uploadDirectory, os.ModePerm)
	}

	uploadPath := filepath.Join(uploadDirectory, video.FileName)
	out, err := os.Create(uploadPath)
	if err != nil {
		w.WriteHeader(500)
		w.Write(responses.Error{
			StatusCode: 500,
			Message:    "Unable to create file.",
		}.ToJSON())
		return
	} else {
		defer out.Close()
	}

	// Copy the file to the destination and error handle.
	if _, err := io.Copy(out, file); err != nil {
		w.WriteHeader(500)
		w.Write(responses.Error{
			StatusCode: 500,
			Message:    "Unable to copy file.",
		}.ToJSON())
		return
	}

	if _, err = database.Exec(
		`INSERT INTO videos VALUES (?, ?, ?, ?, ?, ?)`,
		video.Id,
		video.SeriesId,
		video.Episode,
		video.Title,
		video.FileName,
		video.UploadDate,
	); err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			Message:    "Unable to insert data into the database.",
		}.ToJSON())
		return
	}

	w.WriteHeader(200)
	w.Write(responses.Video{
		StatusCode: 200,
		Data:       video,
	}.ToJSON())
}
