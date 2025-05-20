package movieCover

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/placeholders"
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
func Read(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var id string = r.PathValue("id")
	var uploadDirectory string = path.Join(*appDirectory, "storage", "covers")
	var coverFileName string

	if id == "" {
		log.Print("SYS : Did not find id. Sending placeholder cover.")
		fmt.Print("SYS : Did not find id. Sending placeholder cover.")

		responses.File{
			StatusCode: 200,
			FileBuffer: placeholders.Cover(),
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(
		"SELECT file_name FROM movie_covers WHERE movie_id=?",
		id,
	).Scan(
		&coverFileName,
	); err != nil {
		log.Printf("ERR : Error while querying covers, sending placeholder cover. %v", err)
		fmt.Printf("ERR : Error while querying covers, sending placeholder cover. %v", err)

		responses.File{
			StatusCode: 200,
			FileBuffer: placeholders.Cover(),
		}.ToClient(w)
		return
	}

	if buffer, err := os.ReadFile(path.Join(uploadDirectory, coverFileName)); err != nil {
		log.Printf("ERR : Error while querying covers. %v", err)
		fmt.Printf("ERR : Error while querying covers. %v", err)

		responses.File{
			StatusCode: 200,
			FileBuffer: placeholders.Cover(),
		}.ToClient(w)
	} else {
		responses.File{
			StatusCode: 200,
			FileBuffer: buffer,
		}.ToClient(w)
	}
}
