package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/types"
	_ "github.com/mattn/go-sqlite3"
)

func GetInfoHandler(w http.ResponseWriter, r *http.Request) {
	var video types.Video
	var videoIdentifier string = r.URL.Query().Get("v")

	w.Header().Set("Content-Type", "application/json")

	// Check if parameters exceed maximum
	if len(r.URL.Query()) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(responses.Error{
			StatusCode: 200,
			ErrorCode:  "1",
			Message:    "Too many query params. Maximum is 1.",
		}.ToJSON())
		return
	}

	database, err := sql.Open("sqlite3", "./db/videos.db")

	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	if videoIdentifier == "" {
		rows, err := database.Query(`SELECT * FROM videos;`)
		var videoArray []types.Video

		if err != nil {
			if err != sql.ErrNoRows {
				log.Fatalf("ERR : %v", err)
			}
		}

		defer rows.Close()

		for rows.Next() {
			var video types.Video

			if err := rows.Scan(&video.Id, &video.SeriesId, &video.Episode, &video.Title, &video.FileName, &video.UploadDate); err != nil {
				log.Fatalf("ERR : %v", err)
			}
			videoArray = append(videoArray, video)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(responses.Videos{
			StatusCode: 200,
			Data:       videoArray,
		}.ToJSON())
		return
	}

	err = database.QueryRow("SELECT * FROM videos WHERE id = ?;", videoIdentifier).Scan(&video.Id, &video.SeriesId, &video.Episode, &video.Title, &video.FileName, &video.UploadDate)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(400)
			w.Write(responses.Error{
				StatusCode: 400,
				ErrorCode:  "40",
				Message:    fmt.Sprintf("No video matched the id %v.", videoIdentifier),
			}.ToJSON())
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responses.Error{
			StatusCode: 500,
			ErrorCode:  "40",
			Message:    "Catastrophic server failure has occurred.",
		}.ToJSON())
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	defer database.Close()

	w.WriteHeader(http.StatusOK)
	w.Write(responses.Video{
		StatusCode: 200,
		Data:       video,
	}.ToJSON())
}
