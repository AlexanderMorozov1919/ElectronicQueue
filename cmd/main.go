package main

import (
	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/database"
	"ElectronicQueue/internal/logger"
	"fmt"
)

func main() {

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Ошибка загрузки конфига: %v\n", err)
		return
	}
	// Инициализация логгера
	logger.Init(cfg.LogFile)
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Ошибка синхронизации логов: %v\n", err)
		}
	}()

	log := logger.Default()

	log.Info("Application starting...")
	log.WithField("version", "1.0.0").Info("Configuration loaded")

	// 3. Тестируем логирование
	log.Info("Тестовый запуск логгера")
	log.WithField("example", true).Warn("Предупреждение с дополнительным полем")

	// 4. Имитация ошибки
	log.WithField("config", cfg).Error("Пример ошибки")

	// Проверка записи в файл
	log.Info("Проверка лог-файла в папке logs/")

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.WithError(err).Fatal("Database connection failed")
	}
	log.WithField("dbname", db.Name()).Info("Database connected successfully")

	fmt.Printf("Конфиг логгера:\nФайл: %s\n", cfg.LogFile)

}
