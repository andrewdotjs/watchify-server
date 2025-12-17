package movies

import (
	"database/sql"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
	"github.com/andrewdotjs/watchify-server/internal/upload"
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
func Create(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {
	var uploadDirectory string = path.Join(*appDirectory, "storage", "videos")
	var uploadedVideo, uploadedCover []*multipart.FileHeader
	// var coverId string = uuid.NewString()
	var movieStruct types.Movie

	if err := r.ParseMultipartForm(1 << 40); err != nil { // Error handling if form data exceeds 1TB
		log.Printf("%v", err)
		fmt.Printf("%v", err)
		responses.Error{
			Type:   "null",
			Title:  "Incomplete request",
			Status: 400,
			Detail: "The upload form exceeded 1TB.",
		}.ToClient(w)
		return
	}

	uploadedCover = r.MultipartForm.File["cover"]
	if len(uploadedCover) == 0 {
		log.Print("Received no cover in request.")
		fmt.Print("Received no cover in request.")
		responses.Error{
			Type:   "null",
			Title:  "Bad request",
			Status: 400,
			Detail: "No uploaded cover present in form.",
		}.ToClient(w)
		return
	}

	uploadedVideo = r.MultipartForm.File["video"]
	if len(uploadedVideo) == 0 {
		log.Print("Received no videos in request.")
		fmt.Print("Received no videos in request.")
		responses.Error{
			Type:   "null",
			Title:  "Bad request",
			Status: 400,
			Detail: "No uploaded videos present in form.",
		}.ToClient(w)
		return
	}

	if len(uploadedVideo) > 1 {
		log.Print("Received too many videos in request.")
		fmt.Print("Received too many videos in request.")
		responses.Error{
			Type:   "null",
			Title:  "Bad request",
			Status: 400,
			Detail: "Too many uploaded videos present in form, limit is 1.",
		}.ToClient(w)
		return
	}

	movieStruct = types.Movie{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Hidden:      (r.FormValue("hidden") == "true"),
	}

	movieId, _ := upload.Movie(uploadedVideo[0], &movieStruct, database, &uploadDirectory)

	uploadDirectory = path.Join(*appDirectory, "storage", "covers")
	cover := types.Cover{
		ParentId: *movieId,
		UserId:  "",
	}

	if err := upload.Cover(
		uploadedCover[0],
		&cover,
		database,
		&uploadDirectory,
	); err != nil {
		fmt.Printf("%v", err)
	}

	responses.Status{
		Status: 201,
		Data:   movieStruct,
	}.ToClient(w)
}
