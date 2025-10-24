.PHONY: build run test tidy docker

BINARY := rms

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
