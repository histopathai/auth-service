package config

import (
	"fmt"
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
		MainServiceURL: os.Getenv("MAIN_SERVICE_URL"),
		Server: ServerConfig{
			Port:         getEnvOrDefault("PORT", "8080"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 15),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 15),
			IdleTimeout:  getEnvAsInt("IDLE_TIMEOUT", 60),
			GINMode:      getEnvOrDefault("GIN_MODE", "debug"),
		},
		// EKLENDİ: Logging yapılandırması
		Logging: LoggingConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),  // Örn: debug, info, warn, error
			Format: getEnvOrDefault("LOG_FORMAT", "json"), // "json" veya "text"
		},
	}

	if env != "LOCAL" {
		project_number := os.Getenv("PROJECT_NUMBER")
		service_name := os.Getenv("MAIN_SERVICE_NAME")
		region := os.Getenv("REGION")
		if project_number == "" || service_name == "" || region == "" {
			return nil, fmt.Errorf("PROJECT_NUMBER, MAIN_SERVICE_NAME, and REGION must be set in non-local environments")
		}

		config.MainServiceURL = fmt.Sprintf("https://%s-%s.%s.run.app", service_name, project_number, region)
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
