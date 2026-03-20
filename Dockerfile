# ================================
# Stage 1: Build all Go binaries
# ================================
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all binaries
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/api ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/nomenclature-scheduler ./cmd/nomenclature-scheduler
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/sync-prices ./cmd/sync-prices
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/severavto-scheduler ./cmd/severavto-scheduler

# Build export binaries for prices-upload
RUN go build -o /app/export-bitrix-prices-mrc ./cmd/export-bitrix-prices-mrc
RUN go build -o /app/export-bitrix-prices ./cmd/export-bitrix-prices
RUN go build -o /app/export-bitrix-prices-moto ./cmd/export-bitrix-prices-moto
RUN go build -o /app/export-bitrix-prices-rims ./cmd/export-bitrix-prices-rims

# ================================
# Stage 2: Runtime
# ================================
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    bash \
    supervisor \
    openssh-client \
    sshpass \
    tzdata \
    ca-certificates \
    wget

# Set timezone to Moscow (можно переопределить через ENV)
ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Create app directory
WORKDIR /app

# Copy all binaries from builder
COPY --from=builder /app/* /app/

# Create upload script for prices-upload
RUN cat > /app/upload_prices.sh <<'SCRIPT'
#!/bin/bash
set -e

REMOTE_HOST="${BITRIX_REMOTE_HOST:-147.45.215.76}"
REMOTE_USER="${BITRIX_REMOTE_USER:-root}"
REMOTE_PASSWORD="${BITRIX_REMOTE_PASSWORD}"
REMOTE_PATH="${BITRIX_REMOTE_PATH:-/home/bitrix/www/upload/1c_catalog}"
TEMP_DIR="/tmp/bitrix_export"

mkdir -p "$TEMP_DIR"

echo "📦 Generating price files..."
/app/export-bitrix-prices-mrc --output-dir="$TEMP_DIR" || exit 1
/app/export-bitrix-prices --output-dir="$TEMP_DIR" || exit 1
/app/export-bitrix-prices-moto --output-dir="$TEMP_DIR" || exit 1
/app/export-bitrix-prices-rims --output-dir="$TEMP_DIR" || exit 1

echo "📤 Uploading files..."
for file in "$TEMP_DIR"/*.csv; do
    sshpass -p "$REMOTE_PASSWORD" scp -o StrictHostKeyChecking=no "$file" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"
done

rm -rf "$TEMP_DIR"
echo "✅ Upload completed"
SCRIPT

RUN chmod +x /app/upload_prices.sh

# Create crontab for prices-upload and sync-prices
RUN mkdir -p /etc/crontabs && \
    echo "0 7 * * * /app/upload_prices.sh >> /var/log/prices-upload.log 2>&1" > /etc/crontabs/root && \
    echo "0 */3 * * * /app/sync-prices -type=all >> /var/log/sync-prices.log 2>&1" >> /etc/crontabs/root

# Create supervisord config
RUN cat > /etc/supervisord.conf <<'SUPERVISOR'
[supervisord]
nodaemon=true
user=root
logfile=/var/log/supervisord.log
pidfile=/var/run/supervisord.pid

[program:api]
command=/app/api
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0

[program:nomenclature-scheduler]
command=/app/nomenclature-scheduler
autostart=true
autorestart=true
stdout_logfile=/var/log/nomenclature.log
stderr_logfile=/var/log/nomenclature.err.log

[program:severavto-scheduler]
command=/app/severavto-scheduler
autostart=true
autorestart=true
stdout_logfile=/var/log/severavto.log
stderr_logfile=/var/log/severavto.err.log

[program:crond]
command=crond -f -l 2 -L /var/log/cron.log
autostart=true
autorestart=true
stdout_logfile=/var/log/cron.log
stderr_logfile=/var/log/cron.err.log
SUPERVISOR

# Create log directory
RUN mkdir -p /var/log && \
    touch /var/log/api.log \
          /var/log/nomenclature.log \
          /var/log/sync-prices.log \
          /var/log/severavto.log \
          /var/log/prices-upload.log \
          /var/log/cron.log

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/healthz || exit 1

# Start supervisord
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
