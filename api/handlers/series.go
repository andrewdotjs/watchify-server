package handlers

import (
	"database/sql"
	"net/http"
)

// Gets and returns an array of series stored in the database.
//
// Specifications:
//   - Method        : GET
//   - Endpoint      : api/v1/series
//   - Authorization : False
//
// HTTP request query parameters:
//   - id            : Optional. Will match provided id with series with same id, fails if not exists.
//   - limit         : Optional. Overrides the default limit 30 for returned rows.
//   - search        : Optional. Does a hard search for a specific series.
//
// HTTP response JSON contents:
//   - status_code   : HTTP status code.
//   - error_code    : If error, gives in-house error code for debugging. (not implemented yet)
//   - message       : If error, Message detailing the error.
//   - data          : Series contents, each returning id, episode count, title, description.
func GetSeriesHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {

}

// Uploads a series, its episodes, and its cover to the database and stores them within the
// storage folder.
//
// Specifications:
//   - Method        : POST
//   - Endpoint      : api/v1/series/upload
//   - Authorization : False
//
// HTTP request multipart form:
//   - video-files   : Required. Uploaded video files.
//   - name          : Required. Name of the soon-to-be uploaded series.
//   - description   : Required. Description of the soon-to-be series.
//
// HTTP response JSON contents:
//   - status_code   : HTTP status code.
//   - error_code    : If error, gives in-house error code for debugging. (not implemented yet)
//   - message       : If error, Message detailing the error.
//   - data          : Series id, title
func PostSeriesHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {

}

// Deletes a series, its episodes, and its cover from the database and storage folders.
//
// Specifications:
//   - Method        : DELETE
//   - Endpoint      : api/v1/series/delete
//   - Authorization : False
//
// Possible query parameters:
//   - id            : required, deletes series, videos, and covers that match the id from both db and storage.
//
// HTTP response JSON contents:
//   - status_code   : HTTP status code.
//   - error_code    : If error, gives in-house error code for debugging. (not implemented yet)
//   - message       : If error, Message detailing the error.
func DeleteSeriesHandler(w http.ResponseWriter, r *http.Request, database *sql.DB, appDirectory *string) {

}
