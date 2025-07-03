package database

import (
	"fmt"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB устанавливает соединение с базой данных PostgreSQL через GORM
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	log := logger.Default() // Используем логгер с module=default для подключения

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.NewGORMLogger(), // Используем GORMLogger с module=gorm
	})
	if err != nil {
		log.WithError(err).WithField("dbname", cfg.DBName).Error("Failed to connect to database")
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.WithFields(map[string]interface{}{
		"dbname": cfg.DBName,
		"host":   cfg.DBHost,
	}).Info("Database connection established")

	return db, nil
}
