package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment   string
	ServerAddress string
	DBConfig      *DBConfig
}

type DBConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

func Load() *Config {
	dbConfig := NewDBConfig()

	return &Config{
		Environment:   getEnv("ENVIRONMENT", "development"),
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		DBConfig:      dbConfig,
	}
}

func NewDBConfig() *DBConfig {
	return &DBConfig{
		Host:         getEnv("DB_HOST", "localhost"),
		Port:         getEnv("DB_PORT", "5432"),
		User:         getEnv("DB_USER", "postgres"),
		Password:     getEnv("DB_PASSWORD", "password"),
		Database:     getEnv("DB_NAME", "pr_review"),
		SSLMode:      getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		MaxLifetime:  time.Duration(getEnvAsInt("DB_MAX_LIFETIME_MINUTES", 120)) * time.Minute,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Invalid value for %s, using default: %d", key, defaultValue)
	}
	return defaultValue
}
