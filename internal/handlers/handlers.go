package handlers

import (
	"database/sql"
	"net/http"

	"github.com/andrewdotjs/watchify-server/internal/handlers/covers"
	"github.com/andrewdotjs/watchify-server/internal/handlers/episodes"
	"github.com/andrewdotjs/watchify-server/internal/handlers/movies"
	"github.com/andrewdotjs/watchify-server/internal/handlers/shows"
	"github.com/andrewdotjs/watchify-server/internal/handlers/stream"
	"github.com/andrewdotjs/watchify-server/internal/handlers/videos"
	"github.com/andrewdotjs/watchify-server/internal/logger"
)

// Stream

func Stream(
  mux *http.ServeMux,
  db *sql.DB,
  appDirectory *string,
  log *logger.Logger,
) {
 	mux.Handle("GET /api/v1/stream/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream.Read(w, r, db, appDirectory)
	}))
}

// Videos

func Videos(
  mux *http.ServeMux,
  db *sql.DB,
  appDirectory *string,
  log *logger.Logger,
) {
 	mux.Handle("GET /api/v1/videos/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		videos.Read(w, r, db)
	}))

	mux.Handle("DELETE /api/v1/videos/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		videos.Delete(w, r, db, appDirectory, log)
	}))

	mux.Handle("POST /api/v1/videos", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		videos.Create(w, r, db, appDirectory, log)
	}))

	mux.Handle("GET /api/v1/videos", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		videos.Read(w, r, db)
	}))
}

// Shows

func Shows(
  mux *http.ServeMux,
  db *sql.DB,
  appDirectory *string,
  log *logger.Logger,
) {
 	mux.Handle("GET /api/v1/shows/{id}/episodes", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		episodes.Read(w, r, db, log)
	}))

  mux.Handle("GET /api/v1/shows/{id}/cover", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    covers.Read(w, r, db, appDirectory, log)
  }))

  mux.Handle("PUT /api/v1/shows/{id}/cover", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    covers.Update(w, r, db, appDirectory, log)
  }))

  mux.Handle("DELETE /api/v1/shows/{id}/cover", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    covers.Delete(w, r, db, appDirectory, log)
  }))

	mux.Handle("GET /api/v1/shows/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shows.Read(w, r, db, log)
	}))

	mux.Handle("PUT /api/v1/shows/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shows.Update(w, r, db, appDirectory, log)
	}))

	mux.Handle("DELETE /api/v1/shows/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shows.Delete(w, r, db, appDirectory, log)
	}))

	mux.Handle("POST /api/v1/shows", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shows.Create(w, r, db, appDirectory, log)
	}))

	mux.Handle("GET /api/v1/shows", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shows.Read(w, r, db, log)
	}))
}

// Movies

func Movies(
  mux *http.ServeMux,
  db *sql.DB,
  appDirectory *string,
  log *logger.Logger,
) {
  mux.Handle("GET /api/v1/movies/{id}/cover", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    covers.Read(w, r, db, appDirectory, log)
  }))

  mux.Handle("PUT /api/v1/movies/{id}/cover", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    covers.Update(w, r, db, appDirectory, log)
  }))

  mux.Handle("DELETE /api/v1/movies/{id}/cover", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    covers.Delete(w, r, db, appDirectory, log)
  }))

 	mux.Handle("GET /api/v1/movies/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movies.Read(w, r, db, log)
	}))

	mux.Handle("PUT /api/v1/movies/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movies.Update(w, r, db, appDirectory, log)
	}))

	mux.Handle("DELETE /api/v1/movies/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movies.Delete(w, r, db, appDirectory, log)
	}))

	mux.Handle("POST /api/v1/movies", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movies.Create(w, r, db, appDirectory, log)
	}))

	mux.Handle("GET /api/v1/movies", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movies.Read(w, r, db, log)
	}))


}
