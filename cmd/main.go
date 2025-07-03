package main

import (
	"fmt"
	"log"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/database"
	"ElectronicQueue/internal/handlers"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/services"

	"github.com/gin-gonic/gin"
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
		sugar.Fatalf("Config error: %v", err)
	}
	fmt.Printf("Environment loaded succesfully\n")

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	fmt.Printf("Successful connect to database: \"%s\"\n", db.Name())

	repos := repository.NewRepository(db)
	ticketService := services.NewTicketsService(repos.Ticket, db)
	ticketHandler := handlers.NewTicketsHandler(ticketService)

	r := gin.Default()
	v1 := r.Group("/api/v1/registrar")
	{
		v1.GET("/queue", ticketHandler.GetQueueHandler)
		v1.POST("/tickets/call-next", ticketHandler.CallNextTicketHandler)
	}

	r.Run(":" + cfg.ServerPort)
}
