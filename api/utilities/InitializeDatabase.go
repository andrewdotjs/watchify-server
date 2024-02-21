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

	if err := database.Ping(); err != nil {
		log.Fatalf("ERR : %v", err)
	}

	statement, err := database.Prepare(`
    CREATE TABLE IF NOT EXISTS videos (
      id TEXT PRIMARY KEY,
      series_id TEXT,
      episode_number INTEGER,
      title TEXT,
			file_name TEXT NOT NULL,
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
