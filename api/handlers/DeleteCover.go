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

func DeleteCoverHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var coverIdentifer string = r.URL.Query().Get("c")
	var fileName string

	w.Header().Set("Content-Type", "application/json")

	if coverIdentifer == "" {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			Message:    "c query param was not passed in.",
		}.ToJSON())
		return
	}

	if err := database.QueryRow(`SELECT file_name FROM covers WHERE id=?`, coverIdentifer).Scan(&fileName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(400)
			w.Write(responses.Error{
				StatusCode: 400,
				Message:    "Unable to find cover identifier from c.",
			}.ToJSON())
			return
		} else {
			log.Fatalf("ERR : %v", err)
		}
	}

	if err := os.Remove(path.Join("./storage/covers", fileName)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(400)
			w.Write(responses.Error{
				StatusCode: 400,
				Message:    "Could not delete cover from storage",
			}.ToJSON())
			return
		} else {
			log.Fatalf("ERR : %v", err)
		}
	}

	if _, err := database.Exec(`DELETE FROM videos WHERE id=?`, coverIdentifer); err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			Message:    "Could not delete cover information from database.",
		}.ToJSON())
		return
	}

	w.WriteHeader(200)
	w.Write(responses.Status{
		StatusCode: 200,
	}.ToJSON())
}
