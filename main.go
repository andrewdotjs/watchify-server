package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/andrewdotjs/watchify-server/api/handlers"
	"github.com/andrewdotjs/watchify-server/api/middleware"
	"github.com/andrewdotjs/watchify-server/api/utilities"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	const PORT int = 8080
	var database *sql.DB
	var router *mux.Router
	var appDirectory string

	// Do server initialization
	appDirectory = utilities.InitializeServer()
	database = utilities.InitializeDatabase()
	router = mux.NewRouter()

	// Video collection
	router.HandleFunc("/api/v1/videos/stream", func(w http.ResponseWriter, r *http.Request) {
		handlers.StreamHandler(w, r, database, &appDirectory)
	}).Methods("GET")

	router.HandleFunc("/api/v1/videos/upload", func(w http.ResponseWriter, r *http.Request) {
		handlers.PostVideoHandler(w, r, database, &appDirectory)
	}).Methods("POST")

	router.HandleFunc("/api/v1/videos/delete", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteVideoHandler(w, r, database)
	}).Methods("DELETE")

	router.HandleFunc("/api/v1/videos", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetVideoHandler(w, r, database)
	}).Methods("GET")

	// Cover collection
	router.HandleFunc("/api/v1/covers/upload", func(w http.ResponseWriter, r *http.Request) {
		handlers.PostCoverHandler(w, r, database, &appDirectory)
	}).Methods("POST")

	router.HandleFunc("/api/v1/covers/delete", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteCoverHandler(w, r, database, &appDirectory)
	}).Methods("DELETE")

	router.HandleFunc("/api/v1/covers", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetCoverHandler(w, r, database, &appDirectory)
	}).Methods("GET")

	// Middleware
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
		if err := server.ListenAndServe(); err == http.ErrServerClosed {
			fmt.Println("")
			log.Println("SYS : Recieved shutdown signal, starting shutdown procedure.")
		} else if err != nil {
			fmt.Println("")
			log.Printf("ERR : %v", err)
		}
	}()

	log.Printf("SYS : Listening on port %v", PORT)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	defer database.Close()
	server.Shutdown(context.Background())
	log.Println("SYS : Shutting down...")
	os.Exit(0)
}
