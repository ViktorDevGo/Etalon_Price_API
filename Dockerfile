# Multi-stage build for Etalon Price API

# Builder stage
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all binaries
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/api ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/sync-nomenclature ./cmd/sync-nomenclature
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/sync-prices ./cmd/sync-prices
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/nomenclature-scheduler ./cmd/nomenclature-scheduler

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/api .
COPY --from=builder /app/sync-nomenclature .
COPY --from=builder /app/sync-prices .
COPY --from=builder /app/nomenclature-scheduler .

# Copy migrations
COPY --from=builder /build/migrations ./migrations

# Set timezone
ENV TZ=Europe/Moscow

# Default command (can be overridden in docker-compose)
CMD ["./api"]
