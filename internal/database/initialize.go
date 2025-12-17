package database

import (
	"database/sql"
	"fmt"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/logger"
)

// Initializes the database by ensuring that the database file and
// needed tables are all present and ready to be used during the server's
// runtime. Returns the database as a pointer to an sql.DB struct.
func Initialize(log *logger.Logger, appDirectory *string) *sql.DB {
	var databaseDirectory string = path.Join(*appDirectory, "db", "app.db")

	log.Info("INIT", "Starting database initialization")

	// Open database

	log.Info("INIT", "Opening database")

	database, err := sql.Open("sqlite3", databaseDirectory)
	if err != nil {
		log.Fatal("INIT", fmt.Sprintf("%v", err))
	} else {
	  log.Info("INIT", "Successfully opened the database")
	}

	// Verify connection with database.

	log.Info("INIT", "Verifying the connection with the database")

	if err := database.Ping(); err != nil {
		defer database.Close()
		log.Fatal("INIT", fmt.Sprintf("Verification failed. Reason: %v", err))
	} else {
    log.Info("INIT", "Connection verified")
	}

	log.Info("INIT", "Verifying database integrity")

	if _, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS shows (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			episode_count INTEGER NOT NULL,
			hidden BOOLEAN NOT NULL,
			upload_date TEXT NOT NULL,
			last_modified TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS episodes (
		  id TEXT PRIMARY KEY,
		  parent_id TEXT,
		  episode_number INTEGER,
		  title TEXT,
			description TEXT,
			file_name TEXT NOT NULL,
			file_extension TEXT NOT NULL,
			upload_date TEXT NOT NULL,
			last_modified TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS covers (
			id TEXT PRIMARY KEY,
			parent_id TEXT NOT NULL UNIQUE,
			file_extension TEXT NOT NULL,
			file_name TEXT NOT NULL UNIQUE,
			upload_date TEXT NOT NULL
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
  `); err != nil {
		defer database.Close()
		log.Fatal("INIT", fmt.Sprintf("Verification failed. Reason: %v", err))
	} else {
	  log.Info("INIT", "Database integrity verified")
	}

	log.Info("INIT", "Database is ready")
	return database
}
