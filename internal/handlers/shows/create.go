package shows

import (
	"database/sql"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
	"github.com/andrewdotjs/watchify-server/internal/upload"
	"github.com/google/uuid"
)

// Uploads a series, its episodes, and its cover to the database and stores them within the
// storage folder.
//
// # Specifications:
//   - Method      : POST
//   - Endpoint    : /series
//   - Auth?       : False
//
// # HTTP request multipart form:
//   - video-files : REQUIRED. Uploaded video files.
//   - name        : REQUIRED. Name of the soon-to-be uploaded series.
//   - description : REQUIRED. Description of the series.
//
// # HTTP response JSON contents:
//   - status_code : HTTP status code.
//   - message     : If error, Message detailing the error.
//   - data        : Series id, title
func Create(
  w http.ResponseWriter,
  r *http.Request,
  database *sql.DB,
  appDirectory *string,
  log *logger.Logger,
) {
  var functionId string = uuid.NewString()

	if err := r.ParseMultipartForm(1 << 40); err != nil { // Error handling if form data exceeds 1TB
		log.Error(functionId, fmt.Sprintf("%v", err))
		responses.Error{
			Type:   "null",
			Title:  "Incomplete request",
			Status: 400,
			Detail: "The upload form exceeded 1TB",
		}.ToClient(w)
		return
	}

	uploadDirectory := path.Join(*appDirectory, "storage", "videos")
	currentTime := time.Now().Format("01-02-2006 15:04:05")
	uploadedVideos := r.MultipartForm.File["videos"]

	uploadedCover := r.MultipartForm.File["cover"]
	if len(uploadedCover) == 0 {
	  log.Error(functionId, "Received no cover in request")
		responses.Error{
			Type:   "null",
			Title:  "Bad request",
			Status: 400,
			Detail: "No uploaded cover present in form",
		}.ToClient(w)
		return
	}
	//if err != nil {
	//	var response responses.Error
	//
	//	switch {
	//	case errors.Is(err, http.ErrMissingFile):
	//		log.Print("Received no cover in request.")
	//		response = responses.Error{
	//			Type:   "null",
	//			Title:  "Bad request",
	//			Status: 400,
	//			Detail: "No uploaded cover present in form",
	//		}
	//	default:
	//		log.Print(err)
	//		response = responses.Error{
	//			Type:   "null",
	//			Title:  "Unaccounted Error",
	//			Status: 500,
	//			Detail: fmt.Sprintf("%v", err),
	//		}
	//	}
	//
	//	response.ToClient(w)
	//	return
	//}

	if len(uploadedVideos) == 0 {
		log.Error(functionId, "Received no videos in request")
		responses.Error{
			Type:   "null",
			Title:  "Bad request",
			Status: 400,
			Detail: "No uploaded videos present in form",
		}.ToClient(w)
		return
	}

	show := types.Show{
		Id:           uuid.New().String(),
		Title:        r.FormValue("title"),
		Description:  r.FormValue("description"),
		UploadDate:   currentTime,
		LastModified: currentTime,
	}

	// Handle upload for every file that was passed in the form.
	for index, uploadedFile := range uploadedVideos {
		video := types.Episode{ParentId: show.Id}
		upload.Episode(uploadedFile, &video, database, &uploadDirectory, log, &functionId)
		show.EpisodeCount = index + 1
	}

	uploadDirectory = path.Join(*appDirectory, "storage", "covers")
	cover := types.Cover{ParentId: show.Id}

	upload.Cover(
		uploadedCover[0],
		&cover,
		database,
		&uploadDirectory,
		log,
		&functionId,
	)

	if _, err := database.Exec(`
   	INSERT INTO
      shows
   	VALUES
			(?, ?, ?, ?, ?, ?, ?)
    `,
		show.Id,
		show.Title,
		show.Description,
		show.EpisodeCount,
		show.Hidden,
		show.UploadDate,
		show.LastModified,
	); err != nil {
	  log.Error(functionId, fmt.Sprintf("%v", err))
	}

	responses.Status{
		Status: 201,
		Data:   show,
	}.ToClient(w)
}
