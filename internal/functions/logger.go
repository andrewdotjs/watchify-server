package functions

import (
	"errors"
	"log"
	"os"
	"path"
	"time"
)

func InitializeLogger() (*os.File, string) {
	var currentDate string = time.Now().Format("2006-01-02 1504")
	var executablePath string = ""
	var logPath string = ""
	var logDirectory string = ""
	var file *os.File
	var err error

	executablePath, err = os.Executable()
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	logDirectory = path.Join(executablePath, "..", "logs")

	err = os.Mkdir(logDirectory, 0777)
	if !errors.Is(err, os.ErrExist) {
		log.Fatalf("%v", err)
	}

	logPath = path.Join(logDirectory, currentDate+".log")

	file, err = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(file)
	return file, logPath
}

func RenameLogFile(file *os.File, oldPath string) {
	var currentDate string = time.Now().Format("2006-01-02 1504")
	var newPath string = path.Join(oldPath, "..", currentDate+".log")

	if err := os.Rename(oldPath, newPath); err != nil {
		log.Fatalf("ERR : %v", err)
	}

	defer file.Close()
}
