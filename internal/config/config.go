package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config содержит переменные среды
type Config struct {
	DBUser        string
	DBPassword    string
	DBHost        string
	DBPort        string
	DBName        string
	DBSSLMode     string
	BackendPort   string
	FrontendPort  string
	JWTSecret     string
	JWTExpiration string
	LogFile       string `mapstructure:"LOG_FILE"`
}

// LoadConfig загружает переменные среды из .env и возвращает структуру Config
func LoadConfig() (*Config, error) {
	// Инициализация логгера
	log := logrus.New()

	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found, using environment variables")
	}

	cfg := &Config{
		DBUser:        getEnv("DB_USER"),
		DBPassword:    getEnv("DB_PASSWORD"),
		DBHost:        getEnv("DB_HOST"),
		DBPort:        getEnv("DB_PORT"),
		DBName:        getEnv("DB_NAME"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		BackendPort:   getEnv("BACKEND_PORT", "8080"),
		FrontendPort:  getEnv("FRONTEND_PORT", "3000"),
		JWTSecret:     getEnv("JWT_SECRET"),
		JWTExpiration: getEnv("JWT_EXPIRATION", "24h"),
		LogFile:       getEnv("LOG_FILE"),
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
	return " "
}
