package movies

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/internal"
	"github.com/andrewdotjs/watchify-server/internal/responses"
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
func ReadCover(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var id string = r.PathValue("id")
	var uploadDirectory string = path.Join(*appDirectory, "storage", "covers")
	var coverFileName string

	if id == "" {
		log.Print("SYS : Did not find id. Sending placeholder cover.")
		fmt.Print("SYS : Did not find id. Sending placeholder cover.")

		responses.File{
			StatusCode: 200,
			FileBuffer: internal.PlaceholderCover(),
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(
		`
	  SELECT file_name
		FROM movie_covers
		WHERE movie_id=?
		`,
		id,
	).Scan(
		&coverFileName,
	); err != nil {
		log.Printf("ERR : Error while querying covers, sending placeholder cover. %v", err)
		fmt.Printf("ERR : Error while querying covers, sending placeholder cover. %v", err)

		responses.File{
			StatusCode: 200,
			FileBuffer: internal.PlaceholderCover(),
		}.ToClient(w)
		return
	}

	if buffer, err := os.ReadFile(path.Join(uploadDirectory, coverFileName)); err != nil {
		log.Printf("ERR : Error while querying covers. %v", err)
		fmt.Printf("ERR : Error while querying covers. %v", err)

		responses.File{
			StatusCode: 200,
			FileBuffer: internal.PlaceholderCover(),
		}.ToClient(w)
	} else {
		responses.File{
			StatusCode: 200,
			FileBuffer: buffer,
		}.ToClient(w)
	}
}

// use this route to update a cover for a series
func UpdateCover(w http.ResponseWriter, r *http.Request, db *sql.DB, appDirectory *string) {
	var id string = r.PathValue("id")

	if id == "" {
		log.Print("SYS : Did not find id.")
		fmt.Print("SYS : Did not find id.")

		responses.Error{
			Type:     "null",
			Title:    "Incomplete request",
			Status:   400,
			Detail:   "Did not receive id from url.",
			Instance: r.URL.Path,
		}.ToClient(w)
		return
	}
}
