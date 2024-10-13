package main

import (
	"context"
	"time"

	"github.com/addvanced/gophersocial/internal/db"
	"github.com/addvanced/gophersocial/internal/env"
	"github.com/addvanced/gophersocial/internal/mailer"
	"github.com/addvanced/gophersocial/internal/store"
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

//	@BasePath	/v1

//	@securityDifinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				JWT Authorization header

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

func main() {
	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		env:         env.GetString("ENVIRONMENT", "local"),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		mail: mailConfig{
			fromName:  env.GetString("MAILER_FROM_NAME", "GopherSocial"),
			fromEmail: env.GetString("MAILER_FROM_EMAIL", "kenneth@addvanced.dk"),
			resend: resendConfig{
				fromEmail: env.GetString("RESEND_FROM_EMAIL", env.GetString("EMAIL_FROM_EMAIL", "kenneth@addvanced.dk")),
				apiKey:    env.GetString("RESEND_API_KEY", ""),
			},
			inviteExpDuration: env.GetDuration("USER_INVITE_EXPIRE", time.Hour*24*3),
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
	}

	ctx := context.Background()

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.NewPostgresDB(ctx, &cfg.db)
	if err != nil {
		logger.Fatalw("could not connect to db", "error", err.Error())
	}
	defer db.Close()
	logger.Infoln("Database connection pool established")

	store := store.NewStorage(db, logger)

	mailer := mailer.NewResend(cfg.mail.fromName, cfg.mail.resend.fromEmail, cfg.mail.resend.apiKey)

	app := &application{
		config: cfg,
		store:  store,
		mailer: mailer,
		logger: logger,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
