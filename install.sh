#!/bin/bash

# Загружаем переменные из .env
set -o allexport
source ./.env
set +o allexport

# Проверяем наличие обязательных переменных
: "${DB_USER?Need to set DB_USER}"
: "${DB_PASSWORD?Need to set DB_PASSWORD}"
: "${DB_HOST?Need to set DB_HOST}"
: "${DB_PORT?Need to set DB_PORT}"
: "${DB_NAME?Need to set DB_NAME}"
: "${DB_SSLMODE:=disable}"

echo "Checking if database '$DB_NAME' exists..."

# Пытаемся подключиться к БД для проверки
if PGPASSWORD="$DB_PASSWORD" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    echo "Database '$DB_NAME' already exists."
else
    echo "Database '$DB_NAME' does not exist. Creating..."
    if PGPASSWORD="$DB_PASSWORD" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d postgres -c "CREATE DATABASE \"$DB_NAME\""; then
        echo "Database '$DB_NAME' created successfully."
    else
        echo "Failed to create database '$DB_NAME'."
        exit 1
    fi
fi

# Устанавливаем migrate, если ее нет или обновляем до последней версии
echo "Installing/updating migrate CLI tool..."
if go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; then
  echo "migrate CLI installed/updated successfully."
else
  echo "Failed to install migrate CLI. Make sure Go is installed and configured correctly."
  exit 1
fi

# Проверяем, доступна ли команда migrate
if ! command -v migrate &> /dev/null
then
    echo "'migrate' command could not be found."
    echo "Please ensure $(go env GOPATH)/bin or $HOME/go/bin is in your PATH."
    exit 1
fi

# Применяем миграции
echo "Running database migrations..."
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

if migrate -path "./migrations" -database "${DATABASE_URL}" up; then
  echo "Migrations applied successfully."
else
  echo "Failed to apply migrations."
  exit 1
fi

export PGPASSWORD="$DB_PASSWORD"
export PGCLIENTENCODING="UTF8"
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -p "$DB_PORT" -f migrations/fill_DB.sql

# Устанавливаем модули проекта
echo "Downloading Go modules..."
if go mod download; then
  echo "Go modules downloaded successfully."
else
  echo "Failed to download Go modules."
  exit 1
fi

# Собираем программу
echo "Building Go application..."
if go build -o main.exe cmd/main.go; then
  echo "Build successful. Binary: ./main.exe"
else
  echo "Build failed."
  exit 1
fi

echo "Setup complete."
exit 0