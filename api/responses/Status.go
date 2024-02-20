package responses

import (
	"encoding/json"
	"log"
)

type Status struct {
	StatusCode int `json:"status_code"`
}

func (v Status) ToJSON() []byte {
	json, err := json.Marshal(v)

	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	return json
}
