package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/andrewdotjs/watchify-server/types"
	_ "github.com/mattn/go-sqlite3"
)

func GetInfoHandler(w http.ResponseWriter, r *http.Request) {
	var video types.Video
	w.Header().Set("Content-Type", "application/json")

	if len(r.URL.Query()) != 1 {
		response, err := json.Marshal(types.Message{
			StatusCode: http.StatusBadRequest,
			Message:    "Bad Request: Incorrect amount of query params passed in. Allowed amount is 1.",
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("CATASTROPHIC SERVER FAILURE"))
			log.Fatalf("ERR : %v", err)
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	database, err := sql.Open("sqlite3", "./db/videos.db")

	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	err = database.QueryRow("SELECT * FROM videos WHERE id = ?", r.URL.Query().Get("v")).Scan(&video.Id, &video.SeriesId, &video.Episode, &video.Title, &video.UploadDate)

	if err != nil {
		if err == sql.ErrNoRows {
			response, _ := json.Marshal(types.Message{
				StatusCode: http.StatusNotFound,
				Message:    fmt.Sprintf("Could not find video where id = %v", r.URL.Query().Get("v")),
			})

			w.WriteHeader(http.StatusNotFound)
			w.Write(response)
			return
		}
		defer database.Close()
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("CATASTROPHIC SERVER FAILURE"))
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
