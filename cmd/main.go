package main

import (

	"encoding/json"

	"ElectronicQueue/internal/logger"


	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/database"

	"ElectronicQueue/internal/models/ticket_model"

	"github.com/lib/pq"

	"ElectronicQueue/internal/handlers"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/services"

	_ "ElectronicQueue/docs"

	"github.com/gin-gonic/gin"

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
	log.Info("Тестовый запуск логгера")
	log.WithField("example", true).Warn("Предупреждение с дополнительным полем")
	log.WithField("config", cfg).Error("Пример ошибки")
	log.Info("Проверка лог-файла в папке logs/")


	// Подключение к базе данных
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.WithError(err).Fatal("Database connection failed")
	}
	log.WithField("dbname", cfg.DBName).Info("Database connected successfully")

	// Подключение для LISTEN/NOTIFY
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)
	listener := pq.NewListener(dsn, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.WithError(err).WithField("event", ev).Error("Listener error")
		} else {
			log.WithField("event", ev).Info("Listener event")
		}
	})

	if err := listener.Ping(); err != nil {
		log.WithError(err).Fatal("Failed to ping database listener")
	}
	log.Info("Global listener started")

	if err := listener.Listen("ticket_update"); err != nil {
		log.WithError(err).Fatal("Failed to listen to ticket_update channel")
	}
	log.Info("Listening to ticket_update channel")

	// Настройка GIN
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(logger.GinLogger())

	// SSE endpoint
	r.GET("/tickets", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// SSE поток
		c.Stream(func(w io.Writer) bool {
			select {
			case n := <-listener.Notify:
				if n == nil {
					log.Info("Received nil notification")
					return true
				}
				log.WithField("payload", n.Extra).Info("Received notification")

				var ticket ticket_model.Ticket
				if err := json.Unmarshal([]byte(n.Extra), &ticket); err != nil {
					log.WithError(err).Error("Failed to unmarshal notification")
					return true
				}

				c.SSEvent("message", ticket_model.TicketResponse{
					ID:           ticket.ID,
					TicketNumber: ticket.TicketNumber,
					Status:       ticket.Status,
					CreatedAt:    ticket.CreatedAt,
				})
				return true

			case <-c.Request.Context().Done():
				log.Info("Client disconnected (SSE)")
				return false
			}
		})
	})

	// Обработка сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info("Received shutdown signal, closing...")

		if err := logger.Sync(); err != nil {
			fmt.Printf("Ошибка синхронизации логов: %v\n", err)
		}

		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				fmt.Printf("Ошибка закрытия базы данных: %v\n", err)
			}
		}

		if err := listener.Close(); err != nil {
			log.WithError(err).Error("Ошибка при закрытии pq.Listener")
		}

		os.Exit(0)
	}()


	fmt.Printf("Сервер запущен на порту: %s\n", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}

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
