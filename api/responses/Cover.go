package responses

import (
	"encoding/json"
	"log"

	"github.com/andrewdotjs/watchify-server/types"
)

type Cover struct {
	StatusCode int         `json:"status_code"`
	Data       types.Cover `json:"data,omitempty"`
}

func (v Cover) ToJSON() []byte {
	json, err := json.Marshal(v)

	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	return json
}
