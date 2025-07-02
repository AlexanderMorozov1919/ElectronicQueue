package main

import (
	"fmt"
	"log"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/database"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	fmt.Printf("Successful connect to database: \"%s\"\n", db.Name())
}
