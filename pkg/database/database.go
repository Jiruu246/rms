package database

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// New returns an initialized sqlx.DB. The caller should Close() it.
func New(dsn string) (*sqlx.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty database dsn")
	}

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	// production tuning could be added here (SetMaxOpenConns, SetConnMaxLifetime, etc.)
	return db, nil
}
