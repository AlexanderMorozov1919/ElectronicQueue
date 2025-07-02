package main

import (
	"fmt"
	"log"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/database"
	"ElectronicQueue/internal/utils"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}
	fmt.Printf("Environment loaded succesfully\n")

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	fmt.Printf("Successful connect to database: \"%s\"\n", db.Name())

	_, err = utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration)
	if err != nil {
		log.Fatalf("Failed to create JWT manager: %v", err)
	}
	fmt.Printf("JWT manager successfully initialized\n")
}
