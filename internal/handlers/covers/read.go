package covers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/placeholders"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/google/uuid"
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
func Read(
    w http.ResponseWriter,
    r *http.Request,
    database *sql.DB,
    appDirectory *string,
    log *logger.Logger,
) {
	var id string = r.PathValue("id")
	var uploadDirectory string = path.Join(*appDirectory, "storage", "covers")
	var coverFileName string = ""
	var functionId string = uuid.NewString()

	if id == "" {
	  log.Info(functionId, "Requested cover did not exist, sending placeholder.")

		responses.File{
			StatusCode: 200,
			FileBuffer: placeholders.Cover(),
		}.ToClient(w)
		return
	}

	if err := database.QueryRow(
		`
		  SELECT
				file_name
			FROM
			  covers
			WHERE
			  parent_id=?
		`,
		id,
	).Scan(
		&coverFileName,
	); err != nil {
	  log.Error(functionId, fmt.Sprintf("Database error, sending placeholder. %v", err))

		responses.File{
			StatusCode: 200,
			FileBuffer: placeholders.Cover(),
		}.ToClient(w)
		return
	}

	if buffer, err := os.ReadFile(path.Join(uploadDirectory, coverFileName)); err != nil {
		log.Error(functionId, fmt.Sprintf("File system error, sending placeholder. %v", err))

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
