package utilities

import (
	"database/sql"
	"log"
)

func InitializeDatabase() *sql.DB {
	database, err := sql.Open("sqlite3", "./db/videos.db")

	if err != nil {
		log.Fatalf("ERR : %v", err)
	}

	// Ensure that the needed tables are ready
	statement1, err1 := database.Prepare(`
    CREATE TABLE IF NOT EXISTS videos (
      id TEXT PRIMARY KEY,
      series_id TEXT,
      episode_number INTEGER,
      title TEXT,
			file_name TEXT NOT NULL,
			upload_date TEXT NOT NULL 
    );
  `)

	statement2, err2 := database.Prepare(`
		CREATE TABLE IF NOT EXISTS covers (
			id TEXT PRIMARY KEY,
			series_id TEXT NOT NULL UNIQUE,
			file_name TEXT NOT NULL UNIQUE,
			upload_date TEXT NOT NULL
		);
	`)

	switch {
	case err1 != nil:
		defer database.Close()
		log.Fatalf("ERR : %v", err1)
	case err2 != nil:
		defer database.Close()
		log.Fatalf("ERR : %v", err2)
	}

	statement1.Exec()
	statement2.Exec()

	return database
}
