package main

import (
	"log"
	"time"

	"github.com/addvanced/gophersocial/internal/db"
	"github.com/addvanced/gophersocial/internal/env"
	"github.com/addvanced/gophersocial/internal/store"
)

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
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

	db, err := db.NewPostgresDB(&cfg.db)
	if err != nil {
		log.Panicf("could not connect to db: %+v", err)
	}
	defer db.Close()
	log.Println("Database connection pool established")

	store := store.NewStorage(db)
	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))
}
