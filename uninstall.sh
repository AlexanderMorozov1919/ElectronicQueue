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

cd "$(dirname "$0")" && docker compose down > /dev/null 2>&1
cd "$(dirname "$0")/../electronicqueue-frontend" && docker compose down > /dev/null 2>&1
cd - > /dev/null > /dev/null 2>&1

# --- Аргументы ---
if [[ $# -eq 0 ]]; then
  echo "Usage: $0 [--go] [--flutter] [--go-docker] [--go-docker-hard] [--flutter-docker-hard] [--docker-hard]"
  exit 1
fi
MODE=""
for arg in "$@"; do
  case $arg in
    --go)
      MODE="go"
      ;;
    --flutter)
      MODE="flutter"
      ;;
    --go-docker)
      MODE="go-docker"
      ;;
    --go-docker-hard)
      MODE="go-docker-hard"
      ;;
    --flutter-docker-hard)
      MODE="flutter-docker-hard"
      ;;
    --docker-hard)
      MODE="docker-hard"
      ;;
    *)
      echo "Usage: $0 [--go] [--flutter] [--go-docker] [--go-docker-hard] [--flutter-docker-hard] [--docker-hard]"
      exit 1
      ;;
  esac
done

if [[ "$MODE" == "go" ]]; then
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
  if find . -type f -name "*.exe" -delete; then
    echo "Removed all .exe files."
  else
    echo "Failed to remove .exe files."
    exit 1
  fi

  echo "Uninstall complete."
fi

if [[ "$MODE" == "flutter" ]]; then
  cd "$(dirname "$0")/../electronicqueue-frontend" || { echo "Failed to change directory to electronicqueue-frontend."; exit 1; }
  echo "Cleaning Flutter project..."
  if flutter clean; then
    echo "Flutter project cleaned."
  else
    echo "Failed to clean Flutter project."
    exit 1
  fi
  cd - > /dev/null
fi

if [[ "$MODE" == "go-docker" ]]; then
  echo "Stopping and removing containers, removing volume..."
  if ! docker info > /dev/null 2>&1; then
    echo "Make sure that Docker Engine is running."
    exit 1
  fi
  if docker compose down; then
    if docker volume rm electronicqueue_db-data; then
      echo "Docker containers stopped and volume removed."
    else
      echo "Failed to remove docker volume electronicqueue_db-data."
      exit 1
    fi
  else
    echo "Failed to stop and remove docker containers."
    exit 1
  fi
fi

if [[ "$MODE" == "go-docker-hard" ]]; then
  echo "Full cleanup: containers, images, volumes, orphans..."
  if ! docker info > /dev/null 2>&1; then
    echo "Make sure that Docker Engine is running. Try restarting Docker Desktop."
    exit 1
  fi
  if docker compose down --rmi all --volumes --remove-orphans; then
    echo "Full docker cleanup complete."
  else
    echo "Failed to perform full docker cleanup."
    exit 1
  fi
fi

if [[ "$MODE" == "flutter-docker-hard" ]]; then
  echo "Full cleanup for Flutter: containers, images, volumes, orphans..."
  if ! docker info > /dev/null 2>&1; then
    echo "Make sure that Docker Engine is running. Try restarting Docker Desktop."
    exit 1
  fi
  cd "$(dirname "$0")/../electronicqueue-frontend" || { echo "Failed to change directory to electronicqueue-frontend."; exit 1; }
  if docker compose down --rmi all --volumes --remove-orphans; then
    echo "Full Flutter docker cleanup complete."
  else
    echo "Failed to perform full Flutter docker cleanup."
    exit 1
  fi
  cd - > /dev/null
fi

if [[ "$MODE" == "docker-hard" ]]; then
  "$0" --go-docker-hard || exit 1
  "$0" --flutter-docker-hard || exit 1
  echo "Full docker cleanup for both backend and frontend complete."
fi

exit 0