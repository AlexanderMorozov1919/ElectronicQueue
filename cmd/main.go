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
	"ElectronicQueue/internal/middleware"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/services"

	_ "ElectronicQueue/docs"

	ginSwaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"gorm.io/gorm"
)

// @title Electronic Queue API
// @version 1.0
// @description API для системы электронной очереди
// @host localhost:8080
// @BasePath /
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

	// Инициализация listener для LISTEN/NOTIFY
	listener, err := initListener(cfg, log)
	if err != nil {
		log.WithError(err).Error("Failed to initialize database listener")
	}
	defer listener.Close()

	// Настройка роутера
	r := setupRouter(listener, db, cfg)

	// Обработка сигналов завершения
	handleGracefulShutdown(db, listener, log)

	fmt.Printf("Сервер запущен на порту: %s\n", cfg.BackendPort)
	if err := r.Run(":" + cfg.BackendPort); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}

// initListener инициализирует LISTEN/NOTIFY для PostgreSQL
func initListener(cfg *config.Config, log *logger.AsyncLogger) (*pq.Listener, error) {
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
		return nil, err
	}
	if err := listener.Listen("ticket_update"); err != nil {
		return nil, err
	}
	log.Info("Listening to ticket_update channel")
	return listener, nil
}

// setupRouter настраивает маршруты и middleware
func setupRouter(listener *pq.Listener, db *gorm.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(logger.GinLogger())

	// CORS middleware
	r.Use(middleware.CorsMiddleware())

	// GIN middleware для логирования всех запросов
	r.Use(middleware.RequestLogger())

	// SSE endpoint
	r.GET("/tickets", sseHandler(listener))

	// Инициализация репозитория, сервиса и хендлера для талонов
	ticketRepo := repository.NewTicketRepository(db)
	ticketService := services.NewTicketService(ticketRepo)
	ticketHandler := handlers.NewTicketHandler(ticketService, cfg)

	// Инициализация хендлера для регистратора
	registrarHandler := handlers.NewRegistrarHandler(ticketService)

	// Группа эндпоинтов для рабочего места регистратора
	registrar := r.Group("/api/registrar")
	{
		registrar.POST("/call-next", registrarHandler.CallNext)
		// tickets.PATCH("/:id/status", ticketHandler.UpdateStatus)
		// tickets.DELETE("/:id", ticketHandler.DeleteTicket)
	}

	// Группа эндпоинтов для работы с талонами
	tickets := r.Group("/api/tickets")
	{
		tickets.GET("/start", ticketHandler.StartPage)
		tickets.GET("/services", ticketHandler.Services)
		tickets.POST("/print/selection", ticketHandler.Selection)
		tickets.POST("/print/confirmation", ticketHandler.Confirmation)
	}

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(ginSwaggerFiles.Handler))

	return r
}

// sseHandler возвращает SSE endpoint для обновлений талонов
func sseHandler(listener *pq.Listener) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		log := logger.Default()
		c.Stream(func(w io.Writer) bool {
			select {
			case n := <-listener.Notify:
				if n == nil {
					log.Info("Received nil notification")
					return true
				}
				log.WithField("payload", n.Extra).Info("Received notification from DB")

				var ticket models.Ticket
				if err := json.Unmarshal([]byte(n.Extra), &ticket); err != nil {
					log.WithError(err).Error("Failed to unmarshal notification payload")
					return true
				}

				// Преобразуем Ticket в TicketResponse с помощью нашего нового метода
				response := ticket.ToResponse()

				c.SSEvent("message", response)
				return true
			case <-c.Request.Context().Done():
				log.Info("Client disconnected (SSE)")
				return false
			}
		})
	}
}

// handleGracefulShutdown обрабатывает завершение работы приложения
func handleGracefulShutdown(db *gorm.DB, listener *pq.Listener, log *logger.AsyncLogger) {
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
}
