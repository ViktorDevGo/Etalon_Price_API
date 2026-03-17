#!/bin/bash
set -e

echo "🚀 Etalon Price API - Cloud Database Setup"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Database connection
DB_HOST="c37e696087932476c61fd621.twc1.net"
DB_PORT="5432"
DB_USER="gen_user"
DB_PASS="Poison-79"
DB_NAME="default_db"
DB_DSN="postgresql://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=verify-full"

# SSL certificate
CERT_FILE="$HOME/.cloud-certs/root.crt"

echo "1️⃣  Checking SSL certificate..."
if [ ! -f "$CERT_FILE" ]; then
    echo -e "${YELLOW}⚠️  SSL certificate not found${NC}"
    echo "Running SSL setup..."
    ./setup_ssl.sh
fi

if [ ! -f "$CERT_FILE" ]; then
    echo -e "${RED}❌ SSL certificate still not found${NC}"
    echo "Please set up SSL certificate manually:"
    echo "  mkdir -p ~/.cloud-certs"
    echo "  Download root.crt to ~/.cloud-certs/root.crt"
    exit 1
fi

echo -e "${GREEN}✅ SSL certificate found${NC}"
export PGSSLROOTCERT="$CERT_FILE"
echo ""

echo "2️⃣  Testing database connection..."
if command -v psql &> /dev/null; then
    if psql "$DB_DSN" -c "SELECT version();" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Database connection successful${NC}"
    else
        echo -e "${RED}❌ Database connection failed${NC}"
        echo "Please check your credentials and network connection"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  psql not installed, skipping connection test${NC}"
fi
echo ""

echo "3️⃣  Checking if migrations are needed..."
# Check if tables exist
TABLES_EXIST=$(psql "$DB_DSN" -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='tyres_api';" 2>/dev/null || echo "0")

if [ "$TABLES_EXIST" -eq "0" ]; then
    echo -e "${YELLOW}⚠️  Tables not found, running migrations...${NC}"

    # Check if golang-migrate is installed
    if command -v migrate &> /dev/null; then
        echo "Applying migrations..."
        export PGSSLROOTCERT="$CERT_FILE"
        migrate -path internal/migrations -database "$DB_DSN" up
        echo -e "${GREEN}✅ Migrations applied${NC}"
    else
        echo -e "${RED}❌ golang-migrate not installed${NC}"
        echo "Install it with:"
        echo "  brew install golang-migrate"
        echo "Or download from: https://github.com/golang-migrate/migrate"
        exit 1
    fi
else
    echo -e "${GREEN}✅ Tables already exist${NC}"
fi
echo ""

echo "4️⃣  Building application..."
if go build -o bin/app ./cmd/app && go build -o bin/sync ./cmd/sync; then
    echo -e "${GREEN}✅ Application built${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo ""

echo "=========================================="
echo -e "${GREEN}✅ Setup completed!${NC}"
echo "=========================================="
echo ""
echo "📝 Next steps:"
echo ""
echo "1. Run HTTP server:"
echo "   export PGSSLROOTCERT=$CERT_FILE"
echo "   ./bin/app"
echo ""
echo "2. Or run sync:"
echo "   export PGSSLROOTCERT=$CERT_FILE"
echo "   ./bin/sync --supplier=4tochki --codes=2329500,WHS063930"
echo ""
echo "3. Test with real API:"
echo "   # Start server in background"
echo "   export PGSSLROOTCERT=$CERT_FILE"
echo "   ./bin/app &"
echo ""
echo "   # Wait a bit"
echo "   sleep 3"
echo ""
echo "   # Test sync"
echo "   curl -X POST http://localhost:8080/sync/4tochki \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"codes\":[\"2329500\",\"WHS063930\"]}'"
echo ""
echo "4. Check database:"
echo "   export PGSSLROOTCERT=$CERT_FILE"
echo "   psql '$DB_DSN'"
echo ""
