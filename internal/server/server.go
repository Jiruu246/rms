package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Jiruu246/rms/internal/config"
	"github.com/Jiruu246/rms/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	db     *sqlx.DB
	engine *gin.Engine
	srv    *http.Server
}

func New(cfg *config.Config, log *zap.Logger, db *sqlx.DB) *Server {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	s := &Server{
		cfg:    cfg,
		logger: log,
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
	// health
	s.engine.GET("/health", handler.Health)

	// versioned API group placeholder
	v1 := s.engine.Group("/v1")
	{
		v1.GET("/health", handler.Health)
	}
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	s.logger.Sugar().Infof("listening on %s", s.srv.Addr)
	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	// allow in-flight requests a short time to finish
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctxTimeout)
}
