package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) *sql.DB {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	log.Println("Connected to PostgreSQL successfully")
	return db
}

func Migrate(db *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id            SERIAL PRIMARY KEY,
			email         TEXT UNIQUE NOT NULL,
			username      TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at    TIMESTAMP DEFAULT NOW(),
			updated_at    TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS series (
			id          SERIAL PRIMARY KEY,
			user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title       TEXT NOT NULL,
			genre       TEXT,
			status      TEXT DEFAULT 'plan_to_watch'
			            CHECK (status IN ('watching','completed','dropped','plan_to_watch')),
			rating      INTEGER CHECK (rating >= 1 AND rating <= 10),
			cover_url   TEXT,
			description TEXT,
			episodes    INTEGER,
			created_at  TIMESTAMP DEFAULT NOW(),
			updated_at  TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_series_user_id ON series(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_series_title   ON series(title)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Fatalf("Migration error: %v\nQuery: %s", err, q)
		}
	}

	log.Println("Database migrations applied successfully")
}