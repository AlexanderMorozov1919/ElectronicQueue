package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config содержит переменные среды
type Config struct {
	DBUser                      string
	DBPassword                  string
	DBHost                      string
	DBPort                      string
	DBName                      string
	DBSSLMode                   string
	BackendPort                 string
	FrontendPort                string
	JWTSecret                   string
	JWTExpiration               string
	TicketMode                  string
	TicketHeight                string
	LogDir                      string
	TicketDir                   string
	InternalAPIKey              string
	ExternalAPIKey              string
	PrinterName                 string
	MaintenanceTime             string
	AudioBackgroundMusicEnabled bool
}

// LoadConfig загружает переменные среды из .env и возвращает структуру Config
func LoadConfig() (*Config, error) {
	log := logrus.New()

	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found, using environment variables")
	}

	cfg := &Config{
		DBUser:                      getEnv("DB_USER"),
		DBPassword:                  getEnv("DB_PASSWORD"),
		DBHost:                      getEnv("DB_HOST"),
		DBPort:                      getEnv("DB_PORT"),
		DBName:                      getEnv("DB_NAME"),
		DBSSLMode:                   getEnv("DB_SSLMODE", "disable"),
		BackendPort:                 getEnv("BACKEND_PORT", "8080"),
		FrontendPort:                getEnv("FRONTEND_PORT", "3000"),
		JWTSecret:                   getEnv("JWT_SECRET"),
		JWTExpiration:               getEnv("JWT_EXPIRATION", "24h"),
		TicketMode:                  getEnv("TICKET_MODE", "b/w"),
		TicketHeight:                getEnv("TICKET_HEIGHT", "800"),
		LogDir:                      getEnv("LOG_DIR", "logs"),
		TicketDir:                   getEnv("TICKET_DIR", "tickets"),
		InternalAPIKey:              getEnv("INTERNAL_API_KEY"),
		ExternalAPIKey:              getEnv("EXTERNAL_API_KEY"),
		PrinterName:                 getEnv("PRINTER"),
		MaintenanceTime:             getEnv("MAINTENANCE_TIME", "00:00"),
		AudioBackgroundMusicEnabled: getEnv("BACKGROUND_MUSIC", "true") == "true",
	}

	// Валидация обязательных полей
	if cfg.DBUser == "" {
		return nil, errors.New("DB_USER is not set in the environment")
	}
	if cfg.DBPassword == "" {
		return nil, errors.New("DB_PASSWORD is not set in the environment")
	}
	if cfg.DBHost == "" {
		return nil, errors.New("DB_HOST is not set in the environment")
	}
	if cfg.DBPort == "" {
		return nil, errors.New("DB_PORT is not set in the environment")
	}
	if cfg.DBName == "" {
		return nil, errors.New("DB_NAME is not set in the environment")
	}

	return cfg, nil
}

// getEnv получает переменную окружения с дефолтным значением
func getEnv(key string, defaultValue ...string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}
