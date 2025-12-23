package database

import (
	"database/sql"
	"fmt"
	"path"

	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/google/uuid"
)

// Initializes the database by ensuring that the database file and
// needed tables are all present and ready to be used during the server's
// runtime. Returns the database as a pointer to an sql.DB struct.
func Initialize(log *logger.Logger, appDirectory *string) *sql.DB {
  var sequenceId string = uuid.NewString()
	var databaseDirectory string = path.Join(*appDirectory, "db", "app.db")

	log.Info(sequenceId, "Starting database initialization")

	// Open database

	log.Info(sequenceId, "Opening database")

	database, err := sql.Open("sqlite3", databaseDirectory)
	if err != nil {
		log.Fatal(sequenceId, fmt.Sprintf("%v", err))
	} else {
	  log.Info(sequenceId, "Successfully opened the database")
	}

	// Verify connection with database.

	log.Info(sequenceId, "Verifying the connection with the database")

	if err := database.Ping(); err != nil {
		defer database.Close()
		log.Fatal(sequenceId, fmt.Sprintf("Verification failed. Reason: %v", err))
	} else {
    log.Info(sequenceId, "Connection verified")
	}

	log.Info(sequenceId, "Verifying database integrity")

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
		log.Fatal(sequenceId, fmt.Sprintf("Verification failed. Reason: %v", err))
	} else {
	  log.Info(sequenceId, "Database integrity verified")
	}

	log.Info(sequenceId, "Database is ready")
	return database
}
