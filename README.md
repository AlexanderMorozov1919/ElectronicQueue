# **ElectronicQueue - Сервис электронной очереди**

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white&style=for-the-badge" alt="Go"></a>
  <a href="https://gorm.io/"><img src="https://img.shields.io/badge/GORM-FFCA28?logo=go&logoColor=black&style=for-the-badge" alt="GORM"></a>
  <a href="https://gin-gonic.com/"><img src="https://img.shields.io/badge/Gin-00B386?logo=go&logoColor=white&style=for-the-badge" alt="Gin"></a>
  <a href="https://jwt.io/"><img src="https://img.shields.io/badge/JWT-000000?logo=jsonwebtokens&logoColor=white&style=for-the-badge" alt="JWT"></a>
  <a href="https://www.postgresql.org/"><img src="https://img.shields.io/badge/PostgreSQL-4169E1?logo=postgresql&logoColor=white&style=for-the-badge" alt="PostgreSQL"></a>
  <a href="https://www.docker.com/"><img src="https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white&style=for-the-badge" alt="Docker"></a>
  <a><img src="https://img.shields.io/badge/REST%20API-FF6F00?logo=rest&logoColor=white&style=for-the-badge" alt="REST API"></a>
  <a href="https://swagger.io/"><img src="https://img.shields.io/badge/Swagger-85EA2D?logo=swagger&logoColor=black&style=for-the-badge" alt="Swagger"></a>
</p>

<p align="center">
  <img src="assets/img/ticket_example.png" alt="Пример талона" width="350"/>
</p>

## 📋 Системные требования

