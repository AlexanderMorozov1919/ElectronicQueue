package main

import (
	"fmt"
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
	fmt.Printf("Environment loaded succesfully\n")

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	fmt.Printf("Successful connect to database: \"%s\"\n", db.Name())
}
