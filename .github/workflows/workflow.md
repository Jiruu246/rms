# CI Workflow Documentation

This document explains the GitHub Actions workflows in this repository and how they work together on pull requests.

## Workflow Overview

The repository uses 4 workflow files:

- `.github/workflows/compile-and-test.yml`
- `.github/workflows/test.yml` (reusable workflow)
- `.github/workflows/go-lint.yml`
- `.github/workflows/generated-code.yml`

On pull requests to `main`, CI verifies:

- The project compiles.
- `go mod tidy` does not introduce unexpected changes.
- Integration tests pass.
- Linting passes.
- Generated Ent code is up to date and committed.

## High-Level Flow

1. A pull request targeting `main` is opened or updated.
2. Three PR workflows run independently:
	 - `Compile and Test`
	 - `Go Lint`
	 - `Generated Code Check`
3. Inside `Compile and Test`:
	 - The `compile` job runs first.
	 - If `compile` succeeds, the reusable `Integration Tests` workflow is called.

## 1) Compile and Test

**File:** `.github/workflows/compile-and-test.yml`

### Trigger

- `pull_request` on branch `main`
- Event types:
	- `opened`
	- `synchronize`
	- `reopened`
	- `ready_for_review`

### Concurrency

- Group: `compile-${{ github.workflow }}-${{ github.event.pull_request.number }}`
- Behavior: cancels previous in-progress runs for the same PR.

### Jobs

#### `compile`

- Runs on `ubuntu-latest`.
- Skips draft PRs.
- Steps:
	- Checkout code.
	- Setup Go using `go.mod` version.
	- Download modules (`go mod download`).
	- Run `go mod tidy` and verify `go.mod`/`go.sum` stay unchanged.
	- Build all packages (`go build ./...`).

#### `integration-tests`

- Depends on successful `compile` (`needs: compile`).
- Calls reusable workflow `.github/workflows/test.yml`.
- Uses `secrets: inherit`.

## 2) Integration Tests (Reusable Workflow)

**File:** `.github/workflows/test.yml`

### Trigger

- `workflow_call` only.
- This workflow is not triggered directly by pull request events.

### Required Secrets

- `APP_ENV`
- `APP_POSTGRES_USER`
- `APP_POSTGRES_PASSWORD`
- `APP_JWT_SECRET`
- `APP_ACCESS_TOKEN_EXPIRATION`
- `APP_REFRESH_TOKEN_EXPIRATION`

### Job

#### `integration-tests`

- Runs on `ubuntu-latest`.
- Timeout: 15 minutes.
- Starts a PostgreSQL service container (`postgres:16-alpine`) with health checks.
- Sets `APP_DATABASE_URL` dynamically to connect to local CI PostgreSQL.
- Runs integration tests with:
	- `go test ./...`
	- `-race`
	- `-count=1`
	- `-timeout=10m`
	- `-tags=integration`
- On failure, uploads `/tmp/test-results.xml` as artifact (if present).

## 3) Go Lint

**File:** `.github/workflows/go-lint.yml`

### Trigger

- `pull_request` on branch `main`
- Event types:
	- `opened`
	- `synchronize`
	- `reopened`
	- `ready_for_review`

### Concurrency

- Group: `go-lint-${{ github.workflow }}-${{ github.event.pull_request.number }}`
- Behavior: cancels previous in-progress runs for the same PR.

### Job

#### `golangci-lint`

- Runs on `ubuntu-latest`.
- Skips draft PRs.
- Steps:
	- Checkout code.
	- Setup Go using `go.mod` version.
	- Download modules (`go mod download`).
	- Run `golangci-lint-action@v9` with linter version `v2.11`.

## 4) Generated Code Check

**File:** `.github/workflows/generated-code.yml`

### Trigger

- `pull_request` on branch `main`
- Event types:
	- `opened`
	- `synchronize`
	- `reopened`
	- `ready_for_review`

### Concurrency

- Group: `generated-code-${{ github.workflow }}-${{ github.event.pull_request.number }}`
- Behavior: cancels previous in-progress runs for the same PR.

### Job

#### `ent-generate-check`

- Runs on `ubuntu-latest`.
- Skips draft PRs.
- Steps:
	- Checkout code.
	- Setup Go using `go.mod` version.
	- Download modules (`go mod download`).
	- Regenerate Ent code (`go generate ./internal/ent/...`).
	- Fail if `git diff` is non-empty (generated files not committed).

## Draft PR Behavior

All PR-triggered jobs include:

- `if: ${{ !github.event.pull_request.draft }}`

This means checks do not run while the PR is a draft. They run when the PR is marked ready for review.

## Contributor Checklist for Passing CI

Before pushing a PR, run:

1. `go mod tidy`
2. `go generate ./internal/ent/...`
3. `go build ./...`
4. `go test ./... -race -count=1 -tags=integration`
5. `golangci-lint run`

If CI fails:

- Re-check that generated Ent files are committed.
- Re-check that `go.mod` and `go.sum` are tidy and committed.
- Confirm required secrets are configured in GitHub repository settings for integration tests.
