#!/bin/bash
set -e

echo "🚀 Etalon Price API - Setup Script"
echo "=================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file not found${NC}"
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo -e "${GREEN}✅ .env file created${NC}"
    echo ""
    echo -e "${RED}ВАЖНО: Отредактируйте .env и укажите реальные credentials!${NC}"
    echo "   FOURTOCHKI_LOGIN=ваш_логин"
    echo "   FOURTOCHKI_PASSWORD=ваш_пароль"
    echo ""
    read -p "Press Enter to continue after editing .env..."
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running${NC}"
    echo "Please start Docker Desktop and try again"
    exit 1
fi

echo -e "${GREEN}✅ Docker is running${NC}"
echo ""

# Stop any existing containers
echo "🛑 Stopping existing containers..."
docker-compose down -v 2>/dev/null || true

# Start PostgreSQL
echo "🐘 Starting PostgreSQL..."
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 10

# Check PostgreSQL health
MAX_RETRIES=30
RETRY=0
until docker-compose exec -T postgres pg_isready -U etalon -d etalon_price > /dev/null 2>&1; do
    RETRY=$((RETRY+1))
    if [ $RETRY -eq $MAX_RETRIES ]; then
        echo -e "${RED}❌ PostgreSQL failed to start${NC}"
        exit 1
    fi
    echo "Waiting for PostgreSQL... ($RETRY/$MAX_RETRIES)"
    sleep 2
done

echo -e "${GREEN}✅ PostgreSQL is ready${NC}"
echo ""

# Run migrations
echo "🔄 Running database migrations..."
docker-compose up migrate

echo -e "${GREEN}✅ Migrations completed${NC}"
echo ""

# Build application
echo "🔨 Building application..."
docker-compose build app

echo -e "${GREEN}✅ Application built${NC}"
echo ""

# Start application
echo "🚀 Starting application..."
docker-compose up -d app

# Wait for application to be ready
echo "⏳ Waiting for application to be ready..."
sleep 5

# Check health
MAX_RETRIES=20
RETRY=0
until curl -s http://localhost:8080/healthz > /dev/null 2>&1; do
    RETRY=$((RETRY+1))
    if [ $RETRY -eq $MAX_RETRIES ]; then
        echo -e "${RED}❌ Application failed to start${NC}"
        echo "Check logs with: docker-compose logs app"
        exit 1
    fi
    echo "Waiting for application... ($RETRY/$MAX_RETRIES)"
    sleep 2
done

echo -e "${GREEN}✅ Application is ready${NC}"
echo ""

# Show status
echo "📊 Status:"
docker-compose ps

echo ""
echo "=================================="
echo -e "${GREEN}✅ Setup completed successfully!${NC}"
echo "=================================="
echo ""
echo "📝 Next steps:"
echo ""
echo "1. Test health endpoint:"
echo "   curl http://localhost:8080/healthz"
echo ""
echo "2. Test sync with example codes:"
echo "   curl -X POST http://localhost:8080/sync/4tochki \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"codes\":[\"2329500\",\"WHS063930\"]}'"
echo ""
echo "3. Or use CLI:"
echo "   docker-compose run --rm app /app/sync --supplier=4tochki --codes=2329500,WHS063930"
echo ""
echo "4. View logs:"
echo "   docker-compose logs -f app"
echo ""
echo "5. Stop services:"
echo "   docker-compose down"
echo ""
