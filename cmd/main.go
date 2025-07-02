package main

import (
	"log"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/database"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction(zap.AddCaller(), zap.Fields(
		zap.String("app", "electronic_queue"),
	))
	defer logger.Sync()
	sugar := logger.Sugar()

	cfg, err := config.LoadConfig()
	if err != nil {
		sugar.Fatalf("Config error: %v", err) // Логирование через Zap
	}

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	logger.Info("Database connected", zap.String("name", db.Name()))

}
