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
git clone https://github.com/AlexanderMorozov1919/ElectronicQueue.git

# Frontend
git clone https://github.com/AlexanderMorozov1919/electronicqueue-frontend.git

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
# 🗄️ База данных
DB_USER=postgres            # Имя пользователя для подключения к БД
DB_PASSWORD=1234            # Пароль пользователя для подключения к БД
DB_HOST=localhost           # Адрес сервера базы данных PostgreSQL
DB_PORT=5432                # Порт базы данных PostgreSQL
DB_NAME=el_queue            # Имя базы данных
DB_SSLMODE=disable          # Режим SSL для подключения к БД

# 🌐 Сервер
BACKEND_PORT=8080           # Порт, на котором запускается backend-сервер
FRONTEND_PORT=3000          # Порт, на котором запускается frontend-сервер

# 🚪 Точки входа
FRONTEND_MAINS=main1, main2 # entrypoint-файлы Flutter

# 🔐 Безопасность
JWT_SECRET=your-secret-key  # Секретный ключ для подписи JWT
JWT_EXPIRATION=24h          # Время жизни токена (например, 24h)

# 🎫 Настройки талонов
TICKET_MODE=color           # Режим генерации талона (color | b/w)
TICKET_HEIGHT=800           # Высота талона для печати в пикселях

# 📝 Логи и файлы
LOG_FILE=logs/app.log       # Путь к файлу логов приложения
TICKET_DIR=tickets          # Путь к папке со сгенерированными талонами
```

---

## ⚡ Быстрая установка

```ini
./install.sh [--local] [--docker] [--go] [--go-docker] [--flutter] [--flutter-docker] [--fill] [--rewrite]
```

### 📌 Параметры установщика

| Параметр              | Описание                                                        |
|-----------------------|-----------------------------------------------------------------|
| `--go`                | Сборка и настройка backend на Go (требуется Golang + PostgreSQL) |
| `--go-docker`         | Сборка и настройка backend на Go в Docker (требуется Docker)     |
| `--flutter`           | Сборка и настройка frontend на Flutter (требуется Flutter + Dart)|
| `--flutter-docker`    | Сборка и настройка frontend на Flutter в Docker                 |
| `--local`             | Локальная сборка Go и Flutter                                   |
| `--docker`            | Сборка Go и Flutter в Docker                                    |
| `--fill`              | Заполнение базы тестовыми данными                               |
| `--rewrite`           | Пересоздать базу данных (удалить и создать заново)              |

---

## 🚀 Запуск приложения

```ini
./run.sh [--go|--go-docker] [--flutter|--flutter-docker] [--local|--docker]
```

### ⚙️ Параметры запуска

| Параметр              | Описание                                                        |
|-----------------------|-----------------------------------------------------------------|
| `--go`                | Запуск backend на Go (требуется Golang + PostgreSQL)            |
| `--go-docker`         | Запуск backend на Go в Docker (требуется Docker)                |
| `--flutter`           | Запуск frontend на Flutter (требуется Flutter + Dart)           |
| `--flutter-docker`    | Запуск frontend на Flutter в Docker                             |
| `--local`             | Запуск Go и Flutter локально                                    |
| `--docker`            | Запуск Go и Flutter в Docker                                    |
> ❗️ **Важно**: Дождитесь запуска backend сервера и 5+ секунд после запуска, прежде чем отправлять запросы с frontend сервера

---

## 🧹 Очистка проекта

```ini
./uninstall.sh [--go] [--flutter] [--go-docker] [--go-docker-hard] [--flutter-docker-hard] [--docker-hard]
```

### 🗑️ Параметры очистки

| Параметр                | Действие                                                                 |
|-------------------------|--------------------------------------------------------------------------|
| `--go`                  | Удаляет базу данных PostgreSQL и артефакты сборки backend                |
| `--flutter`             | Очищает проект Flutter (frontend)                                        |
| `--go-docker`           | Останавливает и удаляет контейнеры backend, удаляет volume базы данных   |
| `--go-docker-hard`      | Полная очистка backend: контейнеры, образы, volume, orphans              |
| `--flutter-docker-hard` | Полная очистка frontend: контейнеры, образы, volume, orphans             |
| `--docker-hard`         | Полная очистка backend и frontend в Docker                               |

---

## 🌐 Доступные адреса

| Сервис | URL | Описание |
|--------|-----|----------|
| 🔧 **Backend API** | `http://localhost:{BACKEND_PORT}` | REST API сервер |
| 📚 **Swagger UI** | `http://localhost:{BACKEND_PORT}/swagger/index.html` | Документация API |
| 🖥️ **Терминал** | `http://localhost:{FRONTEND_PORT}` | Интерфейс терминала |
| 📝 **Регистратор** | `http://localhost:{FRONTEND_PORT+1}` | Окно регистратора |

---

## 📚 Документация

Полная документация доступна в **[Swagger UI](http://localhost:8080/swagger/index.html)**

---

## ⚠️ Важно

- Чтобы Docker работал, необходимо запустить **Docker Desktop**.
- Если y Docker возникают ошибки, попробуйте перезагрузить **Docker Desktop**.
- Если появляется ошибка о недоступности API backend сервера *(ClientException: Failed to fetch)*, подождите пока загрузиться **Backend**, **5+ секунд** после загрузки и **обновите страницу**.

---

### 🎉 Готово! Приложение запущено и готово к работе