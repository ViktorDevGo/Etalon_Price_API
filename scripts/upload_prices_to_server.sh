#!/bin/bash
set -e

# Конфигурация
REMOTE_HOST="147.45.215.76"
REMOTE_USER="root"
REMOTE_PASSWORD="k7MF4xi99Ty^^T"
REMOTE_PATH="/home/bitrix/www/upload/1c_catalog"
TEMP_DIR="/tmp/bitrix_export"
DATE_STR=$(date +%Y%m%d)
FILENAME_TYRES="Переоценка_шины_${DATE_STR}.csv"
FILENAME_MOTO="Переоценка_мотошины_${DATE_STR}.csv"

# Определяем рабочую директорию (для Docker и локального запуска)
WORK_DIR="${WORK_DIR:-/Users/viktor/Pro_Koleso/Etalon_Price_API}"

echo "================================================================"
echo "Автоматическая выгрузка цен на сервер 1C-Битрикс"
echo "================================================================"
echo "🕐 Время запуска: $(date '+%Y-%m-%d %H:%M:%S %Z')"
echo ""

# Создаем временную директорию
mkdir -p "$TEMP_DIR"

# 1. Генерируем файлы переоценки
echo "📦 Шаг 1: Генерация файлов переоценки..."
cd "$WORK_DIR"

# 1.1 Легковые шины
echo "  → Генерация файла легковых шин..."
go run cmd/export-bitrix-prices/main.go --output-dir="$TEMP_DIR"

if [ ! -f "$TEMP_DIR/$FILENAME_TYRES" ]; then
    echo "❌ Ошибка: файл легковых шин не создан!"
    exit 1
fi

echo "  ✅ Файл создан: $FILENAME_TYRES"

# 1.2 Мотошины
echo "  → Генерация файла мотошин..."
go run cmd/export-bitrix-prices-moto/main.go --output-dir="$TEMP_DIR"

if [ ! -f "$TEMP_DIR/$FILENAME_MOTO" ]; then
    echo "❌ Ошибка: файл мотошин не создан!"
    exit 1
fi

echo "  ✅ Файл создан: $FILENAME_MOTO"
echo ""

# 2. Загружаем файлы на сервер
echo "📤 Шаг 2: Загрузка файлов на сервер $REMOTE_HOST..."

# 2.1 Легковые шины
echo "  → Загрузка файла легковых шин..."
sshpass -p "$REMOTE_PASSWORD" scp -o StrictHostKeyChecking=no \
    "$TEMP_DIR/$FILENAME_TYRES" \
    "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"

if [ $? -eq 0 ]; then
    echo "  ✅ Файл легковых шин загружен: $REMOTE_PATH/$FILENAME_TYRES"
else
    echo "  ❌ Ошибка загрузки файла легковых шин"
    exit 1
fi

# 2.2 Мотошины
echo "  → Загрузка файла мотошин..."
sshpass -p "$REMOTE_PASSWORD" scp -o StrictHostKeyChecking=no \
    "$TEMP_DIR/$FILENAME_MOTO" \
    "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"

if [ $? -eq 0 ]; then
    echo "  ✅ Файл мотошин загружен: $REMOTE_PATH/$FILENAME_MOTO"
else
    echo "  ❌ Ошибка загрузки файла мотошин"
    exit 1
fi

echo ""

# 3. Очистка временных файлов
echo "🧹 Шаг 3: Очистка временных файлов..."
rm -f "$TEMP_DIR/$FILENAME_TYRES"
rm -f "$TEMP_DIR/$FILENAME_MOTO"
echo "✅ Временные файлы удалены"

echo ""
echo "================================================================"
echo "✅ Выгрузка завершена успешно!"
echo "📊 Загружено файлов: 2"
echo "   • Легковые шины: $FILENAME_TYRES"
echo "   • Мотошины: $FILENAME_MOTO"
echo "🕐 Время завершения: $(date '+%Y-%m-%d %H:%M:%S %Z')"
echo "================================================================"
