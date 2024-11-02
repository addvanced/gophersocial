package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/addvanced/gophersocial/internal/auth"
	"github.com/addvanced/gophersocial/internal/db"
	"github.com/addvanced/gophersocial/internal/env"
	"github.com/addvanced/gophersocial/internal/mailer"
	"github.com/addvanced/gophersocial/internal/store"
	"github.com/addvanced/gophersocial/internal/store/cache"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const VERSION = "0.0.1"

//	@title			GopherSocial API
//	@description	API for GopherSocial, a social network for gophers.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	kenneth@addvanced.dk

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@tag.name			posts
//	@tag.description	Operations related to managing posts
//
//	@tag.name			feed
//	@tag.description	Operations related to the user feed
//
//	@tag.name			ops
//	@tag.description	OPS Specific operations
//
//	@tag.name			users
//	@tag.description	Operations related to managing users

// @BasePath					/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer stop()

	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		env:         env.GetString("ENVIRONMENT", "local"),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		mail: mailConfig{
			fromName:  env.GetString("MAILER_FROM_NAME", "GopherSocial"),
			fromEmail: env.GetString("MAILER_FROM_EMAIL", "kenneth@addvanced.dk"),
			resend: resendConfig{
				fromEmail: env.GetString("RESEND_FROM_EMAIL", env.GetString("EMAIL_FROM_EMAIL", "kenneth@addvanced.dk")),
				apiKey:    env.GetString("RESEND_API_KEY", ""),
			},
			inviteExpDuration: env.GetDuration("USER_INVITE_EXPIRE", time.Hour*24*3),
		},
		auth: authConfig{
			basic: basicAuthConfig{
				username: env.GetString("BASIC_AUTH_USERNAME", "admin"),
				password: env.GetString("BASIC_AUTH_PASSWORD", "admin"),
			},
			jwt: jwtAuthConfig{
				secret:     env.GetString("JWT_TOKEN_SECRET", "example"),
				issuer:     env.GetString("JWT_TOKEN_ISSUER", "gophersocial"),
				expiration: env.GetDuration("JWT_TOKEN_EXPIRE", time.Hour*24*3),
			},
		},
		db: db.NewPostgresConfig(
			env.GetString("DB_USER", "user"),
			env.GetString("DB_PASSWORD", "password"),
			env.GetString("DB_HOST", "localhost"),
			env.GetInt("DB_PORT", 5432),
			env.GetString("DB_NAME", "database"),
			env.GetString("DB_SSL_MODE", ""),
			env.GetInt("DB_MAX_OPEN_CONNS", 30),
			env.GetInt("DB_MAX_IDLE_CONNS", 30),
			env.GetDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		),
		redis: cache.NewRedisConfig(
			env.GetBool("REDIS_ENABLED", false),
			env.GetString("REDIS_HOST", "localhost"),
			env.GetInt("REDIS_PORT", 6379),
			env.GetString("REDIS_PASSWORD", ""),
			env.GetInt("REDIS_DB", 0),
			env.GetDuration("REDIS_TTL", time.Minute),
			env.GetDuration("REDIS_TTL_USERS", env.GetDuration("REDIS_TTL", time.Minute)),
			env.GetDuration("REDIS_TTL_POSTS", env.GetDuration("REDIS_TTL", time.Minute)),
		),
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Errorw("failed to defer logger.Sync()", "error", err.Error())
		}
	}()

	// Database
	db, err := db.NewPostgresDB(ctx, &cfg.db)
	if err != nil {
		logger.Fatalw("could not connect to db", "error", err.Error())
	}
	defer db.Close()
	logger.Infoln("Database connection pool established")

	// Cache
	var rdsDB *redis.Client
	if cfg.redis.Enabled() {
		rdsDB = cache.NewRedisClient(&cfg.redis)
		logger.Infoln("Redis connection pool established")
	} else {
		logger.Warnln("Redis cache is disabled")
	}

	store := store.NewStorage(db, logger)
	cacheStore := cache.NewRedisStorage(&cfg.redis, rdsDB)

	mailer := mailer.NewResend(
		cfg.mail.fromName,
		cfg.mail.resend.fromEmail,
		cfg.mail.resend.apiKey,
	)

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.jwt.secret,
		cfg.auth.jwt.issuer,
		cfg.auth.jwt.issuer,
	)

	app := &application{
		config:        cfg,
		store:         store,
		cacheStorage:  cacheStore,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		logger:        logger,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
