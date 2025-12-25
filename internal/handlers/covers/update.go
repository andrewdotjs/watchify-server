package covers

import (
	"database/sql"
	"net/http"

	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/google/uuid"
)

// use this route to update a cover for a series
func Update(
  w http.ResponseWriter,
  r *http.Request,
  db *sql.DB,
  appDirectory *string,
  log *logger.Logger,
) {
	var id string = r.PathValue("id")
	var functionId string = uuid.NewString()

	if id == "" {
	  log.Error(functionId, "Cover ID was not provided by the request")

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
