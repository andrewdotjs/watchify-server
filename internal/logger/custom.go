package logger

import (
	"fmt"
	"log"
	"time"
)

func (thisLogger *Logger) Custom(level string, transactionId *string, message *string) {
  var currentDateTime string = time.Now().Format("2006/01/02 15:04:05")
  var finalMessage string = fmt.Sprintf("%s %s %-5s %s", currentDateTime, *transactionId, level, *message)

  log.Println(finalMessage)
  fmt.Println(finalMessage)

  thisLogger.Rename()
}
