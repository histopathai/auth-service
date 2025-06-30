package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server    ServerConfig
	Firebase  FirebaseConfig
	Firestore FirestoreConfig
	SMTP      SMTPConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

type FirebaseConfig struct {
	ProjectID string // Optional, mostly unused now
}

type FirestoreConfig struct {
	UsersCollection string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

func LoadConfig() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 15),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 15),
			IdleTimeout:  getEnvAsInt("IDLE_TIMEOUT", 60),
		},
		Firebase: FirebaseConfig{
			ProjectID: getEnv("FIREBASE_PROJECT_ID", ""), // optional
		},
		Firestore: FirestoreConfig{
			UsersCollection: getEnv("FIRESTORE_USERS_COLLECTION", "users"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			Sender:   getEnv("SMTP_SENDER_EMAIL", ""),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
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
