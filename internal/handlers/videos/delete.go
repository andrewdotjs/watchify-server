package videos

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/google/uuid"
)

// Deletes a video from the database and file system.
//
// # Specifications:
//   - Method   : DELETE
//   - Endpoint : /videos/{id}
//   - Auth?    : False
//
// # HTTP request path parameters:
//   - id : REQUIRED. UUID of the video.
func Delete(
  w http.ResponseWriter,
  r *http.Request,
  database *sql.DB,
  appDirectory *string,
  log *logger.Logger,
) {
	var fileName string = ""
	var id string = r.PathValue("id")
	var functionId string = uuid.NewString()

	if _, err := database.Exec(
	  `
  	  DELETE FROM
  			videos
  		WHERE
  		  id=?;
	  `,
		id); err != nil {
		responses.Status{
			Status:  500,
			Message: "Error deleting video information from the database.",
		}.ToClient(w)
		return
	}

	if err := os.Remove(path.Join(*appDirectory, "storage", "videos", fileName)); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			defer database.Close()
			log.Error(functionId, fmt.Sprintf("%v", err))
		}

		responses.Status{
			Status:  500,
			Message: "Error removing video file.",
		}.ToClient(w)
		return
	}

	responses.Status{
		Status: 200,
	}.ToClient(w)
}
