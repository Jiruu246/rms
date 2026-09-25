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
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	var (
		flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	)
	flags.Usage = usage
	if err := flags.Parse(os.Args[2:]); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
	}

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
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	// Create Ent driver
	drv := entsql.OpenDB("postgres", db)

	// Create Ent client
	client := ent.NewClient(ent.Driver(dialect.Debug(drv)))
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close ent client: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute command
	switch command {
	case "apply":
		if err := migrateUp(ctx, client, db); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		fmt.Println("✅ Migration completed successfully")

	case "reset":
		if err := resetDB(ctx, client, db); err != nil {
			log.Fatalf("migration rollback failed: %v", err)
		}
		fmt.Println("✅ Database reset completed successfully")

	case "seed":
		if err := seedDatabase(ctx, client); err != nil {
			log.Fatalf("seeding failed: %v", err)
		}
		fmt.Println("✅ Database seeding completed successfully")

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
func migrateUp(ctx context.Context, client *ent.Client, db *sql.DB) error {
	// TODO: one-off fixup for the user_id -> owner_id rename. Remove
	// this call and the backfillRestaurantOwnerID function once every
	// environment (including any stale branch/staging databases) has run
	// apply successfully past this point — see ticket [link] for the
	// longer-term plan to stop this kind of fixup accumulating in apply.
	if err := backfillRestaurantOwnerID(ctx, db); err != nil {
		return fmt.Errorf("failed to backfill restaurants.owner_id: %w", err)
	}
	return client.Schema.Create(ctx,
		migrate.WithDropColumn(true),
		migrate.WithDropIndex(true),
	)
}

// backfillRestaurantOwnerID copies the legacy restaurants.user_id column into
// owner_id before ent's auto-migration drops user_id and adds owner_id as
// NOT NULL — otherwise that ALTER TABLE fails on any pre-existing row
// (restaurants.owner_id was renamed from user_id).
// It is a no-op once user_id no longer exists, so it is safe to run on every
// apply.
func backfillRestaurantOwnerID(ctx context.Context, db *sql.DB) error {
	var userIDExists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name = 'restaurants' AND column_name = 'user_id'
	)`).Scan(&userIDExists); err != nil {
		return fmt.Errorf("failed to check for restaurants.user_id: %w", err)
	}
	if !userIDExists {
		return nil
	}

	var ownerIDExists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name = 'restaurants' AND column_name = 'owner_id'
	)`).Scan(&ownerIDExists); err != nil {
		return fmt.Errorf("failed to check for restaurants.owner_id: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin backfill transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback backfill transaction: %v", err)
		}
	}()

	if !ownerIDExists {
		if _, err := tx.ExecContext(ctx, `ALTER TABLE restaurants ADD COLUMN owner_id uuid`); err != nil {
			return fmt.Errorf("failed to add owner_id column: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE restaurants SET owner_id = user_id WHERE owner_id IS NULL`); err != nil {
		return fmt.Errorf("failed to backfill owner_id from user_id: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `ALTER TABLE restaurants ALTER COLUMN owner_id SET NOT NULL`); err != nil {
		return fmt.Errorf("failed to enforce owner_id NOT NULL: %w", err)
	}

	return tx.Commit()
}

// resetDB drops all tables and recreates the schema
func resetDB(ctx context.Context, client *ent.Client, db *sql.DB) error {
	// Drop all tables by dropping and recreating schema
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

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

	if err := client.Schema.WriteTo(ctx, file,
		migrate.WithDropColumn(true),
		migrate.WithDropIndex(true),
	); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return fmt.Errorf("failed to write migration: %w (also failed to close file: %v)", err, closeErr)
		}
		return fmt.Errorf("failed to write migration: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close migration file: %w", err)
	}

	fmt.Printf("✅ Migration file created: %s\n", filename)
	return nil
}

func usage() {
	fmt.Printf(`Usage: %s <command> [options]

Commands:
  apply   		Apply all pending migrations
  reset    		Drop all tables and recreate schema (destructive!)
  seed    		Populate database with initial sample data
  create NAME  	Create a new migration file with given name
`, os.Args[0])
}
