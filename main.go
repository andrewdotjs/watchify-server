package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/andrewdotjs/watchify-server/internal/database"
	"github.com/andrewdotjs/watchify-server/internal/handlers"
	"github.com/andrewdotjs/watchify-server/internal/handlers/covers"
	"github.com/andrewdotjs/watchify-server/internal/handlers/movies"
	"github.com/andrewdotjs/watchify-server/internal/handlers/shows"
	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/middleware"
	"github.com/andrewdotjs/watchify-server/internal/server"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
  const PORT int = 80
  var log logger.Logger = logger.Logger{}

	// Initialize logger
	log.Initialize()
	log.Info("INIT", "Logger initialized")

	// Initialization
	appDirectory := server.Initialize()

	// Database initialization
	db := database.Initialize(&log, &appDirectory)
	log.Info("INIT", "Database initialized")

	mux := http.NewServeMux()


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
			log.Info("SHUTDOWN", "Starting shutdown procedure")
		} else if err != nil {
			fmt.Println("")
			log.Error("SHUTDOWN", fmt.Sprintf("%v", err))
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	defer db.Close()
	server.Shutdown(context.Background())

	log.Info("SHUTDOWN", "Shutting down...")
	log.Rename()
	os.Exit(0)
}
