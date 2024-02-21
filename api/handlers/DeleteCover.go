package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/api/responses"
)

func DeleteCoverHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var coverIdentifer string = r.URL.Query().Get("c")
	var fileName string

	w.Header().Set("Content-Type", "application/json")

	if coverIdentifer == "" {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			ErrorCode:  "3",
			Message:    "c query param was not passed in.",
		}.ToJSON())
		return
	}

	err := database.QueryRow(`SELECT file_name FROM covers WHERE id=?`, coverIdentifer).Scan(&fileName)
	if err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			ErrorCode:  "20",
			Message:    "Unable to find cover identifier from c.",
		}.ToJSON())
		return
	}

	err = os.Remove(path.Join("./storage/covers", fileName))

	if err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			ErrorCode:  "310",
			Message:    "Could not delete cover from storage",
		}.ToJSON())
		return
	}

	if _, err = database.Exec(`DELETE FROM videos WHERE id=?`, coverIdentifer); err != nil {
		w.WriteHeader(400)
		w.Write(responses.Error{
			StatusCode: 400,
			Message:    "Could not delete cover information from database.",
		}.ToJSON())
		return
	}

	w.WriteHeader(200)
	w.Write(responses.Status{
		StatusCode: 200,
	}.ToJSON())
}
