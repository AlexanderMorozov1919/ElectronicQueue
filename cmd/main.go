package main

import (
	"ElectronicQueue/internal/logger"

	"fmt"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/database"
	"ElectronicQueue/internal/handlers"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/services"

	_ "ElectronicQueue/docs"

	"github.com/gin-gonic/gin"
)

// @title ElectronicQueue API
// @version 1.0
// @description Это сервер для электронной очереди

// @host localhost:8080
// @BasePath /api/v1
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
	defer logger.Sync()

	log := logger.Default()

	// Подключение к базе данных
	db, err := database.ConnectDB(cfg)
	if err != nil {

		log.WithError(err).Fatal("Database connection failed")
	}
	log.WithField("dbname", db.Name()).Info("Database connected successfully")

	fmt.Printf("Конфиг логгера:\nФайл: %s\n", cfg.LogFile)

	r := gin.Default()

	r.LoadHTMLFiles("frontend/print_ticket.html", "frontend/display_ticket.html")

	repo := repository.NewRepository(db)

	ticketService := services.NewTicketService(repo.Ticket)
	ticketHandler := handlers.NewTicketHandler(ticketService)

	// Регистрация роутов терминала
	r.GET("/terminal/service", ticketHandler.GetServicePage)
	r.GET("/terminal/service/select", ticketHandler.GetSelectServicePage)
	r.POST("/terminal/service/make_appointment", ticketHandler.HandleService("make_appointment"))
	r.POST("/terminal/service/confirm_appointment", ticketHandler.HandleService("confirm_appointment"))
	r.POST("/terminal/service/lab_tests", ticketHandler.HandleService("lab_tests"))
	r.POST("/terminal/service/documents", ticketHandler.HandleService("documents"))
	r.GET("/terminal/service/print_ticket", ticketHandler.HandlePrintTicketPage)
	r.GET("/terminal/service/display_ticket", ticketHandler.HandleDisplayTicketPage)
	r.POST("/terminal/service/display_ticket", ticketHandler.HandleDisplayTicketPost)

	r.Run(":" + cfg.BackendPort)
}
