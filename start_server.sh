#!/bin/bash

# Start HTTP server with cloud database
export PGSSLROOTCERT="$HOME/.cloud-certs/root.crt"

echo "🚀 Starting Etalon Price API Server..."
echo "Database: Cloud PostgreSQL (TWC)"
echo "Provider: 4tochki (sa69263)"
echo ""

# Build if needed
if [ ! -f "bin/app" ]; then
    echo "Building application..."
    go build -o bin/app ./cmd/app
fi

# Run
./bin/app
