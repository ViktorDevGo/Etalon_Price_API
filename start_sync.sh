#!/bin/bash

# Run sync with cloud database
export PGSSLROOTCERT="$HOME/.cloud-certs/root.crt"

echo "🔄 Starting synchronization with 4tochki..."
echo "Database: Cloud PostgreSQL (TWC)"
echo "Provider: 4tochki (sa69263)"
echo ""

# Build if needed
if [ ! -f "bin/sync" ]; then
    echo "Building sync CLI..."
    go build -o bin/sync ./cmd/sync
fi

# Run with provided arguments or defaults
if [ $# -eq 0 ]; then
    echo "Using default test codes..."
    ./bin/sync --supplier=4tochki --codes=2329500,WHS063930
else
    ./bin/sync "$@"
fi
