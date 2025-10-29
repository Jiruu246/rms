package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/Jiruu246/rms/internal/config"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/migrate"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	var (
		flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	)
	flags.Usage = usage
	flags.Parse(os.Args[2:])

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	databaseURL := cfg.DatabaseURL

	if databaseURL == "" {
		log.Fatal("database URL is required (set APP_DATABASE_URL)")
	}

	// Open database connection
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create Ent driver
	drv := entsql.OpenDB("postgres", db)

	// Create Ent client
	client := ent.NewClient(ent.Driver(dialect.Debug(drv)))
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute command
	switch command {
	case "apply":
		if err := migrateUp(ctx, client); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		fmt.Println("✅ Migration completed successfully")

	case "reset":
		if err := resetDB(ctx, client, db); err != nil {
			log.Fatalf("migration rollback failed: %v", err)
		}
		fmt.Println("✅ Database reset completed successfully")

	case "create":
		if len(flags.Args()) == 0 {
			log.Fatal("migration name is required for create command")
		}
		name := flags.Args()[0]
		if err := createMigration(ctx, client, name); err != nil {
			log.Fatalf("failed to create migration: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}
}

// migrateUp applies all pending migrations
func migrateUp(ctx context.Context, client *ent.Client) error {
	return client.Schema.Create(ctx,
		migrate.WithDropColumn(true),
		migrate.WithDropIndex(true),
	)
}

// resetDB drops all tables and recreates the schema
func resetDB(ctx context.Context, client *ent.Client, db *sql.DB) error {
	// Drop all tables by dropping and recreating schema
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DROP SCHEMA IF EXISTS public CASCADE"); err != nil {
		return fmt.Errorf("failed to drop schema: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "CREATE SCHEMA public"); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Recreate schema
	return client.Schema.Create(ctx,
		migrate.WithDropColumn(true),
		migrate.WithDropIndex(true),
	)
}

// TODO: This currently has to be ran when the database is not update because it will compare the schema with the database
// createMigration generates SQL for schema changes
func createMigration(ctx context.Context, client *ent.Client, name string) error {
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("migrations/%s_%s.sql", timestamp, name)

	// Create migrations directory if it doesn't exist
	if err := os.MkdirAll("migrations", 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}
	defer file.Close()

	if err := client.Schema.WriteTo(ctx, file,
		migrate.WithDropColumn(true),
		migrate.WithDropIndex(true),
	); err != nil {
		return fmt.Errorf("failed to write migration: %w", err)
	}

	fmt.Printf("✅ Migration file created: %s\n", filename)
	return nil
}

func usage() {
	fmt.Printf(`Usage: %s <command> [options]

Commands:
  apply   		Apply all pending migrations
  reset    		Drop all tables and recreate schema (destructive!)
  create NAME  	Create a new migration file with given name
`, os.Args[0])
}
