package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/andrewdotjs/watchify-server/types"
	"github.com/andrewdotjs/watchify-server/utilities"
)

func DeleteVideoHandler(w http.ResponseWriter, r *http.Request) {
	var videoIdentifer string = r.URL.Query().Get("v")
	var fileName string
	w.Header().Set("Content-Type", "application/json")

	if videoIdentifer == "" {
		response := utilities.ErrorMessage(http.StatusBadRequest, "v query param was not passed in.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	database, err := sql.Open("sqlite3", "./db/videos.db")

	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	defer database.Close()
	err = database.QueryRow(`SELECT file_name FROM videos WHERE id=?`, videoIdentifer).Scan(&fileName)

	if err != nil {
		response := utilities.ErrorMessage(http.StatusBadRequest, "Unable to find video identifier from v. Does that ID exist?")
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	err = os.Remove(path.Join("./storage/videos", fileName))

	if err != nil {
		response := utilities.ErrorMessage(http.StatusBadRequest, "Unable to delete video from storage. Does that ID exist?")
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	if _, err = database.Exec(`DELETE FROM videos WHERE id=?`, videoIdentifer); err != nil {
		response := utilities.ErrorMessage(http.StatusBadRequest, "Unable to delete video from database. Does that ID exist?")
		log.Printf("ERR : %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	response, err := json.Marshal(types.Message{
		StatusCode: http.StatusOK,
		Message:    "Video was successfully deleted from both storage and database.",
	})

	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
