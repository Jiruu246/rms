package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Jiruu246/rms/internal/config"
	"github.com/Jiruu246/rms/internal/handler"
	"github.com/Jiruu246/rms/internal/middlewares"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	cfg    *config.Config
	db     *sqlx.DB
	engine *gin.Engine
	srv    *http.Server
}

func New(cfg *config.Config, db *sqlx.DB) *Server {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	// Add CORS middleware at the top level
	if cfg.Env == "production" {
		// Use restrictive CORS for production with configured origins
		engine.Use(middlewares.RestrictiveCORS(cfg.AllowedOrigins))
	} else {
		// Use permissive CORS for development
		engine.Use(middlewares.CORS())
	}

	s := &Server{
		cfg:    cfg,
		db:     db,
		engine: engine,
	}

	s.routes()

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":"+"%d", cfg.Port),
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	return s
}

func (s *Server) routes() {
	// initialize repositories
	categoryRepo := repos.NewCategoryRepository(s.db)

	// initialize services
	categoryService := services.NewCategoryService(categoryRepo)

	// initialize handlers
	categoryHandler := handler.NewCategoryHandler(categoryService)

	// API routes
	api := s.engine.Group("/api")
	{
		categories := api.Group("/categories")
		{
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.PATCH("/:id", categoryHandler.UpdateCategory)
			categories.DELETE("/:id", categoryHandler.DeleteCategory)
		}
	}
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	fmt.Printf("listening on %s\n", s.srv.Addr)
	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	// allow in-flight requests a short time to finish
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctxTimeout)
}
