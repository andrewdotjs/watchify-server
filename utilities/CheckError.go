package utilities

import (
	"log"
)

func CheckError(err error) {
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}
}
