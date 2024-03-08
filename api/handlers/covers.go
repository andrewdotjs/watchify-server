package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/api/responses"
	"github.com/andrewdotjs/watchify-server/api/utilities"
	"github.com/gorilla/mux"
)

// Returns the covers stored in the database and file-system. If none are present,
// return a placeholder cover.
//
// # Specifications:
//   - Method   : GET
//   - Endpoint : /series/{id}/cover
//   - Auth?    : False
//
// # HTTP request path parameters:
//   - id       : REQUIRED. UUID of the series.
func GetCoverHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var coverFileName string
	id := mux.Vars(r)["id"]

	if id == "" {
		responses.File{
			StatusCode: 400,
			FileBuffer: utilities.PlaceholderCover(),
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(`
	  SELECT file_name
		FROM covers
		WHERE series_id=?
		`,
		id,
	).Scan(
		&coverFileName,
	); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.File{
			StatusCode: 500,
			FileBuffer: utilities.PlaceholderCover(),
		}.ToClient(w)
		return
	}

	uploadDirectory := path.Join(*appDirectory, "storage", "covers")

	buffer, err := os.ReadFile(path.Join(uploadDirectory, coverFileName))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		responses.File{
			StatusCode: 500,
			FileBuffer: utilities.PlaceholderCover(),
		}.ToClient(w)
		return
	}

	responses.File{
		StatusCode: 200,
		FileBuffer: buffer,
	}.ToClient(w)
}
