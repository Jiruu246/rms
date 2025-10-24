# rms — Gin-based service bootstrap

This repository contains a production-ready bootstrap for a medium-sized Go service using Gin.

Features:

- Structured layout: `cmd/`, `internal/`, `pkg/`.
- Config via environment (Viper), structured `Config` type.
- Zap structured logging.
- Database helper using `sqlx` + `pgx` (Postgres) with a single entrypoint.
- Graceful shutdown, health endpoint, and tests.
- Docker multi-stage build and Make targets.

Quick start

1. Set environment variables (example):

```bash
export APP_PORT=8080
export APP_DATABASE_URL=postgres://user:pass@localhost:5432/dbname
export APP_LOG_LEVEL=debug
```

2. Build and run:

```bash
make tidy
make build
./rms
```

3. Run tests:

```bash
make test
```

Notes and next steps

- Run `go mod tidy` to fetch new dependencies used by the scaffold.
- Add request logging middleware, metrics (Prometheus), tracing (OpenTelemetry), and migration tool (golang-migrate) as next steps.
