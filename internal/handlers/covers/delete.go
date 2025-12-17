package covers

import (
	"database/sql"
	"net/http"
)

// use this route to delete a cover from a series.
func Delete(w http.ResponseWriter, r *http.Request, db *sql.DB, appDirectory *string) {}
