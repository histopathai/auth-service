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

// CORSConfig holds settings for Cross-Origin Resource Sharing
type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	TrustedProxies []string
}

type TLSConfig struct {
	CertPath string
	KeyPath  string
}

// Config is the top-level application configuration
type Config struct {
	ProjectID      string
	Region         string
	ProjectNumber  string
	MainServiceURL string
	Server         ServerConfig
	Cookie         CookieConfig
	CORS           CORSConfig
	Security       SecurityConfig
	TLS            TLSConfig
	Logging        LoggingConfig
}

// LoadConfig reads configuration from environment variables
func LoadConfig() *Config {
	env := getEnv("ENVIRONMENT", "dev")

	// Start with development defaults
	cfg := &Config{
		ProjectID:      getEnv("PROJECT_ID", ""),
		Region:         getEnv("REGION", ""),
		ProjectNumber:  getEnv("PROJECT_NUMBER", ""),
		MainServiceURL: getEnv("MAIN_SERVICE_URL", "http://localhost:8081"),
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

	// Environment-specific overrides
	switch env {
	case "prod":
		cfg.MainServiceURL = getEnv("MAIN_SERVICE_URL", "") // Must be set in prod
		cfg.Server.GINMode = "release"
		cfg.Logging.Level = getEnv("LOG_LEVEL", "info")
		cfg.Logging.Format = getEnv("LOG_FORMAT", "json")

		cfg.Cookie = CookieConfig{
			Name:     "session_id",
			Domain:   getEnv("COOKIE_DOMAIN", ""), // e.g., ".yourdomain.com"
			Secure:   true,
			SameSite: "None",
			HTTPOnly: true,
			MaxAge:   1800, // 30 minutes
		}
		cfg.CORS = CORSConfig{
			AllowedOrigins: []string{
				getEnv("FRONTEND_URL", "https://histopathai.com"),
				"https://localhost:5173",
			},
			AllowCredentials: true,
		}
		cfg.Security = SecurityConfig{
			TrustedProxies: []string{}, // Cloud Run internal IPs
		}

	default: // dev
		cfg.MainServiceURL = getEnv("MAIN_SERVICE_URL", "http://localhost:8081")
		// GINMode, Logging level/format already set to dev defaults

		cfg.Cookie = CookieConfig{
			Name:     "session_id",
			Domain:   "", // Current domain
			Secure:   false,
			SameSite: "Lax",
			HTTPOnly: true,
			MaxAge:   1800,
		}
		cfg.CORS = CORSConfig{
			AllowedOrigins: []string{
				"http://localhost:5173",
				"http://localhost:3000",
				"https://localhost:5173",
				"http://127.0.0.1:5173",
			},
			AllowCredentials: true,
		}
		cfg.Security = SecurityConfig{
			TrustedProxies: nil, // Trust all in dev
		}
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
