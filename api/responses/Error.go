package responses

import (
	"encoding/json"
	"log"
)

type Error struct {
	StatusCode int    `json:"status_code"` // HTTP status code
	ErrorCode  string `json:"error_code"`  // Custom code
	Message    string `json:"message"`     // Message quickly explaining error
}

func (v Error) ToJSON() []byte {
	json, err := json.Marshal(v)

	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	return json
}
