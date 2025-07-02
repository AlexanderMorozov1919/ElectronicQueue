# **ElectronicQueue - Сервис электронной очереди**

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white&style=for-the-badge" alt="Go"></a>
  <a href="https://gorm.io/"><img src="https://img.shields.io/badge/GORM-FFCA28?logo=go&logoColor=black&style=for-the-badge" alt="GORM"></a>
  <a href="https://gin-gonic.com/"><img src="https://img.shields.io/badge/Gin-00B386?logo=go&logoColor=white&style=for-the-badge" alt="Gin"></a>
  <a href="https://www.postgresql.org/"><img src="https://img.shields.io/badge/PostgreSQL-4169E1?logo=postgresql&logoColor=white&style=for-the-badge" alt="PostgreSQL"></a>
  <a><img src="https://img.shields.io/badge/REST%20API-FF6F00?logo=rest&logoColor=white&style=for-the-badge" alt="REST API"></a>
  <a href="https://swagger.io/"><img src="https://img.shields.io/badge/Swagger-85EA2D?logo=swagger&logoColor=black&style=for-the-badge" alt="Swagger"></a>
</p>

## Установка

#### 1. Создание .env файла

```sh
cp .env.example .env
```

#### 2. Настройка переменных окружения

```ini
DB_USER=postgres            # Имя пользователя для подключения к БД
DB_PASSWORD=1234            # Пароль пользователя для подключения к БД
DB_HOST=localhost           # Адрес сервера базы данных PostgreSQL
DB_PORT=5432                # Порт базы данных PostgreSQL
DB_NAME=el_queue            # Имя базы данных
DB_SSLMODE=disable          # Режим SSL для подключения к БД
SERVER_PORT=8080            # Порт, на котором запускается сервер
```

#### 3. Установка (через bash / git bash)

```sh
./install.sh
```

#### 4. Запуск программы

```sh
./main.exe
```
или
```sh
go run cmd/main.go
```

#### 5. Удаление (базы данных)

```sh
./uninstall.sh
```