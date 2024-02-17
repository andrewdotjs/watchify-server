package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/andrewdotjs/watchify-server/api/handlers"
	"github.com/andrewdotjs/watchify-server/api/middleware"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	const PORT int = 8080
	database, err := sql.Open("sqlite3", "./db/videos.db")

	if err != nil {
		log.Fatal(err)
	}

	// Ensure that the needed tables are ready
	statement, err := database.Prepare(`
    CREATE TABLE IF NOT EXISTS videos (
      id VARCHAR(50) PRIMARY KEY,
      series_id VARCHAR(50),
      episode_number INTEGER,
      title VARCHAR(50)
    );
  `)

	if err != nil {
		log.Fatal(err)
	}

	statement.Exec()
	defer database.Close()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/info", handlers.GetInfoHandler).Methods("GET")
	router.Use(middleware.EndpointLogger)

	server := &http.Server{
		Addr:         "0.0.0.0:" + strconv.Itoa(PORT),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      router,
	}

	// Run in goroutine to not interrupt graceful shutdown procedure.
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	log.Printf("SYS : Listening on port %v", PORT)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	server.Shutdown(context.Background())
	log.Println("SYS : Shutting down...")
	os.Exit(0)
}
