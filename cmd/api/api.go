package main

import (
	"log"
	"net/http"
	"time"

	"github.com/addvanced/gophersocial/internal/db"
	"github.com/addvanced/gophersocial/internal/env"
	"github.com/addvanced/gophersocial/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ctxKey string

type application struct {
	config config
	store  store.Storage
}

type config struct {
	addr string
	env  string

	db db.PostgresConfig
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(env.GetDuration("TIMEOUT_IDLE", time.Minute)))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("nothing here..."))
	})

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)

			r.Route("/{postId}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Patch("/", app.updatePostHandler)
				r.Delete("/", app.deletePostHandler)
			})
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: env.GetDuration("TIMEOUT_WRITE", 30*time.Second),
		ReadTimeout:  env.GetDuration("TIMEOUT_READ", 10*time.Second),
		IdleTimeout:  env.GetDuration("TIMEOUT_IDLE", time.Minute),
	}

	log.Printf("server has started on %s", app.config.addr)
	return srv.ListenAndServe()
}
