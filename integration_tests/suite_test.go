package integration_tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Jiruu246/rms/internal/config"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/server"
	"github.com/Jiruu246/rms/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite defines the test suite structure
type IntegrationTestSuite struct {
	suite.Suite
	// TODO: Is this needed?
	cfg        *config.Config
	client     *ent.Client
	adminConn  *pgx.Conn
	server     *server.Server
	testDBName string
}

// SetupSuite runs once before the entire test suite
func (s *IntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	root, _ := os.Getwd()
	log.Printf("Current working directory: %s", root)
	s.Require().NoError(godotenv.Load(filepath.Join(root, ".env.test")), "Failed to load .env file")
	cfg, err := config.Load()
	s.Require().NoError(err, "Failed to load config")
	s.cfg = cfg

	// Create test database
	client, err := s.createTestDatabase(ctx, cfg)
	s.Require().NoError(err, "Failed to create test database")
	s.client = client

	// Run migrations on test database
	err = s.runMigrations(ctx)
	s.Require().NoError(err, "Failed to run migrations")

	// Create server instance with mock middlewares
	s.server = server.New(s.cfg, s.client, server.Middlewares{
		RestrictiveCORS: func(origins []string) gin.HandlerFunc {
			return func(c *gin.Context) {
				c.Next()
			}
		},
		CORS: func() gin.HandlerFunc {
			return func(c *gin.Context) {
				c.Next()
			}
		},
		JWTMiddleware: func(secretKey []byte) gin.HandlerFunc {
			return func(c *gin.Context) {
				c.Next()
			}
		},
	})

	log.Printf("Integration test suite setup completed with database: %s", s.testDBName)
}

// TearDownSuite runs once after the entire test suite
func (s *IntegrationTestSuite) TearDownSuite() {
	ctx := context.Background()

	// Close client connection
	if s.client != nil {
		err := s.client.Close()
		if err != nil {
			log.Printf("Failed to close database client: %v", err)
		}
	}

	// Drop test database
	err := s.dropTestDatabase(ctx)
	if err != nil {
		log.Printf("Failed to drop test database: %v", err)
	}

	// Close admin connection
	if s.adminConn != nil {
		err := s.adminConn.Close(ctx)
		if err != nil {
			log.Printf("Failed to close admin connection: %v", err)
		}
	}

	log.Printf("Integration test suite cleanup completed")
}

// SetupTest runs before each individual test
func (s *IntegrationTestSuite) SetupTest() {
	// Clean up data between tests while keeping schema
	s.cleanupTestData()
}

// createTestDatabase creates a new database for testing
func (s *IntegrationTestSuite) createTestDatabase(ctx context.Context, cfg *config.Config) (*ent.Client, error) {
	// Connect to PostgreSQL server (not to a specific database)
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL server: %v", err)
		return nil, err
	}
	s.adminConn = conn

	s.testDBName = fmt.Sprintf("rms_test_%d", time.Now().Unix())

	// Create test database
	createDBQuery := fmt.Sprintf(`CREATE DATABASE "%s"`, s.testDBName)
	_, err = conn.Exec(ctx, createDBQuery)
	if err != nil {
		log.Printf("Failed to create test database: %v", err)
		return nil, err
	}

	dsn := fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", cfg.PostgresUser, cfg.PostgresPassword, s.testDBName)

	client, err := database.NewEntClient(dsn)
	if err != nil {
		log.Printf("Failed to create Ent client: %v", err)
		return nil, err
	}

	return client, nil
}

// dropTestDatabase drops the test database
func (s *IntegrationTestSuite) dropTestDatabase(ctx context.Context) error {
	if s.adminConn == nil {
		return nil
	}

	// Terminate any active connections to the test database
	terminateQuery := `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`
	_, err := s.adminConn.Exec(ctx, terminateQuery, s.testDBName)
	if err != nil {
		log.Printf("Warning: failed to terminate connections to test database: %v", err)
	}

	// Drop test database
	dropDBQuery := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, s.testDBName)
	_, err = s.adminConn.Exec(ctx, dropDBQuery)
	if err != nil {
		return fmt.Errorf("failed to drop test database: %w", err)
	}

	return nil
}

// runMigrations applies database migrations to the test database
func (s *IntegrationTestSuite) runMigrations(ctx context.Context) error {
	return s.client.Schema.Create(ctx)
}

// cleanupTestData removes all data from test database between tests
func (s *IntegrationTestSuite) cleanupTestData() {
	ctx := context.Background()

	// Delete data from all Ent entities
	// Add more entities as your schema grows

	// Delete all categories
	_, err := s.client.Category.Delete().Exec(ctx)
	if err != nil {
		log.Printf("Warning: failed to delete categories: %v", err)
	}
}

// TestMain runs the test suite
// func TestMain(m *testing.M) {
// 	// Run the test suite
// 	code := m.Run()
// 	os.Exit(code)
// }

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
