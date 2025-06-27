package config

import (
	"os"
	"strconv"
)

// Config holds all the application configurations
type Config struct {
	Server    ServerConfig    `json:"server"`
	Firebase  FirebaseConfig  `json:"firebase"`
	Firestore FirestoreConfig `json:"firestore"`
	SMTP      SMTPConfig      `json:"smtp"`
}

// ServerConfig holds server-related configurations
type ServerConfig struct {
	Port         int `json:"port"`
	ReadTimeout  int `json:"read_timeout"`
	WriteTimeout int `json:"write_timeout"`
	IdleTimeout  int `json:"idle_timeout"`
}

// FirebaseConfig holds Firebase-related configurations
type FirebaseConfig struct {
	ProjectID         string `json:"project_id"`
	ServiceAccountKey string `json:"service_account_key"`
	AuthEmulatorHost  string `mapstructure:"FIREBASE_AUTH_EMULATOR_HOST"`
}

// FirestoreConfig holds Firestore-related configurations
type FirestoreConfig struct {
	UsersCollection string `json:"users_collection"`
	ProjectID       string `json:"project_id"`    // Optional, can be used for Firestore client initialization
	EmulatorHost    string `json:"emulator_host"` // Optional, can be used to connect to Firestore emulator
}

// SMPTConfig holds SMTP server configurations
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Sender   string `json:"sender"`
}

// LoadConfig loads the configuration from environment variables or defaults
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 15),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 15),
			IdleTimeout:  getEnvAsInt("IDLE_TIMEOUT", 60),
		},
		Firebase: FirebaseConfig{
			ProjectID: getEnv("FIREBASE_PROJECT_ID", "your-project-id"),
		},
		Firestore: FirestoreConfig{
			UsersCollection: getEnv("FIRESTORE_USERS_COLLECTION", "users"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.example.com"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", "root"),
			Password: getEnv("SMTP_PASSWORD", "password"),
			Sender:   getEnv("SMTP_SENDER_EMAIL", "smtp.example.com"),
		},
	}
	return config, nil
}

// getEnv retrieves an environment variable or returns a default value if not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value if not set
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
