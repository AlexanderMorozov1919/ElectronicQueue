package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ElectronicQueue/internal/config"
)

// ConnectDB устанавливает соединение с базой данных PostgreSQL через GORM
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	// Инициализируем логгер
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// Логируем ошибку через Zap
		logger.Error("Failed to connect to database",
			zap.String("dbname", cfg.DBName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Логируем успешное подключение через Zap
	logger.Info("Database connection established",
		zap.String("dbname", cfg.DBName),
		zap.String("host", cfg.DBHost),
	)

	return db, nil
}
