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

# Testing
TODO: Instruction for testint, unit testing & integration testing

# CICD
TODO: Instruction & documentation for testing

# Linting
Install go lint locally (check the version of linting in the CI to avoid mismatch rules)
```
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11
```

install go lint and runs before create pr

# To do before make PR
- compile: make sure no error
- lint: resolve linting issue `golangci-lint run`
- generate ORM
- run test

# Hassle free container launching workflow for dev
## First time, or after changing Go code
```
docker compose -f docker-compose.local.yml up --build
```
If your Go source code changes, the Docker image must be rebuilt so the new binary gets copied into the image.


## Subsequent runs (no code changes)

```
docker compose -f docker-compose.local.yml up
```


## Tear down (keeps the postgres volume)

```
docker compose -f docker-compose.local.yml down
```
This will

Stops and removes:

- Containers
- Networks created by Compose

What it does NOT remove
- Named volumes
- Images

For example, if PostgreSQL data is stored in a named volume, that volume remains.

So when you start again, your database still contains all previous data.

## Full reset including the DB volume
docker compose -f docker-compose.local.yml down -v

This will delete
- Containers
- Networks
- Named volumes