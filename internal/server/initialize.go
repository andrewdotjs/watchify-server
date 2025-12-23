package server

import (
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
)

// Initializes the server by ensuring that the needed directories are
// present during the server's runtime. Returns the path of the
// running executable's directory.
func Initialize() (string) {
	permissions := fs.FileMode(0770) // Linux octal permissions

	executable, err := os.Executable()
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	appDirectory := filepath.Dir(executable)
	checkDirectories := []string{"db", "storage"}
	subStorage := []string{"covers", "videos"}

	for _, value := range checkDirectories {
		directory := path.Join(appDirectory, value)
		_, err = os.ReadDir(directory)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatalf("ERR : %v", err)
			}
			log.Printf("SYS : No %v folder detected. Creating %v folder", value, value)
			if err = os.Mkdir(directory, permissions); err != nil {
				log.Fatalf("ERR : %v", err)
			}
		}
	}

	for _, value := range subStorage {
		directory := path.Join(appDirectory, "storage", value)
		_, err = os.ReadDir(directory)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatalf("ERR : %v", err)
			}
			log.Printf("SYS : creating %v folder", value)
			if err = os.Mkdir(directory, permissions); err != nil {
				log.Fatalf("ERR : %v", err)
			}
		}
	}

	return appDirectory
}
