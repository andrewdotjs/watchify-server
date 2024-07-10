package database

import (
	"database/sql"
	"log"
	"path"
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

	_, err = database.Exec(`
    CREATE TABLE IF NOT EXISTS videos (
      id TEXT PRIMARY KEY,
      series_id TEXT,
      episode INTEGER,
      title TEXT NOT NULL,
			file_extension TEXT NOT NULL,
			upload_date TEXT NOT NULL,
			last_modified TEXT NOT NULL
    );
  `)
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS covers (
			id TEXT PRIMARY KEY,
			series_id TEXT NOT NULL UNIQUE,
			file_name TEXT NOT NULL UNIQUE,
			upload_date TEXT NOT NULL
		);
	`)
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS series (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			episodes INTEGER NOT NULL,
			upload_date TEXT NOT NULL,
		  last_modified TEXT NOT NULL
		);
	`)
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	_, err = database.Exec(`
    CREATE TABLE IF NOT EXISTS movies (
      id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			upload_date TEXT NOT NULL,
		  last_modified TEXT NOT NULL
    );
  `)
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	return database
}
