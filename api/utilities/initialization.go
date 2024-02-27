package utilities

import (
	"database/sql"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
)

func InitializeDatabase() *sql.DB {
	database, err := sql.Open("sqlite3", "./db/videos.db")
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	if err := database.Ping(); err != nil {
		log.Fatalf("ERR : %v", err)
	}

	statement, err := database.Prepare(`
    CREATE TABLE IF NOT EXISTS videos (
      id TEXT PRIMARY KEY,
      series_id TEXT,
      episode INTEGER,
      title TEXT,
			file_name TEXT NOT NULL,
			file_extension TEXT NOT NULL,
			upload_date TEXT NOT NULL
    );
  `)
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	statement.Exec()

	statement, err = database.Prepare(`
		CREATE TABLE IF NOT EXISTS covers (
			id TEXT PRIMARY KEY,
			series_id TEXT NOT NULL UNIQUE,
			file_name TEXT NOT NULL UNIQUE,
			upload_date TEXT NOT NULL
		);
	`)
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	statement.Exec()
	return database
}

func InitializeServer() string {
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
			log.Printf("SYS : creating %v folder", value)
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
