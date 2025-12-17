package videos

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/responses"
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
func Delete(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var fileName string
	id := r.PathValue("id")

	if _, err := database.Exec(`
	  DELETE FROM videos
		WHERE id=?;
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
			log.Fatalf("ERR : %v", err)
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
