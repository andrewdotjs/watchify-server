package logger

import (
	"log"
	"os"
	"path"
	"time"
)

func (thisLogger *Logger) Rename() {
 	var currentDate string = time.Now().Format("2006-01-02 150405")
	var newPath string = path.Join(thisLogger.path, "..", currentDate+".log")

	if err := os.Rename(thisLogger.path, newPath); err != nil {
		log.Fatalf("ERR : %v", err)
	}

	file, err := os.OpenFile(newPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	thisLogger.path = newPath
	thisLogger.fileObject = file

	defer thisLogger.fileObject.Close()
}
