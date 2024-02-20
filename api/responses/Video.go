package responses

import (
	"encoding/json"
	"log"

	"github.com/andrewdotjs/watchify-server/types"
)

type Video struct {
	StatusCode int         `json:"status_code"`
	Data       types.Video `json:"data"`
}

func (v Video) ToJSON() []byte {
	json, err := json.Marshal(v)

	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	return json
}
