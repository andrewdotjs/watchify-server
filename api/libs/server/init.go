package server

import (
	"database/sql"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
)

// Initializes the database by ensuring that the database file and
// needed tables are all present and ready to be used during the server's
// runtime. Returns the database as a pointer to an sql.DB struct.
func InitializeDatabase(appDirectory *string) *sql.DB {
	databaseDirectory := path.Join(*appDirectory, "db", "app.db")

	database, err := sql.Open("sqlite3", databaseDirectory)
	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	if err := database.Ping(); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	if _, err := database.Exec(`
    CREATE TABLE IF NOT EXISTS videos (
      id TEXT PRIMARY KEY,
      series_id TEXT,
      episode INTEGER,
      title TEXT NOT NULL,
			file_extension TEXT NOT NULL,
			upload_date TEXT NOT NULL,
			last_modified TEXT NOT NULL
    );
  `); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	if _, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS covers (
			id TEXT PRIMARY KEY,
			series_id TEXT NOT NULL UNIQUE,
			file_name TEXT NOT NULL UNIQUE,
			upload_date TEXT NOT NULL
		);
	`); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	if _, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS series (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			episodes INTEGER NOT NULL,
			upload_date TEXT NOT NULL,
		  last_modified TEXT NOT NULL
		);
	`); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	if _, err := database.Exec(`
    CREATE TABLE IF NOT EXISTS movies (
      id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			upload_date TEXT NOT NULL,
		  last_modified TEXT NOT NULL
    );
  `); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	return database
}

// Initializes the server by ensuring that the needed directories are
// present during the server's runtime. returns the path of the
// running executable's directory.
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
