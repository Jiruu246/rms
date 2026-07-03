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
  docs/               # Generated Swagger/OpenAPI spec — DO NOT hand-edit (see API documentation below)
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

## Pagination system (`pkg/pagination` + per-entity adapters in `internal/repos/*_repo.go`)

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

The adapter lives alongside the entity's other repo code in `internal/repos/<entity>_repo.go` — it is not split into a separate file.

**Worked example:** `internal/repos/category_repo.go`  
Exposes: `ListCategories(ctx, client, req, filters)`, `NewCategoryQueryExecutor(q)`, and `CategoryRepository.List(ctx, req)` (the thin repo-interface method that calls `ListCategories` with the repo's own `ent.Client`).

### Adding a new entity adapter

1. In `internal/repos/<entity>_repo.go`, declare `var <entity>SortFields = map[string]pagination.SortFieldSpec[*ent.<Entity>]{...}`
2. Implement `New<Entity>QueryExecutor(q *ent.<Entity>Query) pagination.QueryExecutor[*ent.<Entity>]`
3. Implement `List<Entity>s(ctx, client, req, filters)` — apply filters to `q` before wrapping
4. Add a `List(ctx, req)` method to the entity's repository interface/struct that calls `List<Entity>s` and maps `*ent.<Entity>` rows to the entity's DTO
5. Add composite DB indexes `(field, id)` for every sortable field in the ent schema

### Sentinel errors (map to HTTP status in handlers)

| Error | HTTP |
|-------|------|
| `pagination.ErrInvalidSortField` | 400 |
| `pagination.ErrCursorSortMismatch` | 400 |
| `pagination.ErrInvalidCursor` | 400 |

### JSON decode gotchas

`map[string]any` JSON unmarshal always produces `float64` for numbers and `string` for time values. Each `SortFieldSpec.Decode` must re-parse to the correct Go type. See `display_order` (float64→int) and `create_time` (string→time.Time via RFC3339Nano) in the Category adapter.

## API documentation (Swagger / swaggo)

Handlers are annotated with `swaggo/swag` comments; the OpenAPI 2.0 spec is generated into `internal/docs/` (`docs.go`, `swagger.json`, `swagger.yaml`) and served at `/swagger/index.html` via `gin-swagger`.

The route is only registered when `cfg.Env != "production"` (see `Server.routes()` in `internal/server/server.go`) — the API spec is not exposed on production deployments.

- General API info (`@title`, `@BasePath`, `@securityDefinitions.apikey BearerAuth`, ...) lives above `func main()` in `cmd/server/main.go`.
- Each handler method has its own `@Summary`/`@Tags`/`@Param`/`@Success`/`@Failure`/`@Router` block directly above the function — keep it next to the code it documents, not in a separate file.
- Response bodies are documented as the real generic envelope types, e.g. `utils.APIResponse[dto.Category]` or `utils.APIResponse[pagination.PageResponse[dto.Category]]` — swag resolves Go generics natively.
- Routes requiring `JwtMiddleware` (see `internal/server/server.go`) get `@Security BearerAuth`; public routes (e.g. `/public/order`) omit it.
- `@Router` paths are relative to `BasePath` (`/api`) and must match the method registered in `internal/server/server.go` exactly (categories/restaurants/menu-items use `PUT` for update; modifiers/modifier-options/orders use `PATCH` — don't assume PATCH everywhere).

Regenerate after changing any handler annotation or DTO:

```sh
swag init -g cmd/server/main.go -o internal/docs --parseDependency --parseInternal
```

Install the CLI once via `go install github.com/swaggo/swag/cmd/swag@latest` if `swag` isn't on `PATH`.

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
# API docs: http://localhost:<port>/swagger/index.html

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

### Go file organization

Organize files for readability rather than enforcing a strict visibility-based layout:

```
package
imports

const
var

types
    interfaces
    structs
    aliases

constructors (New...)

public API

private helpers
```

- **Group by feature, not visibility.** A private helper lives directly below the public function/method that uses it, not collected in a separate section at the end of the file.

  Preferred:
  ```go
  func (s *Service) CreateUser(...) error {
      return s.validate(...)
  }

  func (s *Service) validate(...) error {
      ...
  }
  ```

  Avoid:
  ```go
  // Public methods
  func (s *Service) CreateUser(...) error { ... }
  func (s *Service) DeleteUser(...) error { ... }

  // Private helpers
  func (s *Service) validate(...) error { ... }
  func (s *Service) normalize(...) { ... }
  ```

- **Keep files focused.** When a file grows large or covers multiple responsibilities, split it into multiple files in the same package instead of adding more sections — e.g. `service.go` (type + constructor), `service_create.go`, `service_delete.go`, `types.go`, `errors.go`.

The goal is minimal scrolling to find related code.
