package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/andrewdotjs/watchify-server/types"
	"github.com/andrewdotjs/watchify-server/utilities"
	_ "github.com/mattn/go-sqlite3"
)

func GetInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if len(r.URL.Query()) != 1 {
		var queryArray []string

		// Get all query parameters
		for key, values := range r.URL.Query() {
			for _, value := range values {
				queryArray = append(queryArray, fmt.Sprintf("%s=%s", key, value))
			}
		}

		response, err := json.Marshal(types.Message{
			StatusCode: http.StatusBadRequest,
			Message:    "Bad Request: Incorrect amount of query params passed in. Allowed amount is 1.",
			Queries:    queryArray,
		})

		// Catastrophic Area 001
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("001 - CATASTROPHIC SERVER FAILURE"))
			log.Fatal(err)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	database, err := sql.Open("sqlite3", "./db/videos.db")

	utilities.CheckError(err)

	rows, err := database.Query("SELECT * FROM videos WHERE id = ?", r.URL.Query().Get("v"))

	utilities.CheckError(err)

	response, err := json.Marshal(types.Message{
		StatusCode: http.StatusOK,
		Message:    "Good Request",
	})

	utilities.CheckError(err)

	w.WriteHeader(http.StatusOK)
	w.Write(response)
	defer database.Close()
}
