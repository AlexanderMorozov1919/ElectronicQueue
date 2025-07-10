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

# На случай если оно там случайно осталось
rm -f migrations/fill_db.sql

# --- Аргументы ---
if [[ $# -eq 0 ]]; then
  echo "Usage: $0 [--go] [--go-docker] [--flutter] [--fill] [--rewrite]"
  exit 1
fi
GO_MODE=""
FLUTTER_MODE=""
GO_DOCKER_MODE=""
FILL=""
REWRITE=""
for arg in "$@"; do
  case $arg in
    --go)
      GO_MODE="true"
      ;;
    --go-docker)
      GO_DOCKER_MODE="true"
      ;;
    --flutter)
      FLUTTER_MODE="true"
      ;;
    --fill)
      FILL="true"
      ;;
    --rewrite)
      REWRITE="true"
      ;;
    *)
      echo "Usage: $0 [--go] [--go-docker] [--flutter] [--fill] [--rewrite]"
      exit 1
      ;;
  esac
done

if [[ "$GO_MODE" == "true" ]]; then
  if [[ "$REWRITE" == "true" ]]; then
    # Удаляем базу данных
    echo "Dropping database '$DB_NAME'..."
    if PGPASSWORD="$DB_PASSWORD" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d postgres -c "DROP DATABASE IF EXISTS \"$DB_NAME\";"; then
        echo "Database '$DB_NAME' dropped successfully."
    else
        echo "Failed to drop database '$DB_NAME'."
        exit 1
    fi
  fi

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

  if [[ "$FILL" == "true" ]]; then
    echo "Filling database from fill_db.sql..."
    export PGPASSWORD="$DB_PASSWORD"
    export PGCLIENTENCODING="UTF8"
    if ! psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -p "$DB_PORT" -f fill_db.sql; then
      echo "Failed to fill database from fill_db.sql."
      exit 1
    fi
  fi

  # Устанавливаем модули проекта
  echo "Downloading Go modules..."
  if go mod download; then
    echo "Go modules downloaded successfully."
  else
    echo "Failed to download Go modules."
    exit 1
  fi
  
  # Обновление документации
  echo "Updating Swagger documentation..."
  if ! swag init --dir ./cmd,./internal --output ./docs; then
    echo "Failed to update Swagger documentation."
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

  echo "Golang setup complete."
fi

if [[ "$FLUTTER_MODE" == "true" ]]; then
  cd "$(dirname "$0")/../electronicqueue-frontend"
  echo "Getting Flutter packages..."
  if flutter pub get --no-example; then
    echo "Flutter setup complete."
  else
    echo "Failed to fetch packages."
    exit 1
  fi
fi

if [[ "$GO_DOCKER_MODE" == "true" ]]; then
  if ! docker info > /dev/null 2>&1; then
    echo "Make sure that Docker Engine is running."
    exit 1
  fi
  if [[ "$REWRITE" == "true" ]]; then
    docker compose down
    docker volume rm electronicqueue_db-data
  fi
  if [[ "$FILL" == "true" ]]; then
    echo "Copying fill_db.sql to migrations/ for Docker..."
    if ! cp fill_db.sql migrations/; then
      echo "Failed to copy fill_db.sql to migrations/."
      exit 1
    fi
  fi
  echo "Running Docker Compose build..."
  if ! docker compose build; then
    echo "Docker Compose build failed."
    [[ "$FILL" == "true" ]] && rm -f migrations/fill_db.sql
    exit 1
  fi
  echo "Running Docker Compose migrations..."
  if ! docker compose up -d db; then
    echo "migrations failed."
    [[ "$FILL" == "true" ]] && rm -f migrations/fill_db.sql
    exit 1
  fi
  echo "Waiting for database to be ready..."
  docker compose exec db pg_isready -U ${DB_USER} -d ${DB_NAME}
  
  echo "Database is ready. Migrations applied automatically."
  docker compose down

  rm -f migrations/fill_db.sql
  echo "Docker setup complete."
fi

exit 0