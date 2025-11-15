package config

import (
	"os"
	"strconv"
)

// LoggingConfig holds settings for the logger
type LoggingConfig struct {
	Level  string
	Format string
}

// ServerConfig holds settings for the HTTP server
type ServerConfig struct {
	Port         string
	Environment  string
	BaseURL      string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
	GINMode      string
}

// CookieConfig holds settings for session cookies
type CookieConfig struct {
	Name     string
	Domain   string
	Secure   bool
	SameSite string
	HTTPOnly bool
	MaxAge   int // in seconds
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	TrustedProxies []string
}

type TLSConfig struct {
	CertPath string
	KeyPath  string
}

type Config struct {
	ProjectID      string
	Region         string
	ProjectNumber  string
	MainServiceURL string
	AllowedOrigin  string
	Server         ServerConfig
	Cookie         CookieConfig
	Security       SecurityConfig
	TLS            TLSConfig
	Logging        LoggingConfig
}

func LoadConfig() *Config {
	env := getEnv("ENVIRONMENT", "dev")

	cfg := &Config{
		ProjectID:      getEnv("PROJECT_ID", ""),
		Region:         getEnv("REGION", ""),
		ProjectNumber:  getEnv("PROJECT_NUMBER", ""),
		MainServiceURL: getEnv("MAIN_SERVICE_URL", "http://localhost:8081"),
		AllowedOrigin:  getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Environment:  env,
			BaseURL:      getEnv("BASE_URL", "http://localhost:8080"),
			ReadTimeout:  getEnvInt("READ_TIMEOUT", 15),
			WriteTimeout: getEnvInt("WRITE_TIMEOUT", 15),
			IdleTimeout:  getEnvInt("IDLE_TIMEOUT", 60),
			GINMode:      "debug",
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
	}

	cfg.Cookie = CookieConfig{
		Name:     "session_id",
		Domain:   getEnv("COOKIE_DOMAIN", ""),
		Secure:   true,
		SameSite: "None",
		HTTPOnly: true,
		MaxAge:   1800,
	}

	// Environment-specific overrides
	if env == "prod" {
		cfg.Server.GINMode = "release"
		cfg.Logging.Level = getEnv("LOG_LEVEL", "info")
		cfg.Logging.Format = getEnv("LOG_FORMAT", "json")
	} else {

		cfg.TLS = TLSConfig{
			CertPath: getEnv("CERT_PATH", ""),
			KeyPath:  getEnv("KEY_PATH", ""),
		}
	}

	return cfg
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an environment variable as an integer or returns a default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
