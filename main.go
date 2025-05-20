package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/andrewdotjs/watchify-server/internal/functions"
	"github.com/andrewdotjs/watchify-server/internal/handlers"
	"github.com/andrewdotjs/watchify-server/internal/handlers/movies"
	movieCover "github.com/andrewdotjs/watchify-server/internal/handlers/movies/cover"
	"github.com/andrewdotjs/watchify-server/internal/handlers/series"
	seriesCover "github.com/andrewdotjs/watchify-server/internal/handlers/series/cover"
	"github.com/andrewdotjs/watchify-server/internal/middleware"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	const PORT int = 80

	// Do server initialization
	appDirectory := functions.InitializeServer()
	logFile, logPath := functions.InitializeLogger()
	database := functions.InitializeDatabase(&appDirectory)
	mux := http.NewServeMux()

	// Video collection
	mux.Handle("GET /api/v1/videos/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.ReadVideo(w, r, database)
	}))

	mux.Handle("DELETE /api/v1/videos/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteVideo(w, r, database, &appDirectory)
	}))

	mux.Handle("POST /api/v1/videos", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateVideo(w, r, database, &appDirectory)
	}))

	mux.Handle("GET /api/v1/videos", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.ReadVideo(w, r, database)
	}))

	// Stream collection
	mux.Handle("GET /api/v1/stream/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.StreamHandler(w, r, database, &appDirectory)
	}))

	// Series collection
	mux.Handle("GET /api/v1/series/{id}/episodes", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		series.Read(w, r, database)
	}))

	mux.Handle("GET /api/v1/series/{id}/cover", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seriesCover.Read(w, r, database, &appDirectory)
	}))

	mux.Handle("GET /api/v1/series/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		series.Read(w, r, database)
	}))

	mux.Handle("PUT /api/v1/series/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		series.Update(w, r, database, &appDirectory)
	}))

	mux.Handle("DELETE /api/v1/series/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		series.Delete(w, r, database, &appDirectory)
	}))

	mux.Handle("POST /api/v1/series", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		series.Create(w, r, database, &appDirectory)
	}))

	mux.Handle("GET /api/v1/series", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		series.Read(w, r, database)
	}))

	// Movies collection

	mux.Handle("GET /api/v1/movies/{id}/cover", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movieCover.Read(w, r, database, &appDirectory)
	}))

	mux.Handle("GET /api/v1/movies/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movies.Read(w, r, database)
	}))

	mux.Handle("PUT /api/v1/movies/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movies.Update(w, r, database, &appDirectory)
	}))

	mux.Handle("DELETE /api/v1/movies/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movies.Delete(w, r, database, &appDirectory)
	}))

	mux.Handle("POST /api/v1/movies", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movies.Create(w, r, database, &appDirectory)
	}))

	mux.Handle("GET /api/v1/movies", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movies.Read(w, r, database)
	}))

	// Middleware
	muxHandler := middleware.LogEndpoint(mux)
	muxHandler = middleware.CORS(muxHandler)

	server := &http.Server{
		Addr:         "0.0.0.0:" + strconv.Itoa(PORT),
		WriteTimeout: 15 * time.Minute,
		ReadTimeout:  15 * time.Minute,
		IdleTimeout:  60 * time.Second,
		Handler:      muxHandler,
	}

	// Run in goroutine to not interrupt graceful shutdown procedure.
	go func() {
		if err := server.ListenAndServe(); err == http.ErrServerClosed {
			fmt.Println("")
			log.Println("SYS : Received shutdown signal, starting shutdown procedure.")
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
	functions.RenameLogFile(logFile, logPath)
	os.Exit(0)
}
