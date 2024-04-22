package utilities

import (
	"errors"
	"log"
	"os"
	"path"
	"time"
)

func InitializeLogger() (*os.File, string) {
	var currentDate string = time.Now().Format("2006-01-02 15:04:05")

	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	logDirectory := path.Join(executablePath, "..", "logs")

	if err := os.Mkdir(logDirectory, 0666); err != nil && !errors.Is(err, os.ErrExist) {
		log.Fatalf("%v", err)
	}

	logPath := path.Join(logDirectory, currentDate+".log")

	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(file)
	return file, logPath
}

func RenameLogFile(file *os.File, oldPath string) {
	var currentDate string = time.Now().Format("2006-01-02 15:04:05")
	var newPath string = path.Join(oldPath, "..", currentDate+".log")

	if err := os.Rename(oldPath, newPath); err != nil {
		log.Fatalf("ERR : %v", err)
	}

	defer file.Close()
}
