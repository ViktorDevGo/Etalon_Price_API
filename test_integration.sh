#!/bin/bash
set -e

echo "🧪 Integration Test Script"
echo "=========================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

API_URL="http://localhost:8080"

# Function to test endpoint
test_endpoint() {
    local name=$1
    local url=$2
    local method=${3:-GET}
    local data=${4:-}

    echo -n "Testing $name... "

    if [ "$method" = "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" "$url")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✅ OK${NC}"
        echo "Response: $body" | head -c 200
        echo ""
        return 0
    else
        echo -e "${RED}❌ FAILED (HTTP $http_code)${NC}"
        echo "Response: $body"
        return 1
    fi
}

echo "1️⃣  Checking if API is running..."
if ! curl -s "$API_URL/healthz" > /dev/null 2>&1; then
    echo -e "${RED}❌ API is not running on $API_URL${NC}"
    echo "Start it with: make docker-up"
    exit 1
fi
echo -e "${GREEN}✅ API is running${NC}"
echo ""

echo "2️⃣  Testing endpoints..."
test_endpoint "Health Check" "$API_URL/healthz"
test_endpoint "Readiness Check" "$API_URL/readyz"
test_endpoint "Version" "$API_URL/version"
test_endpoint "Stats" "$API_URL/stats"
echo ""

echo "3️⃣  Testing sync with example codes..."
echo -e "${YELLOW}⚠️  This will make a real API call to 4tochki!${NC}"
echo -e "${YELLOW}⚠️  Make sure you have valid credentials in .env${NC}"
echo ""

read -p "Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Skipped."
    exit 0
fi

SYNC_DATA='{
  "codes": ["2329500", "WHS063930"]
}'

echo "Sending sync request..."
response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/sync/4tochki" \
    -H "Content-Type: application/json" \
    -d "$SYNC_DATA")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo ""
echo "HTTP Status: $http_code"
echo "Response:"
echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
echo ""

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Sync completed successfully!${NC}"

    # Check database
    echo ""
    echo "4️⃣  Checking database..."
    echo "Connecting to PostgreSQL..."

    docker-compose exec -T postgres psql -U etalon -d etalon_price << EOF
\echo '--- Tyres ---'
SELECT COUNT(*) as tyres_count FROM tyres_api;
\echo ''
\echo '--- Rims ---'
SELECT COUNT(*) as rims_count FROM rims_api;
\echo ''
\echo '--- Offers ---'
SELECT COUNT(*) as offers_count FROM product_offers_api;
\echo ''
\echo '--- Last Sync Run ---'
SELECT id, provider, status, total_products, success_products, failed_products
FROM sync_runs_api
ORDER BY created_at DESC
LIMIT 1;
EOF

    echo ""
    echo -e "${GREEN}✅ Integration test PASSED!${NC}"
else
    echo -e "${RED}❌ Sync failed!${NC}"
    echo ""
    echo "Common issues:"
    echo "1. Invalid credentials in .env"
    echo "2. Invalid product codes"
    echo "3. Network issues"
    echo ""
    echo "Check logs with: docker-compose logs app"
    exit 1
fi

echo ""
echo "=================================="
echo -e "${GREEN}✅ All tests passed!${NC}"
echo "=================================="
