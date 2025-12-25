package covers

import (
	"database/sql"
	"net/http"

	"github.com/andrewdotjs/watchify-server/internal/logger"
)

// use this route to delete a cover
func Delete(
  w http.ResponseWriter,
  r *http.Request,
  db *sql.DB,
  appDirectory *string,
  log *logger.Logger,
) {}
