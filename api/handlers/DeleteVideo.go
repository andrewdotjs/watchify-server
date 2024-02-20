package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/api/responses"
)

// Method: DELETE
func DeleteVideoHandler(w http.ResponseWriter, r *http.Request) {
	var videoIdentifer string = r.URL.Query().Get("v")
	var fileName string

	w.Header().Set("Content-Type", "application/json")

	if videoIdentifer == "" {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			ErrorCode:  "3",
			Message:    "v query param was not passed in.",
		}.ToJSON())
		return
	}

	database, err := sql.Open("sqlite3", "./db/videos.db")

	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	defer database.Close()
	err = database.QueryRow(`SELECT file_name FROM videos WHERE id=?`, videoIdentifer).Scan(&fileName)

	if err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			ErrorCode:  "20",
			Message:    "Unable to find video identifier from v.",
		}.ToJSON())
		return
	}

	err = os.Remove(path.Join("./storage/videos", fileName))

	if err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			ErrorCode:  "310",
			Message:    "Could not delete video from storage",
		}.ToJSON())
		return
	}

	if _, err = database.Exec(`DELETE FROM videos WHERE id=?`, videoIdentifer); err != nil {
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
