#!/bin/bash

# Загружаем переменные из .env
set -o allexport
source .env
set +o allexport

# Проверяем наличие обязательных переменных
: "${DB_USER?Need to set DB_USER}"
: "${DB_PASSWORD?Need to set DB_PASSWORD}"
: "${DB_HOST?Need to set DB_HOST}"
: "${DB_PORT?Need to set DB_PORT}"
: "${DB_NAME?Need to set DB_NAME}"

# Удаляем базу данных
echo "Dropping database '$DB_NAME'..."
if PGPASSWORD="$DB_PASSWORD" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d postgres -c "DROP DATABASE IF EXISTS \"$DB_NAME\";"; then
    echo "Database '$DB_NAME' dropped successfully."
else
    echo "Failed to drop database '$DB_NAME'."
    exit 1
fi

# Удаляем исполняемые файлы
echo "Cleaning up build artifacts..."
find . -type f -name "*.exe" -delete
echo "Removed all .exe files."

echo "Uninstall complete."
read -p "Нажмите Enter для выхода..."
exit