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
	"ElectronicQueue/internal/pubsub"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/services"
	"ElectronicQueue/internal/utils"

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

	logger.Init(cfg.LogDir)
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Ошибка синхронизации логов: %v\n", err)
		}
	}()
	log := logger.Default()

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.WithError(err).Fatal("Database connection failed")
	}
	log.WithField("dbname", cfg.DBName).Info("Database connected successfully")

	notificationChannel := make(chan string, 100)
	listenerCtx, cancelListener := context.WithCancel(context.Background())

	psBroker := pubsub.NewBroker()
	go psBroker.ListenAndPublish(notificationChannel)

	pool, err := initPgxPool(listenerCtx, cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize database listener with pgx")
	}

	go listenForNotifications(listenerCtx, pool, notificationChannel, log)

	r := setupRouter(psBroker, db, cfg)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	handleGracefulShutdown(db, pool, cancelListener, log)

	fmt.Printf("Сервер запущен на порту: %s\n", cfg.BackendPort)
	if err := r.Run(":" + cfg.BackendPort); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}

// initPgxPool инициализирует пул соединений pgx
func initPgxPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
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
	return pool, nil
}

// listenForNotifications слушает LISTEN/NOTIFY и отправляет в указанный канал
func listenForNotifications(ctx context.Context, pool *pgxpool.Pool, notifications chan<- string, log *logger.AsyncLogger) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.WithError(err).Error("Listener: Failed to acquire connection from pool")
		return
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN ticket_update")
	if err != nil {
		log.WithError(err).Error("Listener: Failed to execute LISTEN command")
		return
	}
	log.Info("Listener: Listening to 'ticket_update' channel")

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Info("Listener context cancelled, shutting down.")
				return
			}
			log.WithError(err).Error("Listener: Error waiting for notification")
			time.Sleep(5 * time.Second)
			continue
		}
		notifications <- notification.Payload
	}
}

// setupRouter настраивает маршруты и middleware
func setupRouter(broker *pubsub.Broker, db *gorm.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(logger.GinLogger())
	r.Use(middleware.CorsMiddleware())

	jwtManager, err := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration)
	if err != nil {
		logger.Default().WithError(err).Fatal("Failed to initialize JWT Manager")
	}

	// --- Инициализация всех репозиториев ---
	repo := repository.NewRepository(db)

	// --- Инициализация всех сервисов ---
	ticketService := services.NewTicketService(repo.Ticket, repo.Service)
	doctorService := services.NewDoctorService(repo.Ticket, repo.Doctor)
	authService := services.NewAuthService(repo.Registrar, jwtManager)
	databaseService := services.NewDatabaseService(repository.NewDatabaseRepository(db)) // Для универсального API
	patientService := services.NewPatientService(repo.Patient)
	appointmentService := services.NewAppointmentService(repo.Appointment)

	// --- Инициализация всех обработчиков ---
	ticketHandler := handlers.NewTicketHandler(ticketService, cfg)
	doctorHandler := handlers.NewDoctorHandler(doctorService, broker)
	registrarHandler := handlers.NewRegistrarHandler(ticketService)
	authHandler := handlers.NewAuthHandler(authService)
	databaseHandler := handlers.NewDatabaseHandler(databaseService)
	audioHandler := handlers.NewAudioHandler()
	patientHandler := handlers.NewPatientHandler(patientService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)

	// SSE-эндпоинт для табло очереди регистратуры
	r.GET("/tickets", sseHandler(broker, "reception_sse"))

	// --- Определение групп маршрутов ---
	auth := r.Group("/api/auth")
	{
		auth.POST("/login/registrar", authHandler.LoginRegistrar)
		auth.POST("/create/registrar", authHandler.CreateRegistrar)
	}

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

	doctorGroup := r.Group("/api/doctor")
	{
		doctorGroup.GET("/active", doctorHandler.GetAllActiveDoctors)

		// Маршруты для окна врача
		doctorGroup.GET("/tickets/registered", doctorHandler.GetRegisteredTickets)
		doctorGroup.GET("/tickets/in-progress", doctorHandler.GetInProgressTickets)
		doctorGroup.POST("/start-appointment", doctorHandler.StartAppointment)
		doctorGroup.POST("/complete-appointment", doctorHandler.CompleteAppointment)
		doctorGroup.GET("/screen-updates", doctorHandler.DoctorScreenUpdates)
	}

	// Группа для регистратора, защищенная JWT токеном
	registrar := r.Group("/api/registrar").Use(middleware.RequireRole(jwtManager, "registrar"))
	{
		// Основные действия регистратора
		registrar.POST("/call-next", registrarHandler.CallNext)
		registrar.PATCH("/tickets/:id/status", registrarHandler.UpdateStatus)
		registrar.DELETE("/tickets/:id", registrarHandler.DeleteTicket)

		// Новые маршруты для формы записи на прием
		registrar.GET("/patients/search", patientHandler.SearchPatients)
		registrar.POST("/patients", patientHandler.CreatePatient)
		registrar.GET("/schedules/doctor/:doctor_id", appointmentHandler.GetDoctorSchedule)
		registrar.POST("/appointments", appointmentHandler.CreateAppointment)
	}

	dbAPI := r.Group("/api/database").Use(middleware.RequireAPIKey(cfg.ExternalAPIKey))
	{
		dbAPI.POST("/:table/select", databaseHandler.GetData)
		dbAPI.POST("/:table/insert", databaseHandler.InsertData)
		dbAPI.PATCH("/:table/update", databaseHandler.UpdateData)
		dbAPI.DELETE("/:table/delete", databaseHandler.DeleteData)
	}

	audioGroup := r.Group("/api/audio")
	{
		audioGroup.GET("/announce", audioHandler.GenerateAnnouncement)
	}

	return r
}

type NotificationPayload struct {
	Action string                `json:"action"`
	Data   models.TicketResponse `json:"data"`
}

// sseHandler подписывает клиента на события от брокера
func sseHandler(broker *pubsub.Broker, handlerID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		log := logger.Default().WithField("handler_id", handlerID)

		clientChan := broker.Subscribe()
		defer broker.Unsubscribe(clientChan)

		c.Stream(func(w io.Writer) bool {
			select {
			case payloadStr, ok := <-clientChan:
				if !ok {
					log.Info("Client channel closed.")
					return false
				}

				log.WithField("payload", payloadStr).Info("SSE Handler: Sending message to client")
				var payload NotificationPayload
				if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
					log.WithError(err).Error("Failed to unmarshal notification payload")
					return true
				}

				c.SSEvent(payload.Action, payload.Data)
				return true

			case <-c.Request.Context().Done():
				log.Info("Client disconnected.")
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

		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				fmt.Printf("Ошибка закрытия базы данных: %v\n", err)
			} else {
				log.Info("Database connection closed.")
			}
		}

		os.Exit(0)
	}()
}
