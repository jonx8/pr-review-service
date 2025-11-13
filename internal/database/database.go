package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jonx8/pr-review-service/internal/config"
	_ "github.com/lib/pq"
)

func InitDB(cfg config.DBConfig) (*sql.DB, error) {
	log.Printf("Connecting to database: %s:%s/%s", cfg.Host, cfg.Port, cfg.Database)

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	return db, nil
}
