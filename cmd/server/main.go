package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Jiruu246/rms/internal/config"
	"github.com/Jiruu246/rms/internal/server"
	"github.com/Jiruu246/rms/pkg/database"
	"github.com/Jiruu246/rms/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stdout, "failed to load config: %v\n", err)
	}

	// load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// init logger
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// init database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Sugar().Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	// create server
	srv := server.New(cfg, log, db)

	// run server with graceful shutdown
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Sugar().Fatalf("server start failed: %v", err)
		}
	}()

	// wait for signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info("shutdown initiated")

	ctxShut, cancel := context.WithTimeout(ctx, time.Duration(cfg.ShutdownTimeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShut); err != nil {
		log.Sugar().Fatalf("server forced to shutdown: %v", err)
	}

	log.Info("server exited")
}
