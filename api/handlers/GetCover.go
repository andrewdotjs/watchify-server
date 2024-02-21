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

func GetCoverHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	coverIdentifier := r.URL.Query().Get("c")
	uploadDirectory := "./storage/covers"
	var coverFileName string

	w.Header().Set("Content-Type", "application/json")

	if coverIdentifier == "" {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			Message:    "c query param not passed in.",
		}.ToJSON())
		return
	}

	if err := database.QueryRow("SELECT file_name FROM covers WHERE id=?;", coverIdentifier).Scan(&coverFileName); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
		// if !errors.Is(err, sql.ErrNoRows) {}
	}

	buffer, err := os.ReadFile(path.Join(uploadDirectory, coverFileName))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("ERR : %v", err)
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(200)
	w.Write(buffer)
}
