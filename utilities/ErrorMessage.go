package utilities

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/andrewdotjs/watchify-server/types"
)

func ErrorMessage(statusCode int, message string) []byte {
	response, err := json.Marshal(types.Message{
		StatusCode: statusCode,
		Message:    message,
	})

	if err != nil {
		errorMessage := fmt.Sprintf("an error has occured. %v", err)
		log.Printf("ERR : %v", err)
		return []byte(errorMessage)
	}

	return response
}
