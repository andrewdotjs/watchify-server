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
	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/middleware"
	"github.com/andrewdotjs/watchify-server/internal/server"
	"github.com/google/uuid"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
  const PORT int = 80
  var functionId string = uuid.NewString()

  var log logger.Logger = logger.Logger{}

	// Initialize logger
	log.Initialize()
	log.Info(functionId, "Logger initialized")

	// Initialization
	appDirectory := server.Initialize()

	// Database initialization
	db := database.Initialize(&log, &appDirectory)
	log.Info(functionId, "Database initialized")

	mux := http.NewServeMux()

	handlers.Shows(mux, db, &appDirectory, &log)
	handlers.Movies(mux, db, &appDirectory, &log)
	handlers.Stream(mux, db, &appDirectory, &log)
	handlers.Videos(mux, db, &appDirectory, &log)

	// Middleware
	muxHandler := middleware.LogEndpoint(mux, &log ,&functionId)
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
			log.Info(functionId, "Starting shutdown procedure")
		} else if err != nil {
			fmt.Println("")
			log.Error(functionId, fmt.Sprintf("%v", err))
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	defer db.Close()
	server.Shutdown(context.Background())

	log.Info(functionId, "Shutting down...")
	os.Exit(0)
}
