package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/database"
	"ElectronicQueue/internal/handlers"
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
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

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// GIN middleware для логирования всех запросов
	r.Use(func(c *gin.Context) {
		fmt.Printf("[GIN] %s %s\n", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

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

				var ticket models.Ticket
				if err := json.Unmarshal([]byte(n.Extra), &ticket); err != nil {
					log.WithError(err).Error("Failed to unmarshal notification")
					return true
				}

				c.SSEvent("message", models.TicketResponse{
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

	// Инициализация репозитория, сервиса и хендлера для талонов
	ticketRepo := repository.NewTicketRepository(db)
	ticketService := services.NewTicketService(ticketRepo)
	ticketHandler := handlers.NewTicketHandler(ticketService)

	// Группа эндпоинтов для работы с талонами
	tickets := r.Group("/api/tickets")
	{
		tickets.GET("/services", ticketHandler.GetAvailableServices)
		tickets.POST("/next-step", ticketHandler.GetNextStep)
		tickets.POST("/confirm", ticketHandler.ConfirmAction)
		tickets.POST("/", ticketHandler.CreateTicketHandler) // legacy, если нужно
	}

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

	fmt.Printf("Сервер запущен на порту: %s\n", cfg.BackendPort)
	if err := r.Run(":" + cfg.BackendPort); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}
