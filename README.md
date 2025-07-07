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

## Установка

> **Для Backend**: требуется установленный [Golang (1.24.2)](https://go.dev/dl/) и [PostgreSQL](https://www.postgresql.org/download/)

> **Для Frontend**: требуется установленный [Flutter](https://docs.flutter.dev/get-started/install) и [Dart](https://dart.dev/get-dart) <br>
**Через VS Code**: [Flutter + Dart](https://docs.flutter.dev/install/with-vs-code)

#### 1. Клонирование репозиториев

##### Backend
```sh
git clone https://github.com/AlexanderMorozov1919/ElectronicQueue.git

cd ElectronicQueue
```

##### Frontend
```sh
git clone https://github.com/AlexanderMorozov1919/electronicqueue-frontend.git

cd electronicqueue-frontend
```

#### 2. Создание .env файла

```sh
cp .env.example .env
```

#### 3. Настройка переменных окружения

```ini
DB_USER=postgres            # Имя пользователя для подключения к БД
DB_PASSWORD=1234            # Пароль пользователя для подключения к БД
DB_HOST=localhost           # Адрес сервера базы данных PostgreSQL
DB_PORT=5432                # Порт базы данных PostgreSQL
DB_NAME=el_queue            # Имя базы данных
DB_SSLMODE=disable          # Режим SSL для подключения к БД

BACKEND_PORT=8080           # Порт, на котором запускается backend-сервер

JWT_SECRET=your-secret-key  # Секретный ключ для подписи JWT
JWT_EXPIRATION=24h          # Время жизни токена (например, 24h)

LOG_FILE=logs/app.log       # Путь к файлу логов приложения
```

#### 4. Установка (через bash / git bash)

```sh
./install.sh
```

#### 5. Запуск программы

##### Backend (Запуск из директории ElectronicQueue)
```sh
go run cmd/main.go
```

##### Frontend (Запуск из директории electronicqueue-frontend)
```sh
flutter run -d chrome --web-port=XXXX # Порт Frontend сервера
```

#### 6. Удаление (базы данных)

```sh
./uninstall.sh
```