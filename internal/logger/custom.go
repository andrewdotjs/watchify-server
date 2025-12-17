package logger

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

func (thisLogger *Logger) Custom(level string, location string, message string) {
  var id string = uuid.NewString()
  var currentDateTime string = time.Now().Format("2006-01-02 1504")
  var finalMessage string = fmt.Sprintf("%s %s %-5s %-20s %s", currentDateTime, id, level, location, message)

  log.Println(finalMessage)
  fmt.Println(finalMessage)

  thisLogger.Rename()
}
