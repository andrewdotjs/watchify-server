package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/andrewdotjs/watchify-server/types"
	"github.com/andrewdotjs/watchify-server/utilities"
	_ "github.com/mattn/go-sqlite3"
)

func GetInfoHandler(w http.ResponseWriter, r *http.Request) {
	var video types.Video
	var videoIdentifier string = r.URL.Query().Get("v")

	w.Header().Set("Content-Type", "application/json")

	if len(r.URL.Query()) > 1 {
		response := utilities.ErrorMessage(http.StatusOK, "You cannot have more than 1 query param at this time.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	database, err := sql.Open("sqlite3", "./db/videos.db")

	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	if videoIdentifier == "" {
		rows, err := database.Query(`SELECT * FROM videos`)
		videos := types.Videos{}

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
			videos.VideoArray = append(videos.VideoArray, video)
		}

		response, _ := json.Marshal(videos)

		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return
	}

	err = database.QueryRow("SELECT * FROM videos WHERE id = ?", videoIdentifier).Scan(&video.Id, &video.SeriesId, &video.Episode, &video.Title, &video.FileName, &video.UploadDate)

	if err != nil {
		if err == sql.ErrNoRows {
			response := utilities.ErrorMessage(http.StatusOK, fmt.Sprintf("Could not find video where v=%s", videoIdentifier))

			w.WriteHeader(http.StatusNotFound)
			w.Write(response)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("CATASTROPHIC SERVER FAILURE"))
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	response, _ := json.Marshal(types.Message{
		StatusCode: http.StatusOK,
		Video:      video,
	})

	w.WriteHeader(http.StatusOK)
	w.Write(response)
	defer database.Close()
}
