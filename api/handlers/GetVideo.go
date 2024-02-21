package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/types"
	_ "github.com/mattn/go-sqlite3"
)

func GetVideoHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	videoIdentifier := r.URL.Query().Get("v")
	var video types.Video

	w.Header().Set("Content-Type", "application/json")

	if videoIdentifier == "" {
		var queryLimit int = 20

		if r.URL.Query().Get("limit") != "" {
			queryLimit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
		}

		switch {
		case queryLimit < 1:
			queryLimit = 1
		case queryLimit > 20:
			queryLimit = 20
		}

		var videoArray []types.Video

		rows, err := database.Query(fmt.Sprintf(`SELECT * FROM videos LIMIT %v;`, queryLimit))
		if err != nil {
			if err != sql.ErrNoRows {
				log.Fatalf("ERR : %v", err)
			}
		}

		defer rows.Close()

		for rows.Next() {
			var video types.Video

			if err := rows.Scan(
				&video.Id,
				&video.SeriesId,
				&video.Episode,
				&video.Title,
				&video.FileName,
				&video.UploadDate,
			); err != nil {
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

	err := database.QueryRow(
		"SELECT * FROM videos WHERE id = ?;",
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
			w.WriteHeader(400)
			w.Write(responses.Error{
				StatusCode: 400,
				ErrorCode:  "40",
				Message:    fmt.Sprintf("No video matched the id %v.", videoIdentifier),
			}.ToJSON())
			return
		}

		defer database.Close()
		w.WriteHeader(500)
		w.Write(responses.Error{
			StatusCode: 500,
			ErrorCode:  "40",
			Message:    "Catastrophic server failure has occurred.",
		}.ToJSON())
		log.Fatalf("ERR : %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responses.Video{
		StatusCode: 200,
		Data:       video,
	}.ToJSON())
}
