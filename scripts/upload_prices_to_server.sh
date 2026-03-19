#!/bin/bash
set -e

# Конфигурация (из переменных окружения или значения по умолчанию)
REMOTE_HOST="${BITRIX_REMOTE_HOST:-147.45.215.76}"
REMOTE_USER="${BITRIX_REMOTE_USER:-root}"
REMOTE_PASSWORD="${BITRIX_REMOTE_PASSWORD:-k7MF4xi99Ty^^T}"
REMOTE_PATH="${BITRIX_REMOTE_PATH:-/home/bitrix/www/upload/1c_catalog}"
TEMP_DIR="/tmp/bitrix_export"
DATE_STR=$(date +%Y%m%d)
FILENAME_MRC="Переоценка_МРЦ_${DATE_STR}.csv"
FILENAME_TYRES="Переоценка_шины_${DATE_STR}.csv"
FILENAME_MOTO="Переоценка_мотошины_${DATE_STR}.csv"
FILENAME_RIMS="Переоценка_диски_${DATE_STR}.csv"

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

# ВАЖНО: Порядок генерации:
# 1) МРЦ (НЕ обновляет isimport) - захватывает все данные
# 2) Легковые шины (обновляет isimport = 1)
# 3) Мотошины (обновляет isimport = 1)
# 4) Диски (обновляет isimport = 1)

# 1.1 МРЦ (ПЕРВЫМ!)
echo "  → Генерация файла МРЦ..."
go run cmd/export-bitrix-prices-mrc/main.go --output-dir="$TEMP_DIR"

if [ ! -f "$TEMP_DIR/$FILENAME_MRC" ]; then
    echo "❌ Ошибка: файл МРЦ не создан!"
    exit 1
fi

echo "  ✅ Файл создан: $FILENAME_MRC"

# 1.2 Легковые шины
echo "  → Генерация файла легковых шин..."
go run cmd/export-bitrix-prices/main.go --output-dir="$TEMP_DIR"

if [ ! -f "$TEMP_DIR/$FILENAME_TYRES" ]; then
    echo "❌ Ошибка: файл легковых шин не создан!"
    exit 1
fi

echo "  ✅ Файл создан: $FILENAME_TYRES"

# 1.3 Мотошины
echo "  → Генерация файла мотошин..."
go run cmd/export-bitrix-prices-moto/main.go --output-dir="$TEMP_DIR"

if [ ! -f "$TEMP_DIR/$FILENAME_MOTO" ]; then
    echo "❌ Ошибка: файл мотошин не создан!"
    exit 1
fi

echo "  ✅ Файл создан: $FILENAME_MOTO"

# 1.4 Диски
echo "  → Генерация файла дисков..."
go run cmd/export-bitrix-prices-rims/main.go --output-dir="$TEMP_DIR"

if [ ! -f "$TEMP_DIR/$FILENAME_RIMS" ]; then
    echo "❌ Ошибка: файл дисков не создан!"
    exit 1
fi

echo "  ✅ Файл создан: $FILENAME_RIMS"
echo ""

# 2. Загружаем файлы на сервер
echo "📤 Шаг 2: Загрузка файлов на сервер $REMOTE_HOST..."

# 2.1 МРЦ
echo "  → Загрузка файла МРЦ..."
sshpass -p "$REMOTE_PASSWORD" scp -o StrictHostKeyChecking=no \
    "$TEMP_DIR/$FILENAME_MRC" \
    "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"

if [ $? -eq 0 ]; then
    echo "  ✅ Файл МРЦ загружен: $REMOTE_PATH/$FILENAME_MRC"
else
    echo "  ❌ Ошибка загрузки файла МРЦ"
    exit 1
fi

# 2.2 Легковые шины
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

# 2.3 Мотошины
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

# 2.4 Диски
echo "  → Загрузка файла дисков..."
sshpass -p "$REMOTE_PASSWORD" scp -o StrictHostKeyChecking=no \
    "$TEMP_DIR/$FILENAME_RIMS" \
    "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"

if [ $? -eq 0 ]; then
    echo "  ✅ Файл дисков загружен: $REMOTE_PATH/$FILENAME_RIMS"
else
    echo "  ❌ Ошибка загрузки файла дисков"
    exit 1
fi

echo ""

# 3. Очистка временных файлов
echo "🧹 Шаг 3: Очистка временных файлов..."
rm -f "$TEMP_DIR/$FILENAME_MRC"
rm -f "$TEMP_DIR/$FILENAME_TYRES"
rm -f "$TEMP_DIR/$FILENAME_MOTO"
rm -f "$TEMP_DIR/$FILENAME_RIMS"
echo "✅ Временные файлы удалены"

echo ""
echo "================================================================"
echo "✅ Выгрузка завершена успешно!"
echo "📊 Загружено файлов: 4"
echo "   • МРЦ: $FILENAME_MRC"
echo "   • Легковые шины: $FILENAME_TYRES"
echo "   • Мотошины: $FILENAME_MOTO"
echo "   • Диски: $FILENAME_RIMS"
echo "🕐 Время завершения: $(date '+%Y-%m-%d %H:%M:%S %Z')"
echo "================================================================"
