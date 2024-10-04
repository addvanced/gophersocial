package main

import (
	"log"
	"time"

	"github.com/addvanced/gophersocial/internal/db"
	"github.com/addvanced/gophersocial/internal/env"
	"github.com/addvanced/gophersocial/internal/store"
)

func main() {
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

	log.Printf("Connecting to database on %s", dbCfg.ConnString())

	database, err := db.NewPostgresDB(&dbCfg)
	if err != nil {
		log.Panicf("could not connect to db: %+v", err)
	}
	defer database.Close()
	log.Println("Database connection pool established")

	store := store.NewStorage(database)
	db.Seed(store)
}
