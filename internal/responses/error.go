package responses

import (
	"encoding/json"
	"log"
	"net/http"
)

type Error struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

func (errorMessage Error) ToClient(w http.ResponseWriter) {
	if json, err := json.Marshal(errorMessage); err != nil {
		log.Fatalf("ERR : %v", err)
	} else {
		w.Header().Set("Content-Type", "application/problem+json")
		w.Header().Set("Content-Language", "en")
		w.WriteHeader(errorMessage.Status)
		w.Write(json)
	}
}
