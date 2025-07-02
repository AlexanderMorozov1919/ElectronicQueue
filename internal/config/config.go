package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config содержит переменные среды
type Config struct {
	DBUser        string
	DBPassword    string
	DBHost        string
	DBPort        string
	DBName        string
	DBSSLMode     string
	ServerPort    string
	JWTSecret     string
	JWTExpiration string
}

// LoadConfig загружает переменные среды из .env и возвращает структуру Config
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, relying on environment variables. %v", err)
	}

	cfg := &Config{
		DBUser:        getEnv("DB_USER"),
		DBPassword:    getEnv("DB_PASSWORD"),
		DBHost:        getEnv("DB_HOST"),
		DBPort:        getEnv("DB_PORT"),
		DBName:        getEnv("DB_NAME"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		JWTSecret:     getEnv("JWT_SECRET"),
		JWTExpiration: getEnv("JWT_EXPIRATION", "24h"),
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
