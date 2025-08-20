package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
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

	repo := repository.NewRepository(db)
	processService, err := services.NewBusinessProcessService(repo.BusinessProcess)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize Business Process Service")
	}

	notificationChannel := make(chan string, 100)
	listenerCtx, cancelListener := context.WithCancel(context.Background())

	psBroker := pubsub.NewBroker()
	go psBroker.ListenAndPublish(notificationChannel)

	pool, err := initPgxPool(listenerCtx, cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize database listener with pgx")
	}

	go listenForNotifications(listenerCtx, pool, notificationChannel, log)

	r := setupRouter(psBroker, db, cfg, processService)

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
		log.WithError(err).Error("Listener: Failed to execute LISTEN command for ticket_update")
		return
	}
	log.Info("Listener: Listening to 'ticket_update' channel")

	_, err = conn.Exec(ctx, "LISTEN schedule_update")
	if err != nil {
		log.WithError(err).Error("Listener: Failed to execute LISTEN command for schedule_update")
		return
	}
	log.Info("Listener: Listening to 'schedule_update' channel")

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
		log.WithField("channel", notification.Channel).Info("Listener: Received notification")
		notifications <- notification.Payload
	}
}

// setupRouter настраивает маршруты и middleware
func setupRouter(broker *pubsub.Broker, db *gorm.DB, cfg *config.Config, processService *services.BusinessProcessService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(logger.GinLogger())
	r.Use(middleware.CorsMiddleware())

	jwtManager, err := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration)
	if err != nil {
		logger.Default().WithError(err).Fatal("Failed to initialize JWT Manager")
	}

	repo := repository.NewRepository(db)

	ticketService := services.NewTicketService(repo.Ticket, repo.Service, repo.ReceptionLog, repo.Patient, repo.Appointment, repo.RegistrarPriority)
	doctorService := services.NewDoctorService(repo.Ticket, repo.Doctor, repo.Schedule, broker)
	authService := services.NewAuthService(repo.Registrar, repo.Doctor, repo.Administrator, jwtManager)
	databaseService := services.NewDatabaseService(repository.NewDatabaseRepository(db))
	patientService := services.NewPatientService(repo.Patient)
	appointmentService := services.NewAppointmentService(repo.Appointment, repo.Ticket)
	cleanupService := services.NewCleanupService(repo.Cleanup)
	tasksTimerService := services.NewTasksTimerService(cleanupService, cfg)
	scheduleService := services.NewScheduleService(repo.Schedule, repo.Doctor)
	adService := services.NewAdService(repo.Ad)
	registrarService := services.NewRegistrarService(repo.RegistrarPriority, repo.Service)

	go tasksTimerService.Start(context.Background())

	ticketHandler := handlers.NewTicketHandler(ticketService, cfg)
	doctorHandler := handlers.NewDoctorHandler(doctorService, broker)
	registrarHandler := handlers.NewRegistrarHandler(ticketService, registrarService)
	authHandler := handlers.NewAuthHandler(authService)
	databaseHandler := handlers.NewDatabaseHandler(databaseService)
	audioHandler := handlers.NewAudioHandler(cfg)
	patientHandler := handlers.NewPatientHandler(patientService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService, broker)
	processHandler := handlers.NewBusinessProcessHandler(processService)
	adHandler := handlers.NewAdHandler(adService)

	r.GET("/tickets", middleware.CheckBusinessProcess(processService, "reception"), sseHandler(broker, "reception_sse"))

	r.GET("/api/doctor/screen-updates/:cabinet_number", middleware.CheckBusinessProcess(processService, "queue_doctor"), doctorHandler.DoctorScreenUpdates)

	auth := r.Group("/api/auth")
	{
		auth.POST("/login/registrar", authHandler.LoginRegistrar)
		auth.POST("/login/doctor", authHandler.LoginDoctor)
		auth.POST("/login/administrator", authHandler.LoginAdministrator)
	}

	admin := r.Group("/api/admin").Use(middleware.RequireAPIKey(cfg.InternalAPIKey))
	{
		admin.POST("/create/doctor", authHandler.CreateDoctor)
		admin.POST("/create/registrar", authHandler.CreateRegistrar)
		admin.DELETE("/tickets/:id", registrarHandler.DeleteTicket)
		admin.POST("/schedules", scheduleHandler.CreateSchedule)
		admin.DELETE("/schedules/:id", scheduleHandler.DeleteSchedule)
		admin.POST("/create/administrator", authHandler.CreateAdministrator)
		admin.GET("/processes", processHandler.GetAllProcesses)
		admin.PATCH("/processes/:name", processHandler.UpdateProcess)

		admin.GET("/ads", adHandler.GetAllAds)
		admin.POST("/ads", adHandler.CreateAd)
		admin.GET("/ads/:id", adHandler.GetAdByID)
		admin.PATCH("/ads/:id", adHandler.UpdateAd)
		admin.DELETE("/ads/:id", adHandler.DeleteAd)
	}

	tickets := r.Group("/api/tickets").Use(middleware.CheckBusinessProcess(processService, "terminal"))
	{
		tickets.GET("/start", ticketHandler.StartPage)
		tickets.GET("/services", ticketHandler.Services)
		tickets.POST("/print/selection", ticketHandler.Selection)
		tickets.POST("/print/confirmation", ticketHandler.Confirmation)
		tickets.POST("/appointment/phone", ticketHandler.CheckInByPhone)
		tickets.GET("/download/:ticket_number", ticketHandler.DownloadTicket)
		tickets.GET("/view/:ticket_number", ticketHandler.ViewTicket)
	}
	r.GET("/api/tickets/active", middleware.CheckBusinessProcess(processService, "reception"), ticketHandler.GetAllActive)

	publicDoctorGroup := r.Group("/api/doctor").Use(middleware.CheckBusinessProcess(processService, "registry", "queue_doctor"))
	{
		publicDoctorGroup.GET("/active", doctorHandler.GetAllActiveDoctors)
		publicDoctorGroup.GET("/cabinets/active", doctorHandler.GetActiveCabinets)
	}

	protectedDoctorGroup := r.Group("/api/doctor").
		Use(middleware.RequireRole(jwtManager, "doctor")).
		Use(middleware.CheckBusinessProcess(processService, "doctor"))
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

	registrar := r.Group("/api/registrar").
		Use(middleware.RequireRole(jwtManager, "registrar")).
		Use(middleware.CheckBusinessProcess(processService, "registry"))
	{
		registrar.POST("/call-next", registrarHandler.CallNext)
		registrar.POST("/call-specific", registrarHandler.CallSpecific)
		registrar.GET("/tickets", registrarHandler.GetTickets)
		registrar.GET("/tickets/current", registrarHandler.GetCurrentTicket)
		registrar.PATCH("/tickets/:id/status", registrarHandler.UpdateStatus)
		registrar.GET("/patients/search", patientHandler.SearchPatients)
		registrar.POST("/patients", patientHandler.CreatePatient)
		registrar.GET("/schedules/doctor/:doctor_id", appointmentHandler.GetDoctorSchedule)
		registrar.POST("/appointments", appointmentHandler.CreateAppointment)
		registrar.GET("/patients/:patient_id/appointments", appointmentHandler.GetPatientAppointments)
		registrar.DELETE("/appointments/:id", appointmentHandler.DeleteAppointment)
		registrar.PATCH("/appointments/:id/confirm", appointmentHandler.ConfirmAppointment)
		registrar.GET("/reports/daily", registrarHandler.GetDailyReport)
		registrar.GET("/services", registrarHandler.GetAllServices)
		registrar.GET("/priorities", registrarHandler.GetPriorities)
		registrar.POST("/priorities", registrarHandler.SetPriorities)
	}

	dbAPI := r.Group("/api/database").
		Use(middleware.RequireAPIKey(cfg.ExternalAPIKey)).
		Use(middleware.CheckBusinessProcess(processService, "database"))
	{
		dbAPI.POST("/:table/select", databaseHandler.GetData)
		dbAPI.POST("/:table/insert", databaseHandler.InsertData)
		dbAPI.PATCH("/:table/update", databaseHandler.UpdateData)
		dbAPI.DELETE("/:table/delete", databaseHandler.DeleteData)
	}

	processes := r.Group("/api/processes")
	{
		processes.GET("/:name", processHandler.GetProcessStatusByName)
	}

	adGroup := r.Group("/api/ads").Use(middleware.CheckBusinessProcess(processService, "reception", "schedule"))
	{
		adGroup.GET("/enabled", adHandler.GetEnabledAds)
	}

	audioGroup := r.Group("/api/audio").Use(middleware.CheckBusinessProcess(processService, "reception"))
	{
		audioGroup.GET("/announce", audioHandler.GenerateAnnouncement)
	}

	scheduleGroup := r.Group("/api/schedules").Use(middleware.CheckBusinessProcess(processService, "schedule"))
	{
		scheduleGroup.GET("/today/updates", scheduleHandler.GetTodayScheduleUpdates)
	}

	return r
}

type NotificationPayload struct {
	Action string                `json:"action"`
	Data   models.TicketResponse `json:"data"`
}

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

				if !strings.Contains(payloadStr, "ticket_number") {
					return true
				}

				log.WithField("payload", payloadStr).Info("SSE Handler: Sending message to client")
				var payload NotificationPayload
				if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
					log.WithError(err).Warn("Failed to unmarshal notification payload, skipping.")
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
