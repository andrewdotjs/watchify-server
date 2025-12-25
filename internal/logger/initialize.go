package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

func (thisLogger *Logger) Initialize() {
	var currentDate string = time.Now().Format("2006-01-02 150405")
	var executablePath string = ""
	var logPath string = ""
	var logDirectory string = ""
	var file *os.File
	var err error
	var header string = fmt.Sprintf("%-19s %-36s %-5s %s \n", "Datetime", "ID", "Level", "Message")

	executablePath, err = os.Executable()
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	logDirectory = path.Join(executablePath, "..", "logs")

	err = os.Mkdir(logDirectory, 0744)
	if !errors.Is(err, os.ErrExist) {
		log.Fatalf("%v", err)
	}

	logPath = path.Join(logDirectory, currentDate+".log")

	file, err = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}


	thisLogger.fileObject = file
	thisLogger.path = logPath

	if _, err := file.Write([]byte(header)); err != nil {
	  return
	}

	log.SetOutput(file)
}
