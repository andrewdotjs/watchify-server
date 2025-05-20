package functions

import (
	"database/sql"
	"fmt"
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
	var databaseDirectory string = path.Join(*appDirectory, "db", "app.db")

	log.Println("SYS : Initializing database.")
	fmt.Println("SYS : Initializing database.")

	// Open database

	log.Println("SYS : Attempting to open database.")
	log.Println("SYS : Attempting to open database.")

	database, err := sql.Open("sqlite3", databaseDirectory)
	if err != nil {
		log.Fatalf("ERR : Failed to open database. \nReason: %v", err)
		fmt.Printf("ERR : Failed to open database. \nReason: %v", err)
	} else {
		log.Println("SYS : Successfully opened the database.")
		fmt.Println("SYS : Successfully opened the database.")
	}

	// Verify connection with database.

	log.Println("SYS : Attempting to verify the connection with the database.")
	fmt.Println("SYS : Attempting to verify the connection with the database.")

	if err := database.Ping(); err != nil {
		defer database.Close()
		log.Fatalf("ERR : Verification failed. \nReason: %v", err)
		log.Printf("ERR : Verification failed. \nReason: %v", err)
	} else {
		log.Println("SYS : Connection verified.")
		fmt.Println("SYS : Connection verified.")
	}

	log.Println("SYS : Verifying database integrity.")
	fmt.Println("SYS : Verifying database integrity.")

	if _, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS series (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			episode_count INTEGER NOT NULL,
			hidden BOOLEAN NOT NULL,
			upload_date TEXT NOT NULL,
			last_modified TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS series_episodes (
		id TEXT PRIMARY KEY,
		series_id TEXT,
		episode_number INTEGER,
		title TEXT,
				description TEXT,
				file_name TEXT NOT NULL,
				file_extension TEXT NOT NULL,
				upload_date TEXT NOT NULL,
				last_modified TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS series_covers (
			id TEXT PRIMARY KEY,
			series_id TEXT NOT NULL UNIQUE,
			file_extension TEXT NOT NULL,
			file_name TEXT NOT NULL UNIQUE,
			upload_date TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS series_comments (
			id TEXT PRIMARY KEY,
			series_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			message TEXT NOT NULL,
			episodes INTEGER NOT NULL,
			likes INTEGER NOT NULL,
			dislikes INTEGER NOT NULL,
			upload_date TEXT NOT NULL,
		  last_modified TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS movies (
      id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			hidden BOOLEAN NOT NULL,
			file_extension TEXT NOT NULL,
			file_name TEXT NOT NULL,
			upload_date TEXT NOT NULL,
		  last_modified TEXT NOT NULL
    );

		CREATE TABLE IF NOT EXISTS movie_covers (
			id TEXT PRIMARY KEY,
			movie_id TEXT NOT NULL,
			user_id TEXT,
			file_extension TEXT NOT NULL,
			file_name TEXT NOT NULL,
			upload_date TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS movie_comments (
			id TEXT PRIMARY KEY,
			movie_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			message TEXT NOT NULL,
			likes INTEGER NOT NULL,
			dislikes INTEGER NOT NULL,
			upload_date TEXT NOT NULL,
			last_modified TEXT NOT NULL
		)
  `); err != nil {
		defer database.Close()
		log.Fatalf("ERR : Verification failed. \nReason: %v", err)
		fmt.Printf("ERR : Verification failed. \nReason: %v", err)
	} else {
		log.Println("SYS : Database integrity verified.")
		fmt.Println("SYS : Database integrity verified.")
	}

	log.Println("SYS : Database initialized.")
	fmt.Println("SYS : Database initialized.")

	return database
}

// Initializes the server by ensuring that the needed directories are
// present during the server's runtime. Returns the path of the
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
