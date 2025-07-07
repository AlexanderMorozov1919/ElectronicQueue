package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/database"

	_ "ElectronicQueue/docs"

	ginSwaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/lib/pq"

	"gorm.io/gorm"

)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Ошибка загрузки конфига: %v\n", err)
		return
	}

	// Инициализация логгера
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		multi := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multi)
	} else {
		log.Info("Ошибка открытия файла логов, будет использоваться только вывод в консоль")
	}

	// Подключение к базе данных
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.WithError(err).Fatal("Database connection failed")
	}
	log.WithField("dbname", cfg.DBName).Info("Database connected successfully")

	// Инициализация listener для LISTEN/NOTIFY
	listener, err := initListener(cfg, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize database listener")
	}
	defer listener.Close()

	// Настройка роутера
	r := setupRouter(listener, db)

	// Обработка сигналов завершения
	handleGracefulShutdown(db, listener, log)

	fmt.Printf("Сервер запущен на порту: %s\n", cfg.BackendPort)
	if err := r.Run(":" + cfg.BackendPort); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}

	fmt.Printf("Successful connect to database: \"%s\"\n", db.Name())
}

func setupRouter(listener *Listener, db *pgxpool.Pool) *gin.Engine {
	r := gin.Default()

	// Пример маршрута
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})


// setupRouter настраивает маршруты и middleware
func setupRouter(listener *pq.Listener, db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(logger.GinLogger())

	// CORS middleware
	r.Use(corsMiddleware())

	// GIN middleware для логирования всех запросов
	r.Use(requestLogger())

	// SSE endpoint
	r.GET("/tickets", sseHandler(listener))

	// Инициализация репозитория, сервиса и хендлера для талонов
	ticketRepo := repository.NewTicketRepository(db)
	ticketService := services.NewTicketService(ticketRepo)
	ticketHandler := handlers.NewTicketHandler(ticketService)

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

func handleGracefulShutdown(db *pgxpool.Pool, listener *Listener, log *logrus.Logger) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info("Shutdown signal received")

		// Остановка слушателя
		listener.Close()

		// Отключение от базы данных
		db.Close()

		log.Info("Services stopped")
	}()
}

func initListener(cfg *config.Config, log *logrus.Logger) (*Listener, error) {
	// Инициализация listener для LISTEN/NOTIFY
	listener := &Listener{
		// Настройки
	}

	go func() {
		for {
			// Логика обработки сообщений
			time.Sleep(10 * time.Second) // Пример задержки
		}
	}()

	return listener, nil
}

type Listener struct {
	// Настройки listener
}

func (l *Listener) Close() {
	// Логика остановки listener
}
