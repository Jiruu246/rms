# RMS — Claude Code Context

## Project overview

Restaurant Management System — Go backend, gin HTTP framework, ent ORM (v0.14.5), PostgreSQL.

Module: `github.com/Jiruu246/rms`  
Go version: 1.25.1

## Directory layout

```
cmd/                  # CLI entrypoints (server, migrate)
internal/
  config/             # Config loading (viper + godotenv)
  cookies/            # Cookie helpers
  data_structures/    # Shared data structures
  dto/                # Request/response DTOs
  ent/                # Generated ent ORM code — DO NOT hand-edit
    schema/           # ← edit here; run `go generate ./internal/ent` to regenerate
    predicate/        # Generated per-entity predicate types
  handler/            # HTTP handlers (gin)
  middlewares/        # Gin middleware (auth, etc.)
  repos/              # Data access layer — all DB queries live here
  services/           # Business logic layer
  server/             # HTTP server wiring
pkg/
  database/           # DB connection helpers
  logger/             # Structured logger
  pagination/         # Reusable cursor-pagination engine (see below)
  utils/
integration_tests/
```

## Pagination system (`pkg/pagination` + `internal/repos/*_pagination.go`)

Cursor (keyset/seek) pagination — NOT offset. Every sort ends with `id ASC` as an implicit tie-breaker.

### Engine files

| File | Role |
|------|------|
| `pkg/pagination/cursor.go` | `Cursor`, `SortSpec`; `EncodeCursor`/`DecodeCursor` (base64 URL-safe JSON) |
| `pkg/pagination/page.go` | `PageRequest`, `PageResponse[T]`, `ParsePageRequest`, `ParseSortParam` |
| `pkg/pagination/sort.go` | `SortFieldSpec[Row]` (per-entity declaration struct), internal `resolvedField` |
| `pkg/pagination/predicate.go` | Recursive OR/AND keyset predicate builder |
| `pkg/pagination/engine.go` | `QueryExecutor[Row]`, `Run[Row]`, sentinel errors |
| `pkg/pagination/engine_test.go` | 21 unit tests; no DB required |

### Key types

```go
type SortFieldSpec[Row any] struct {
    Asc, Desc func(*sql.Selector)              // ORDER BY helpers from ent codegen
    Extract   func(row Row) any                // read field value for cursor encoding
    Eq, Lt, Gt func(v any) func(*sql.Selector) // WHERE predicate builders
    Decode    func(v any) (any, error)          // JSON round-trip: string→time, float64→int, etc.
}

type QueryExecutor[Row any] func(
    ctx context.Context,
    orders []func(*sql.Selector),
    cursorPred func(*sql.Selector), // nil on first page
    limit int,
) ([]Row, error)
```

### Adapter pattern (per entity, ~80 lines)

ent generates entity-specific named types (`category.OrderOption`, `predicate.Category`) that share `func(*sql.Selector)` as their underlying type. The engine works entirely with `func(*sql.Selector)`; each adapter closure converts via a simple type conversion:

```go
catOrders[i] = category.OrderOption(orders[i])  // same underlying type — valid
q = q.Where(predicate.Category(cursorPred))
```

**Worked example:** `internal/repos/category_pagination.go`  
Exposes: `ListCategories(ctx, client, req, filters)` and `NewCategoryQueryExecutor(q)`.

### Adding a new entity adapter

1. Create `internal/repos/<entity>_pagination.go`
2. Declare `var <entity>SortFields = map[string]pagination.SortFieldSpec[*ent.<Entity>]{...}`
3. Implement `New<Entity>QueryExecutor(q *ent.<Entity>Query) pagination.QueryExecutor[*ent.<Entity>]`
4. Implement `List<Entity>s(ctx, client, req, filters)` — apply filters to `q` before wrapping
5. Add composite DB indexes `(field, id)` for every sortable field in the ent schema

### Sentinel errors (map to HTTP status in handlers)

| Error | HTTP |
|-------|------|
| `pagination.ErrInvalidSortField` | 400 |
| `pagination.ErrCursorSortMismatch` | 400 |
| `pagination.ErrInvalidCursor` | 400 |

### JSON decode gotchas

`map[string]any` JSON unmarshal always produces `float64` for numbers and `string` for time values. Each `SortFieldSpec.Decode` must re-parse to the correct Go type. See `display_order` (float64→int) and `create_time` (string→time.Time via RFC3339Nano) in the Category adapter.

## ent ORM

- Schema files live in `internal/ent/schema/` — edit these, then regenerate.
- Regenerate: `go generate ./internal/ent` (or `make generate` if wired in Makefile).
- Primary keys are UUIDs (`uuid.UUID`).
- All entities have `create_time` and `update_time` via ent's `mixin.Time`.
- Generated predicate types live in `internal/ent/predicate/` — one named type per entity, all with underlying type `func(*sql.Selector)`.

## Stack

- HTTP: `gin-gonic/gin` v1.11
- ORM: `entgo.io/ent` v0.14.5
- DB driver: `jackc/pgx/v5`
- Auth: `golang-jwt/jwt/v5`
- Config: `spf13/viper` + `joho/godotenv`
- UUID: `google/uuid`
- Tests: `stretchr/testify`

## Running locally

```sh
# Start DB (see docker-compose.local.yml)
docker compose -f docker-compose.local.yml up -d

# Run server
go run ./cmd/server

# Run tests (unit — no DB)
go test ./pkg/pagination/...

# Integration tests
go test ./integration_tests/...
```

## Conventions

- Filters are applied to the ent query **before** wrapping it in `QueryExecutor` — they are orthogonal to pagination.
- Handlers parse `PageRequest` via `pagination.ParsePageRequest(c.Query("limit"), c.Query("cursor"), c.Query("sort"))`.
- Default sort must be applied by the `List*` function when `req.Sort` is empty — `Run` does not apply defaults.
- `prev_cursor` backward pagination is not implemented (field exists in `PageResponse` but is always empty).
