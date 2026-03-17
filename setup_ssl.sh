#!/bin/bash
set -e

echo "🔒 Setting up SSL certificate for cloud database"
echo "==============================================="

CERT_DIR="$HOME/.cloud-certs"
CERT_FILE="$CERT_DIR/root.crt"

# Create directory if not exists
mkdir -p "$CERT_DIR"

# Check if certificate already exists
if [ -f "$CERT_FILE" ]; then
    echo "✅ SSL certificate already exists at $CERT_FILE"
    exit 0
fi

echo "📥 Downloading SSL certificate..."

# Download TimewWeb Cloud root certificate
# This is the public root CA certificate for TWC PostgreSQL
curl -s -o "$CERT_FILE" https://timeweb.cloud/blog/wp-content/uploads/2023/10/root.crt || {
    echo "⚠️  Failed to download certificate from TimewWeb"
    echo "Please download manually from https://timeweb.cloud/help/pages/viewpage.action?pageId=103154709"
    echo "And save to: $CERT_FILE"
    exit 1
}

echo "✅ SSL certificate downloaded successfully"
echo "Location: $CERT_FILE"
echo ""
echo "You can now connect to the database with SSL verification."
