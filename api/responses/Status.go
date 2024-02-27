package responses

import (
	"encoding/json"
	"log"
	"net/http"
)

type Status struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message,omitempty"`
	Data       any    `json:"data,omitempty"`
}

// Takes a built Status struct and converts it into JSON-compatible bytes
// using the "encoding/json" library then sends to client through provided
// ResponseWriter.
func (status Status) ToClient(w http.ResponseWriter) {
	json, err := json.Marshal(status)
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status.StatusCode)
	w.Write(json)
}
