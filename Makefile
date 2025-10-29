.PHONY: build run test tidy docker migrate-up migrate-down migrate-status migrate-create migrate-fresh ent-generate

BINARY := rms
MIGRATE := ./cmd/migrate

build:
	go build -o $(BINARY) ./cmd/server

run: build
	./$(BINARY)

tidy:
	go mod tidy

test:
	go test ./... -v

docker: build
	docker build -t github.com/Jiruu246/rms:latest .

# Ent code generation
ent-generate:
	go generate ./internal/ent

# Database Migration Commands
migrate-apply:
	go run $(MIGRATE) apply

migrate-reset:
	go run $(MIGRATE) reset

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=migration_name"; exit 1; fi
	go run $(MIGRATE) create $(name)
