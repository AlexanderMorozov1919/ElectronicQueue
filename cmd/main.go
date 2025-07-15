package main

import (
	"context"
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

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "ElectronicQueue/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gorm.io/gorm"
)

// @title Electronic Queue API
// @version 1.0
// @description API для системы электронной очереди
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Ошибка загрузки конфига: %v\n", err)
		return
	}

	// Инициализация логгера
	logger.Init(cfg.LogDir)
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Ошибка синхронизации логов: %v\n", err)
		}
	}()
	log := logger.Default()

	// Подключение к базе данных через GORM
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.WithError(err).Fatal("Database connection failed")
	}
	log.WithField("dbname", cfg.DBName).Info("Database connected successfully")

	// Канал для передачи уведомлений из листенера в SSE хендлеры
	notificationChannel := make(chan string)
	// Контекст для управления жизненным циклом листенера
	listenerCtx, cancelListener := context.WithCancel(context.Background())

	pool, err := initListener(listenerCtx, cfg, log, notificationChannel)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize database listener with pgx")
	}

	// Настройка роутера
	r := setupRouter(notificationChannel, db, cfg)

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Обработка сигналов завершения
	handleGracefulShutdown(db, pool, cancelListener, log)

	fmt.Printf("Сервер запущен на порту: %s\n", cfg.BackendPort)
	if err := r.Run(":" + cfg.BackendPort); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}

// initListener инициализирует LISTEN/NOTIFY через pgx
func initListener(ctx context.Context, cfg *config.Config, log *logger.AsyncLogger, notifications chan<- string) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	// Запускаем горутину, которая будет слушать уведомления
	go func() {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			log.WithError(err).Error("Failed to acquire connection from pool for listener")
			return
		}
		defer conn.Release()

		_, err = conn.Exec(ctx, "LISTEN ticket_update")
		if err != nil {
			log.WithError(err).Error("Failed to execute LISTEN command")
			return
		}
		log.Info("Listening to 'ticket_update' channel with pgx")

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Info("Listener context cancelled, shutting down.")
					return
				}
				log.WithError(err).Error("Error waiting for notification")
				time.Sleep(5 * time.Second)
				continue
			}
			notifications <- notification.Payload
		}
	}()

	return pool, nil
}

// setupRouter настраивает маршруты и middleware
func setupRouter(notifications <-chan string, db *gorm.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(logger.GinLogger())
	r.Use(middleware.CorsMiddleware())

	r.GET("/tickets", sseHandler(notifications))

	ticketRepo := repository.NewTicketRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	ticketService := services.NewTicketService(ticketRepo, serviceRepo)
	ticketHandler := handlers.NewTicketHandler(ticketService, cfg)
	doctorService := services.NewDoctorService(ticketRepo)
	doctorHandler := handlers.NewDoctorHandler(doctorService)
	registrarHandler := handlers.NewRegistrarHandler(ticketService)

	tickets := r.Group("/api/tickets")
	{
		tickets.GET("/start", ticketHandler.StartPage)
		tickets.GET("/services", ticketHandler.Services)
		tickets.GET("/active", ticketHandler.GetAllActive)
		tickets.POST("/print/selection", ticketHandler.Selection)
		tickets.POST("/print/confirmation", ticketHandler.Confirmation)
		tickets.GET("/download/:ticket_number", ticketHandler.DownloadTicket)
		tickets.GET("/view/:ticket_number", ticketHandler.ViewTicket)
	}

	doctor := r.Group("/api/doctor")
	{
		doctor.POST("/start-appointment", doctorHandler.StartAppointment)
		doctor.POST("/complete-appointment", doctorHandler.CompleteAppointment)
	}

	registrar := r.Group("/api/registrar")
	{
		registrar.POST("/call-next", registrarHandler.CallNext)
		registrar.PATCH("/tickets/:id/status", registrarHandler.UpdateStatus)
		registrar.DELETE("/tickets/:id", registrarHandler.DeleteTicket)
	}

	databaseRepo := repository.NewDatabaseRepository(db)
	databaseService := services.NewDatabaseService(databaseRepo)
	databaseHandler := handlers.NewDatabaseHandler(databaseService)

	dbAPI := r.Group("/api/database").Use(middleware.RequireAPIKey(cfg.ExternalAPIKey))
	{
		dbAPI.POST("/:table/select", databaseHandler.GetData)
		dbAPI.POST("/:table/insert", databaseHandler.InsertData)
		dbAPI.PATCH("/:table/update", databaseHandler.UpdateData)
		dbAPI.DELETE("/:table/delete", databaseHandler.DeleteData)
	}

	return r
}

type NotificationPayload struct {
	Action string        `json:"action"`
	Data   models.Ticket `json:"data"`
}

func sseHandler(notifications <-chan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		log := logger.Default()

		c.Stream(func(w io.Writer) bool {
			select {
			case payloadStr := <-notifications:
				log.WithField("payload", payloadStr).Info("Received notification from channel")
				var payload NotificationPayload
				if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
					log.WithError(err).Error("Failed to unmarshal notification payload")
					return true
				}
				c.SSEvent(payload.Action, payload.Data.ToResponse())
				return true
			case <-c.Request.Context().Done():
				log.Info("Client disconnected (SSE)")
				return false
			}
		})
	}
}

func handleGracefulShutdown(db *gorm.DB, pool *pgxpool.Pool, cancel context.CancelFunc, log *logger.AsyncLogger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received shutdown signal, closing...")

		cancel()

		if pool != nil {
			pool.Close()
			log.Info("pgx listener pool closed.")
		}

		if err := logger.Sync(); err != nil {
			fmt.Printf("Ошибка синхронизации логов: %v\n", err)
		}

		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				fmt.Printf("Ошибка закрытия базы данных: %v\n", err)
			}
		}

		os.Exit(0)
	}()
}
