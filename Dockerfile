# Build stage
FROM golang:1.25.1-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
RUN go mod download

COPY . .
RUN go build -o /app ./cmd/server

# Final image
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app /app
ENTRYPOINT ["/app"]
