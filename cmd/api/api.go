package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/addvanced/gophersocial/docs"
	"github.com/addvanced/gophersocial/internal/auth"
	"github.com/addvanced/gophersocial/internal/db"
	"github.com/addvanced/gophersocial/internal/env"
	"github.com/addvanced/gophersocial/internal/mailer"
	"github.com/addvanced/gophersocial/internal/store"
	"github.com/addvanced/gophersocial/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

var ErrInvalidParameter = errors.New("invalid URL parameter")

type ctxKey string

type application struct {
	config        config
	store         store.Storage
	cacheStorage  cache.Storage
	mailer        mailer.Client
	authenticator auth.Authenticator
	logger        *zap.SugaredLogger
}

type config struct {
	addr        string
	env         string
	apiURL      string
	frontendURL string

	mail  mailConfig
	auth  authConfig
	db    db.PostgresConfig
	redis cache.RedisConfig
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
	jwt   jwtAuthConfig
}

type basicAuthConfig struct {
	username string
	password string
}

type jwtAuthConfig struct {
	secret     string
	issuer     string
	expiration time.Duration
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
		_, _ = w.Write([]byte("nothing here..."))
	})

	r.Route("/v1", func(r chi.Router) {
		r.With(app.BasicAuthMiddleware()).
			Get("/health", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware())
			r.Post("/", app.createPostHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.addPostToCtxMiddleware)

				r.Get("/", app.getPostHandler)
				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
				r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware())

				r.Get("/", app.getUserHandler)

				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware())
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)

		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	// Docs
	docs.SwaggerInfo.Version = VERSION
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

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

func (app *application) GetIDFromURL(ctx context.Context) (int64, error) {
	return app.GetInt64URLParam(ctx, "id")
}

func (app *application) GetInt64URLParam(ctx context.Context, paramKey string) (int64, error) {
	paramStr := strings.TrimSpace(chi.URLParamFromCtx(ctx, paramKey))
	if paramStr != "" {
		if param, err := strconv.ParseInt(paramStr, 10, 64); err == nil {
			return param, nil
		}
	}
	return 0, ErrInvalidParameter
}

func (app *application) GetStringURLParam(ctx context.Context, paramKey string) (string, error) {
	param := strings.TrimSpace(chi.URLParamFromCtx(ctx, paramKey))
	if param == "" {
		return "", ErrInvalidParameter
	}
	return param, nil
}
