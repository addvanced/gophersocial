package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/addvanced/gophersocial/docs"
	"github.com/addvanced/gophersocial/internal/db"
	"github.com/addvanced/gophersocial/internal/env"
	"github.com/addvanced/gophersocial/internal/mailer"
	"github.com/addvanced/gophersocial/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type ctxKey string

type application struct {
	config config
	store  store.Storage
	mailer mailer.Client
	logger *zap.SugaredLogger
}

type config struct {
	addr        string
	env         string
	apiURL      string
	frontendURL string

	mail mailConfig
	auth authConfig
	db   db.PostgresConfig
}

type mailConfig struct {
	fromName  string
	fromEmail string

	resend resendConfig

	inviteExpDuration time.Duration
}

type resendConfig struct {
	apiKey    string
	fromEmail string
}

type authConfig struct {
	basic basicAuthConfig
}

type basicAuthConfig struct {
	username string
	password string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.Timeout(env.GetDuration("TIMEOUT_IDLE", time.Minute)))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("nothing here..."))
	})

	r.Route("/v1", func(r chi.Router) {
		r.With(app.BasicAuthMiddleware()).
			Get("/health", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)

			r.Route("/{postId}", func(r chi.Router) {
				r.Use(app.addPostToCtxMiddleware)

				r.Get("/", app.getPostHandler)
				r.Patch("/", app.updatePostHandler)
				r.Delete("/", app.deletePostHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userId}", func(r chi.Router) {
				r.Use(app.addUserToCtxMiddleware)

				r.Get("/", app.getUserHandler)

				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)

		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	// Docs
	docs.SwaggerInfo.Version = VERSION
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: env.GetDuration("TIMEOUT_WRITE", 30*time.Second),
		ReadTimeout:  env.GetDuration("TIMEOUT_READ", 10*time.Second),
		IdleTimeout:  env.GetDuration("TIMEOUT_IDLE", time.Minute),
	}

	app.logger.Infow("server has started", "addr", app.config.addr)
	return srv.ListenAndServe()
}
