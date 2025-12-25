package videos

import (
	"database/sql"
	"net/http"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
	"github.com/andrewdotjs/watchify-server/internal/upload"
)

// Allows the client to upload a video to the file system and store its information to
// the database.
//
// # Specifications:
//   - Method      : POST
//   - Endpoint    : /videos
//   - Auth?       : False
//
// # HTTP form data:
//   - series-id   : REQUIRED. Series id.
//   - title       : REQUIRED. Video title.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - message     : If error, message detailing the error.
func Create(
  w http.ResponseWriter,
  r *http.Request,
  database *sql.DB,
  appDirectory *string,
) {
	var video types.Episode

	// Error handling if form data exceeds 1GB
	if err := r.ParseMultipartForm(1 << 30); err != nil {
		responses.Status{
			Status:  400,
			Message: "Did the file exceed 1GB?",
		}.ToClient(w)
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("video")
	if err != nil {
		responses.Status{
			Status:  400,
			Message: "Unable to get file from form. Was fileName set to video?",
		}.ToClient(w)
		return
	}

	defer file.Close()

	video.ParentId = r.FormValue("show-id")

	uploadDirectory := path.Join(*appDirectory, "storage", "videos")
	upload.Episode(handler, &video, database, &uploadDirectory)

	responses.Status{
		Status: 201,
	}.ToClient(w)
}
