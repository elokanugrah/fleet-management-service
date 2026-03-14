package postgres

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func NewConnection(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open postgres connection: %v", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping postgres: %v", err)
	}

	log.Println("PostgreSQL connected successfully")
	return db
}
