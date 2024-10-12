package main

import (
	"context"
	"time"

	"github.com/addvanced/gophersocial/internal/db"
	"github.com/addvanced/gophersocial/internal/env"
	"github.com/addvanced/gophersocial/internal/store"
	"go.uber.org/zap"
)

const VERSION = "0.0.1"

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		env:  env.GetString("ENVIRONMENT", "local"),

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
	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