### Docker (Деплой)
- **Docker** `1.24.2+` — [Скачать](https://docs.docker.com/desktop/)

### Backend (Локальная разработка)
- **Go** `1.24.2+` — [Скачать](https://go.dev/dl/)
- **PostgreSQL** `15+` — [Скачать](https://www.postgresql.org/download/)

### Frontend (Локальная разработка)
- **Flutter** `3.32.5+` — [Скачать](https://docs.flutter.dev/get-started/install)
- **Dart** `3.8.1+` — [Скачать](https://dart.dev/get-dart)

> 💡 **Совет**: [VS Code с расширением Flutter + Dart](https://docs.flutter.dev/install/with-vs-code)

---

## 📦 Установка

### 1️⃣ Клонирование репозиториев

```bash
# Backend
git clone -b develop https://github.com/AlexanderMorozov1919/ElectronicQueue.git

# Frontend
git clone -b develop https://github.com/AlexanderMorozov1919/electronicqueue-frontend.git

# Главный каталог
cd ElectronicQueue
```

### 2️⃣ Конфигурация окружения

```bash
cp .env.example .env
```

### 3️⃣ Настройка переменных

Отредактируйте файл `.env`:

```ini
DB_USER=postgres
DB_PASSWORD=1234
DB_HOST=localhost
DB_PORT=5432
DB_NAME=el_queue
DB_SSLMODE=disable

BACKEND_PORT=8081
FRONTEND_PORT=3000
API_BASE_URL=http://localhost:8081
BROWSER=chrome

JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

LOG_DIR=logs
TICKET_DIR=tickets

INTERNAL_API_KEY=iak12345
EXTERNAL_API_KEY=eak12345

TICKET_MODE=b/w
PRINTER_PAPER_WIDTH_MM=80
PRINTER_DPI=200
PRINTER_NAME=ThermalPrinter

ONE_C_URL=http://localhost:8081/umc/ru/hs/electronicqueue
ONE_C_API_KEY="..."
ONE_C_CONF=1C.conf

MAINTENANCE_TIME=00:00
BACKGROUND_MUSIC=false
```

---

## ⚡ Быстрая установка

```ini
./install [local] [docker] [go] [go-docker] [flutter] [flutter-docker] [fill] [rewrite]
```

### 📌 Параметры установщика

| Параметр              | Описание                                                          |
|-----------------------|-------------------------------------------------------------------|
| `go`                  | Сборка и настройка backend на Go (требуется Golang + PostgreSQL)  |
| `go-docker`           | Сборка и настройка backend на Go в Docker (требуется Docker)      |
| `flutter`             | Сборка и настройка frontend на Flutter (требуется Flutter + Dart) |
| `flutter-docker`      | Сборка и настройка frontend на Flutter в Docker (требуется Docker)|
| `local`               | Локальная сборка Go и Flutter                                     |
| `docker`              | Сборка Go и Flutter в Docker                                      |
| `fill`                | Заполнение базы данных тестовыми значениями                       |
| `rewrite`             | Пересоздать базу данных (удалить и создать заново)                |

---

## 🚀 Запуск приложения

```ini
./run [go|go-docker] [flutter|flutter-docker] [local|docker]
```

### ⚙️ Параметры запуска

| Параметр              | Описание                                                        |
|-----------------------|-----------------------------------------------------------------|
| `go`                  | Запуск backend на Go (требуется Golang + PostgreSQL)            |
| `go-docker`           | Запуск backend на Go в Docker (требуется Docker)                |
| `flutter`             | Запуск frontend на Flutter (требуется Flutter + Dart)           |
| `flutter-docker`      | Запуск frontend на Flutter в Docker (требуется Docker)          |
| `local`               | Запуск Go и Flutter локально                                    |
| `docker`              | Запуск Go и Flutter в Docker                                    |


## 📦 Запуск из архивов

Если вы не хотите собирать проект, можно воспользоваться готовыми архивами:

- **backend_build.zip** — находится в репозитории `ElectronicQueue`
- **frontend_build.zip** — находится в репозитории `electronicqueue-frontend`

### 1️⃣ Backend

1. Распакуйте архив `backend_build.zip` в отдельную папку.
    ```
    unzip backend_build.zip -d backend_build
    ```
2. Перейдите в папку с распакованным backend.
    ```
    cd backend_build
    ```
2. Запустите backend, используя бинарник для вашей ОС:
   **Windows**:  
    ```
    ./windows-amd64.exe
    ```
   **Linux**:  
    ```
    ./linux-amd64
    ```
   **macOS**:  
    ```
    ./darwin-amd64
    ```
    Backend будет доступен по адресу: `http://localhost:{BACKEND_PORT}`

### 2️⃣ Frontend

1. Распакуйте архив `frontend_build.zip` в отдельную папку
    ```
    unzip frontend_build.zip -d frontend_build
    ```
2. Перейдите в папку с распакованным frontend
    ```
    cd frontend_build
    ```
3. Запустите локальный сервер командой:
   ```
   python -m http.server {FRONTEND_PORT}
   ```
   Frontend будет доступен по адресу: `http://localhost:{FRONTEND_PORT}/{имя_сервиса}/`

## 🧹 Очистка проекта

```ini
./uninstall [go] [go-docker] [flutter] [flutter-docker] [local] [docker]
```

### 🗑️ Параметры очистки

| Параметр                | Действие                                                                 |
|-------------------------|--------------------------------------------------------------------------|
| `go`                    | Удаляет базу данных PostgreSQL и артефакты сборки Go                     |
| `go-docker`             | Полная очистка backend: контейнеры, образы, volume, orphans              |
| `flutter`               | Очищает проект Flutter                                                   |
| `flutter-docker`        | Полная очистка frontend: контейнеры, образы, volume, orphans             |
| `local`                 | Очистка локального окружения, объединяющая очистку Go и Flutter          |
| `docker`                | Полная очистка backend и frontend в Docker                               |

---

## 🌐 Доступные адреса

| Сервис                        | URL                                                  | Описание                  |
|-------------------------------|------------------------------------------------------|---------------------------|
| 🔧 **Backend API**            | `http://localhost:{BACKEND_PORT}`                    | REST API сервер           |
| 📚 **Swagger UI**             | `http://localhost:{BACKEND_PORT}/swagger/index.html` | Документация API          |
| 👨‍⚕️ **Доктор**                 | `http://localhost:{FRONTEND_PORT}`                   | Окно доктора              |
| 🏥 **Очередь к доктору**      | `http://localhost:{FRONTEND_PORT+1}`                 | Очередь к доктору         |
| 🔳 **Очередь к регистратору** | `http://localhost:{FRONTEND_PORT+2}`                 | Электронное табло очереди |
| 📝 **Регистратор**            | `http://localhost:{FRONTEND_PORT+3}`                 | Окно регистратора         |
| 🖥️ **Терминал**               | `http://localhost:{FRONTEND_PORT+4}`                 | Интерфейс терминала       |
| 📅 **Расписание**             | `http://localhost:{FRONTEND_PORT+5}`                 | Расписание врачей         |

---

## 📚 Документация

Полная документация доступна в **[Swagger UI](https://elqueue.iwtcode.com/swagger/index.html)**

---

## ⚠️ Важно

- Чтобы Docker работал, необходимо запустить **Docker Desktop**.
- Если y Docker возникают ошибки, попробуйте перезагрузить **Docker Desktop**.

---

### 🎉 Готово! Приложение запущено и готово к работе