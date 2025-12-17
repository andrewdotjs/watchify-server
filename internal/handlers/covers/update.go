package covers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/andrewdotjs/watchify-server/internal/responses"
)

// use this route to update a cover for a series
func Update(w http.ResponseWriter, r *http.Request, db *sql.DB, appDirectory *string) {
	id := r.PathValue("id")

	if id == "" {
		log.Print("SYS : Did not find id.")
		fmt.Print("SYS : Did not find id.")

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
