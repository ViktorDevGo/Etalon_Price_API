#!/bin/bash
export DB_DSN="postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require"
export FOURTOCHKI_LOGIN="sa69263"
export FOURTOCHKI_PASSWORD="jkHP4Nj3)z"
export FOURTOCHKI_WSDL_URL="http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl"

echo "Тестирование с валидными кодами от 4tochki..."
echo "Коды: 1027210, 1026634"
echo ""

./bin/sync --supplier=4tochki --codes-file=valid_codes.txt
