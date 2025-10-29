package database

import (
	"fmt"

	"github.com/Jiruu246/rms/internal/ent"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewEntClient creates a new Ent client with the given DSN
func NewEntClient(dsn string) (*ent.Client, error) {
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return client, nil
}
