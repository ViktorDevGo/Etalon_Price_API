#!/bin/bash
export DB_DSN="postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require"
export FOURTOCHKI_LOGIN="sa69263"
export FOURTOCHKI_PASSWORD="jkHP4Nj3)z"
export FOURTOCHKI_WSDL_URL="http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl"

echo "========================================="
echo "Проверка сохраненных данных в БД"
echo "========================================="
echo ""

./bin/check-duplicates << 'SQL'
SELECT
    code,
    supplier,
    brand,
    model,
    width,
    height,
    diameter,
    load_index,
    speed_index,
    season,
    description,
    created_at,
    updated_at
FROM tyres_api
WHERE supplier = '4tochki'
ORDER BY code;
SQL
