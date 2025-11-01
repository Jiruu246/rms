package database

import (
	"fmt"

	"entgo.io/ent/dialect"
	"github.com/Jiruu246/rms/internal/ent"
	_ "github.com/lib/pq"
)

// NewEntClient creates a new Ent client with the given DSN
func NewEntClient(dsn string) (*ent.Client, error) {
	client, err := ent.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return client, nil
}
