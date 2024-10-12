package main

import (
	"context"
	"time"

	"github.com/addvanced/gophersocial/internal/db"
	"github.com/addvanced/gophersocial/internal/env"
	"github.com/addvanced/gophersocial/internal/store"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	dbCfg := db.NewPostgresConfig(
		env.GetString("DB_USER", "user"),
		env.GetString("DB_PASSWORD", "password"),
		env.GetString("DB_HOST", "localhost"),
		env.GetInt("DB_PORT", 5432),
		env.GetString("DB_NAME", "database"),
		env.GetString("DB_SSL_MODE", ""),
		env.GetInt("DB_MAX_OPEN_CONNS", 30),
		env.GetInt("DB_MAX_IDLE_CONNS", 30),
		env.GetDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
	)

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	logger.Infow("Connecting to database", "dsn", dbCfg.ConnString())

	database, err := db.NewPostgresDB(ctx, &dbCfg)
	if err != nil {
		logger.Fatalw("could not connect to db", "error", err.Error())
	}
	defer database.Close()
	logger.Infoln("Database connection pool established")

	store := store.NewStorage(database, logger)
	db.Seed(ctx, store)
}
