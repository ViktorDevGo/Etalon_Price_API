#!/bin/bash
set -e

# Конфигурация
REMOTE_HOST="147.45.215.76"
REMOTE_USER="root"
REMOTE_PASSWORD="k7MF4xi99Ty^^T"
REMOTE_PATH="/home/bitrix/www/upload/1c_catalog"
TEMP_DIR="/tmp/bitrix_export"
DATE_STR=$(date +%Y%m%d)
FILENAME="Переоценка_шины_${DATE_STR}.csv"

# Определяем рабочую директорию (для Docker и локального запуска)
WORK_DIR="${WORK_DIR:-/Users/viktor/Pro_Koleso/Etalon_Price_API}"

echo "================================================================"
echo "Автоматическая выгрузка цен на сервер 1C-Битрикс"
echo "================================================================"
echo "🕐 Время запуска: $(date '+%Y-%m-%d %H:%M:%S %Z')"
echo ""

# Создаем временную директорию
mkdir -p "$TEMP_DIR"

# 1. Генерируем файл переоценки
echo "📦 Шаг 1: Генерация файла переоценки..."
cd "$WORK_DIR"
go run cmd/export-bitrix-prices/main.go --output-dir="$TEMP_DIR"

if [ ! -f "$TEMP_DIR/$FILENAME" ]; then
    echo "❌ Ошибка: файл не создан!"
    exit 1
fi

echo "✅ Файл создан: $TEMP_DIR/$FILENAME"
echo ""

# 2. Загружаем файл на сервер
echo "📤 Шаг 2: Загрузка на сервер $REMOTE_HOST..."

# Используем sshpass для автоматической передачи пароля
sshpass -p "$REMOTE_PASSWORD" scp -o StrictHostKeyChecking=no \
    "$TEMP_DIR/$FILENAME" \
    "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"

if [ $? -eq 0 ]; then
    echo "✅ Файл успешно загружен на сервер"
    echo "📁 Путь: $REMOTE_HOST:$REMOTE_PATH/$FILENAME"
else
    echo "❌ Ошибка загрузки на сервер"
    exit 1
fi

echo ""

# 3. Очистка временных файлов
echo "🧹 Шаг 3: Очистка временных файлов..."
rm -f "$TEMP_DIR/$FILENAME"
echo "✅ Временные файлы удалены"

echo ""
echo "================================================================"
echo "✅ Выгрузка завершена успешно!"
echo "🕐 Время завершения: $(date '+%Y-%m-%d %H:%M:%S %Z')"
echo "================================================================"
