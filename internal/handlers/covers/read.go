package covers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/placeholders"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
	"github.com/google/uuid"
)

// Returns the covers stored in the database and filesystem. If none are present,
// return a placeholder cover.
//
// # Specifications:
//   - Method   : GET
//   - Endpoint : /shows/{id}/cover
//   - Auth?    : False
//
// # HTTP request path parameters:
//   - id       : REQUIRED. UUID of the show.
func Read(
    w http.ResponseWriter,
    r *http.Request,
    database *sql.DB,
    appDirectory *string,
    log *logger.Logger,
) {
	var id string = r.PathValue("id")
	var uploadDirectory string = path.Join(*appDirectory, "storage", "covers")
	var functionId string = uuid.NewString()
	var cover types.Cover = types.Cover{}

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
			id, file_extension, file_name, upload_date
		FROM
			movies
		WHERE
			id = ?
		`,
		id,
	).Scan(
		&cover.Id,
		&cover.FileExtension,
		&cover.FileName,
		&cover.UploadDate,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		  log.Info(functionId, "No movie found with provided ID")
			responses.Error{
				Type:     "null",
				Title:    "Data not found",
				Status:   400,
				Detail:   "No movie could be found with the given id.",
				Instance: r.URL.Path,
			}.ToClient(w)
		default:
		  log.Error(functionId, fmt.Sprintf("An unknown error occurred. %v", err))
			responses.Error{
				Type:     "null",
				Title:    "Unknown Error",
				Status:   500,
				Detail:   fmt.Sprintf("%v", err),
				Instance: r.URL.Path,
			}.ToClient(w)
		}

		return
	}

	if buffer, err := os.ReadFile(path.Join(uploadDirectory, cover.FileName)); err != nil {
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
