package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/api/responses"
)

func DeleteVideoHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var videoIdentifer string = r.URL.Query().Get("v")
	var fileName string

	w.Header().Set("Content-Type", "application/json")

	if videoIdentifer == "" {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			Message:    "v query param was not passed in.",
		}.ToJSON())
		return
	}

	if err := database.QueryRow(`SELECT file_name FROM videos WHERE id=?;`, videoIdentifer).Scan(&fileName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(400)
			w.Write(responses.Error{
				StatusCode: 400,
				Message:    "Unable to find video identifier from v.",
			}.ToJSON())
			return
		} else {
			log.Fatalf("ERR : %v", err)
		}
	}

	if err := os.Remove(path.Join("./storage/videos", fileName)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(400)
			w.Write(responses.Error{
				StatusCode: 400,
				Message:    "Could not delete video from storage",
			}.ToJSON())
			return
		} else {
			log.Fatalf("ERR : %v", err)
		}
	}

	if _, err := database.Exec(`DELETE FROM videos WHERE id=?;`, videoIdentifer); err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			Message:    "Could not delete video information from database.",
		}.ToJSON())
		return
	}

	w.WriteHeader(200)
	w.Write(responses.Status{
		StatusCode: 200,
	}.ToJSON())
}
