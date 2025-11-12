package config

import (
	"os"
	"strconv"
)

type LoggingConfig struct {
	Level  string
	Format string
}

type Config struct {
	ProjectID      string
	Region         string
	ProjectNumber  string
	MainServiceURL string
	Server         ServerConfig
	Logging        LoggingConfig // EKLENDİ
}

type ServerConfig struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
	GINMode      string
}

func LoadConfig() (*Config, error) {
	env := os.Getenv("ENV")

	config := &Config{
		ProjectID:      os.Getenv("PROJECT_ID"),
		Region:         os.Getenv("REGION"),
		MainServiceURL: os.Getenv("MAIN_SERVICE_URL"), // ← Terraform'dan gelir
		Server: ServerConfig{
			Port:         getEnvOrDefault("PORT", "8080"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 15),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 15),
			IdleTimeout:  getEnvAsInt("IDLE_TIMEOUT", 60),
			GINMode:      getEnvOrDefault("GIN_MODE", "debug"),
		},
		Logging: LoggingConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
		},
	}

	// LOCAL development override
	if env == "LOCAL" {
		if config.MainServiceURL == "" {
			config.MainServiceURL = "http://localhost:8081"
		}
	}

	if env != "LOCAL" {
		config.Server.GINMode = "release"
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
