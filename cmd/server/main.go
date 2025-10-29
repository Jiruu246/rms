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

	// init database
	db, err := database.NewEntClient(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init database: %v\n", err)
	}
	defer db.Close()

	// create server
	srv := server.New(cfg, db)

	// run server with graceful shutdown
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server start failed: %v\n", err)
		}
	}()

	// wait for signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("shutdown initiated")

	ctxShut, cancel := context.WithTimeout(ctx, time.Duration(cfg.ShutdownTimeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShut); err != nil {
		fmt.Fprintf(os.Stderr, "server forced to shutdown: %v\n", err)
	}

	fmt.Println("server exited")
}
