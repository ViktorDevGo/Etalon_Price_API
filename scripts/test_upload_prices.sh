#!/bin/bash
# Тестовый скрипт для проверки генерации файлов БЕЗ загрузки на сервер

set -e

TEMP_DIR="/tmp/bitrix_export_test"
DATE_STR=$(date +%Y%m%d)
FILENAME_MRC="Переоценка_МРЦ_${DATE_STR}.csv"
FILENAME_TYRES="Переоценка_шины_${DATE_STR}.csv"
FILENAME_MOTO="Переоценка_мотошины_${DATE_STR}.csv"
WORK_DIR="${WORK_DIR:-/Users/viktor/Pro_Koleso/Etalon_Price_API}"

echo "================================================================"
echo "ТЕСТОВЫЙ РЕЖИМ: Генерация файлов переоценки"
echo "================================================================"
echo "🕐 Время запуска: $(date '+%Y-%m-%d %H:%M:%S %Z')"
echo ""

# Создаем временную директорию
mkdir -p "$TEMP_DIR"

# 1. Генерируем файлы переоценки
echo "📦 Шаг 1: Генерация файлов переоценки..."
cd "$WORK_DIR"

# ВАЖНО: Порядок генерации - МРЦ, Легковые, Мотошины

# 1.1 МРЦ (ПЕРВЫМ!)
echo "  → Генерация файла МРЦ..."
go run cmd/export-bitrix-prices-mrc/main.go --output-dir="$TEMP_DIR"

if [ ! -f "$TEMP_DIR/$FILENAME_MRC" ]; then
    echo "❌ Ошибка: файл МРЦ не создан!"
    exit 1
fi

MRC_LINES=$(wc -l < "$TEMP_DIR/$FILENAME_MRC")
echo "  ✅ Файл создан: $FILENAME_MRC ($MRC_LINES строк)"

# 1.2 Легковые шины
echo "  → Генерация файла легковых шин..."
go run cmd/export-bitrix-prices/main.go --output-dir="$TEMP_DIR"

if [ ! -f "$TEMP_DIR/$FILENAME_TYRES" ]; then
    echo "❌ Ошибка: файл легковых шин не создан!"
    exit 1
fi

TYRES_LINES=$(wc -l < "$TEMP_DIR/$FILENAME_TYRES")
echo "  ✅ Файл создан: $FILENAME_TYRES ($TYRES_LINES строк)"

# 1.3 Мотошины
echo "  → Генерация файла мотошин..."
go run cmd/export-bitrix-prices-moto/main.go --output-dir="$TEMP_DIR"

if [ ! -f "$TEMP_DIR/$FILENAME_MOTO" ]; then
    echo "❌ Ошибка: файл мотошин не создан!"
    exit 1
fi

MOTO_LINES=$(wc -l < "$TEMP_DIR/$FILENAME_MOTO")
echo "  ✅ Файл создан: $FILENAME_MOTO ($MOTO_LINES строк)"
echo ""

# 2. Показываем содержимое (первые 5 строк каждого файла)
echo "📄 Предварительный просмотр файлов:"
echo ""
echo "--- МРЦ (первые 5 строк) ---"
head -5 "$TEMP_DIR/$FILENAME_MRC"
echo ""
echo "--- Легковые шины (первые 5 строк) ---"
head -5 "$TEMP_DIR/$FILENAME_TYRES"
echo ""
echo "--- Мотошины (первые 5 строк) ---"
head -5 "$TEMP_DIR/$FILENAME_MOTO"
echo ""

# 3. Копируем файлы на Desktop для просмотра
echo "📋 Копирование файлов на Desktop для проверки..."
cp "$TEMP_DIR/$FILENAME_MRC" ~/Desktop/
cp "$TEMP_DIR/$FILENAME_TYRES" ~/Desktop/
cp "$TEMP_DIR/$FILENAME_MOTO" ~/Desktop/
echo "✅ Файлы скопированы на Desktop"
echo ""

echo "================================================================"
echo "✅ ТЕСТОВАЯ генерация завершена!"
echo "📊 Созданные файлы:"
echo "   • МРЦ: $FILENAME_MRC ($((MRC_LINES-1)) позиций)"
echo "   • Легковые шины: $FILENAME_TYRES ($((TYRES_LINES-1)) позиций)"
echo "   • Мотошины: $FILENAME_MOTO ($((MOTO_LINES-1)) позиций)"
echo "📁 Файлы доступны на Desktop для проверки"
echo "🕐 Время завершения: $(date '+%Y-%m-%d %H:%M:%S %Z')"
echo "================================================================"
echo ""
echo "⚠️  ВНИМАНИЕ: Это тестовый режим, файлы НЕ загружены на сервер!"
echo "   Для реальной загрузки используйте: ./scripts/upload_prices_to_server.sh"
