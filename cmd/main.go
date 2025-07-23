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
	"gorm.io/gorm"

	_ "ElectronicQueue/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	defer pool.Close()

	go listenForNotifications(listenerCtx, pool, notificationChannel, log)

	r := setupRouter(psBroker, db, cfg)

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Graceful shutdown handler
	handleGracefulShutdown(db, cancelListener, log)

	log.Info(fmt.Sprintf("Сервер запущен на порту: %s", cfg.BackendPort))
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

	// --- Инициализация репозиториев ---
	repo := repository.NewRepository(db)

	// --- Инициализация сервисов ---
	ticketService := services.NewTicketService(repo.Ticket, repo.Service)
	doctorService := services.NewDoctorService(repo.Ticket, repo.Doctor, repo.Schedule, broker)
	authService := services.NewAuthService(repo.Registrar, repo.Doctor, jwtManager)
	databaseService := services.NewDatabaseService(repository.NewDatabaseRepository(db))
	patientService := services.NewPatientService(repo.Patient)
	appointmentService := services.NewAppointmentService(repo.Appointment)
	cleanupService := services.NewCleanupService(repo.Cleanup)
	tasksTimerService := services.NewTasksTimerService(cleanupService, cfg)
	scheduleService := services.NewScheduleService(repo.Schedule, repo.Doctor)

	// Запускаем планировщик задач в фоне
	go tasksTimerService.Start(context.Background())

	// --- Инициализация обработчиков ---
	ticketHandler := handlers.NewTicketHandler(ticketService, cfg)
	doctorHandler := handlers.NewDoctorHandler(doctorService, broker)
	registrarHandler := handlers.NewRegistrarHandler(ticketService)
	authHandler := handlers.NewAuthHandler(authService)
	databaseHandler := handlers.NewDatabaseHandler(databaseService)
	audioHandler := handlers.NewAudioHandler()
	patientHandler := handlers.NewPatientHandler(patientService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService)

	// SSE-эндпоинт для табло очереди регистратуры
	r.GET("/tickets", sseHandler(broker, "reception_sse"))

	// SSE-эндпоинт для экрана у кабинета врача
	r.GET("/api/doctor/screen-updates/:cabinet_number", doctorHandler.DoctorScreenUpdates)

	// Группа для аутентификации и создания пользователей
	auth := r.Group("/api/auth")
	{
		auth.POST("/login/registrar", authHandler.LoginRegistrar)
		auth.POST("/login/doctor", authHandler.LoginDoctor)
	}

	admin := r.Group("/api/admin").Use(middleware.RequireAPIKey(cfg.InternalAPIKey))
	{
		admin.POST("/create/doctor", authHandler.CreateDoctor)
		admin.POST("/create/registrar", authHandler.CreateRegistrar)
		admin.DELETE("/tickets/:id", registrarHandler.DeleteTicket)
		admin.POST("/schedules", scheduleHandler.CreateSchedule)
		admin.DELETE("/schedules/:id", scheduleHandler.DeleteSchedule)
	}

	// Группа для терминала (получение талонов)
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

	// Публичная группа для информации о врачах и расписании
	publicDoctorGroup := r.Group("/api/doctor")
	{
		publicDoctorGroup.GET("/active", doctorHandler.GetAllActiveDoctors)
		publicDoctorGroup.GET("/cabinets/active", doctorHandler.GetActiveCabinets)
		publicDoctorGroup.GET("/schedule/:doctor_id", appointmentHandler.GetDoctorSchedule)

	}

	// Группа для генерации аудио
	audioGroup := r.Group("/api/audio")
	{
		audioGroup.GET("/announce", audioHandler.GenerateAnnouncement)
	}

	// API для работы с базой данных (защищено статическим API-ключом)
	dbAPI := r.Group("/api/database").Use(middleware.RequireAPIKey(cfg.ExternalAPIKey))
	{
		dbAPI.POST("/:table/select", databaseHandler.GetData)
		dbAPI.POST("/:table/insert", databaseHandler.InsertData)
		dbAPI.PATCH("/:table/update", databaseHandler.UpdateData)
		dbAPI.DELETE("/:table/delete", databaseHandler.DeleteData)
	}

	// Действия, доступные только врачу
	protectedDoctorGroup := r.Group("/api/doctor").Use(middleware.RequireRole(jwtManager, "doctor"))
	{
		protectedDoctorGroup.GET("/tickets/registered", doctorHandler.GetRegisteredTickets)
		protectedDoctorGroup.GET("/tickets/in-progress", doctorHandler.GetInProgressTickets)
		protectedDoctorGroup.POST("/start-appointment", doctorHandler.StartAppointment)
		protectedDoctorGroup.POST("/complete-appointment", doctorHandler.CompleteAppointment)
		protectedDoctorGroup.POST("/start-break", doctorHandler.StartBreak)
		protectedDoctorGroup.POST("/end-break", doctorHandler.EndBreak)
		protectedDoctorGroup.POST("/set-active", doctorHandler.SetDoctorActive)
		protectedDoctorGroup.POST("/set-inactive", doctorHandler.SetDoctorInactive)
	}

	// Действия, доступные только регистратору
	registrar := r.Group("/api/registrar").Use(middleware.RequireRole(jwtManager, "registrar"))
	{
		registrar.POST("/call-next", registrarHandler.CallNext)
		registrar.PATCH("/tickets/:id/status", registrarHandler.UpdateStatus)
		registrar.GET("/patients/search", patientHandler.SearchPatients)
		registrar.POST("/patients", patientHandler.CreatePatient)
		registrar.POST("/appointments", appointmentHandler.CreateAppointment)
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

// handleGracefulShutdown корректно завершает работу приложения
func handleGracefulShutdown(db *gorm.DB, cancel context.CancelFunc, log *logger.AsyncLogger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Получен сигнал завершения, закрываю соединения...")

		cancel()

		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				log.WithError(err).Error("Ошибка при закрытии соединения с БД")
			} else {
				log.Info("Соединение с базой данных закрыто.")
			}
		}

		if err := logger.Sync(); err != nil {
			fmt.Printf("Ошибка синхронизации логов: %v\n", err)
		}

		os.Exit(0)
	}()
}
