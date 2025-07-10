#!/bin/bash

# Загружаем переменные из .env
set -o allexport
source ./.env
set +o allexport

# Пути до main файлов Flutter frontend
FLUTTER_MAIN_FILES=(
  "lib/terminal/main_terminal.dart"
  "lib/registry_window/main_registry.dart"
  # Добавьте сюда другие main-файлы по необходимости
)

# Kill processes on BACKEND_PORT, FRONTEND_PORT, FRONTEND_PORT+N
kill_by_port() {
  local PORT=$1
  if command -v lsof >/dev/null 2>&1; then
    local PID
    PID=$(lsof -ti tcp:$PORT)
    if [ -n "$PID" ]; then
      echo "Killing process on port $PORT (PID $PID) via lsof"
      kill -9 $PID
      sleep 1
    fi
  elif command -v fuser >/dev/null 2>&1; then
    local PID
    PID=$(fuser $PORT/tcp 2>/dev/null)
    if [ -n "$PID" ]; then
      echo "Killing process on port $PORT (PID $PID) via fuser"
      kill -9 $PID
      sleep 1
    fi
  fi
  # Windows: netstat + taskkill
  if [[ "$(uname -s 2>/dev/null)" =~ (MINGW|MSYS|CYGWIN|Windows_NT) ]]; then
    local PIDS
    PIDS=$(netstat -ano | grep :$PORT | awk '{print $5}' | sort | uniq)
    for PID in $PIDS; do
      if [ -n "$PID" ]; then
        echo "Killing process on port $PORT (PID $PID) via taskkill"
        taskkill //PID $PID //F 2>/dev/null
        sleep 1
      fi
    done
  fi
}

if [[ $# -eq 0 ]]; then
  echo "Usage: $0 [--go|--go-docker] [--flutter]"
  exit 1
fi
RUN_GO=""
RUN_GO_DOCKER=""
RUN_FLUTTER=""
for arg in "$@"; do
  case $arg in
    --go)
      RUN_GO="yes"
      ;;
    --go-docker)
      RUN_GO_DOCKER="yes"
      ;;
    --flutter)
      RUN_FLUTTER="yes"
      ;;
    *)
      echo "Usage: $0 [--go|--go-docker] [--flutter]"
      exit 1
      ;;
  esac
done

# Kill processes on BACKEND_PORT, FRONTEND_PORT, FRONTEND_PORT+N only if needed
if { [[ "$RUN_GO" == "yes" && -n "$BACKEND_PORT" ]] || [[ "$RUN_GO_DOCKER" == "yes" && -n "$BACKEND_PORT" ]]; }; then
  kill_by_port $BACKEND_PORT
fi
if [[ "$RUN_FLUTTER" == "yes" && -n "$FRONTEND_PORT" ]]; then
  N=${#FLUTTER_MAIN_FILES[@]}
  for ((i=0; i<N; i++)); do
    PORT=$((FRONTEND_PORT+i))
    kill_by_port $PORT
  done
fi

# Run Go backend
if [[ "$RUN_GO" == "yes" ]]; then
  echo "Starting Go backend (main.exe)..."
  if (cd "$(dirname "$0")" && go run cmd/main.go &); then
    echo "Go backend started."
  else
    echo "Failed to start Go backend."
    exit 1
  fi
elif [[ "$RUN_GO_DOCKER" == "yes" ]]; then
  echo "Starting Go backend via Docker Compose..."
  if (cd "$(dirname "$0")" && docker compose up &); then
    echo "Go backend (Docker Compose) started."
  else
    echo "Failed to start Go backend (Docker Compose)."
    exit 1
  fi
fi

# Run Flutter frontend
if [[ "$RUN_FLUTTER" == "yes" ]]; then
  if cd "$(dirname "$0")/../electronicqueue-frontend" && flutter pub get --no-example; then
    echo "Flutter setup complete."
  else
    echo "Failed to fetch packages."
    exit 1
  fi
  if [ -z "$FRONTEND_PORT" ]; then
    echo "FRONTEND_PORT is not set in .env!"
    exit 1
  fi
  PORT=$FRONTEND_PORT
  for MAIN_FILE in "${FLUTTER_MAIN_FILES[@]}"; do
    echo "Starting Flutter frontend: $MAIN_FILE on port $PORT..."
    (cd "$(dirname "$0")/../electronicqueue-frontend" && flutter run -t "$MAIN_FILE" -d chrome --web-port=$PORT &)
    PORT=$((PORT+1))
  done
  echo "All Flutter frontends started."
fi
