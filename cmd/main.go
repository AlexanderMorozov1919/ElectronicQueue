package main

import (
	"fmt"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/database"

	//_ "ElectronicQueue/internal/handler" // Раскомментируем, когда создадим обработчики
	//_ "ElectronicQueue/internal/service" // Раскомментируем, когда создадим сервисы
	_ "ElectronicQueue/internal/repository"

	_ "ElectronicQueue/docs"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

// @title ElectronicQueue API
// @version 1.0
// @description Это сервер для электронной очереди

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Инициализация логгера
	logger, _ := zap.NewProduction(zap.AddCaller(), zap.Fields(
		zap.String("app", "electronic_queue"),
	))
	defer logger.Sync()
	sugar := logger.Sugar()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		sugar.Fatalf("Config error: %v", err)
	}
	fmt.Printf("Environment loaded succesfully\n")

	// Подключение к базе данных
	db, err := database.ConnectDB(cfg)
	if err != nil {
		sugar.Fatalf("Database connection error: %v", err)
	}
	sugar.Infof("Successful connect to database: \"%s\"\n", db.Name())

	// Инициализация роутера Gin
	router := gin.Default()

	// Запуск сервера
	sugar.Infof("Starting server on %s", "localhost:8080") // Заменить на cfg.ServerAddress
	if err := router.Run(); err != nil {                   // router.Run(cfg.ServerAddress)
		sugar.Fatalf("Failed to run server: %v", err)
	}
}
