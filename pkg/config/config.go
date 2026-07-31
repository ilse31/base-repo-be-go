package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	Redis          RedisConfig
	JWT            JWTConfig
	Scheduler      SchedulerConfig
	Mail           MailConfig
	Observability  ObservabilityConfig
	RateLimit      RateLimitConfig
}

type ServerConfig struct {
	Port           string
	Env            string
	AllowedOrigins []string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret                 string
	AccessTokenExpiration  int // hours
	RefreshTokenExpiration int // hours
}

// SchedulerConfig configures background periodic jobs. HealthInterval is a
// cron schedule expression (e.g. "@every 10m" or "*/10 * * * *").
type SchedulerConfig struct {
	HealthInterval string
}

// MailConfig configures outbound transactional email via Maileroo.
// FrontendResetURL is the client-side page users land on to reset their
// password; the reset token is appended as the "token" query parameter.
type MailConfig struct {
	APIKey           string
	FromAddress      string
	FromName         string
	FrontendResetURL string
	Enabled          bool
}

// RateLimitConfig configures the Redis-backed HTTP rate limiter: up to
// Requests calls per identifier (client IP) within Window.
type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

// ObservabilityConfig configures distributed tracing (OpenTelemetry) and
// error tracking (Sentry). Both are opt-in; when disabled the app runs with
// no-op providers and no external calls.
type ObservabilityConfig struct {
	ServiceName       string
	TracingEnabled    bool
	OTLPEndpoint      string // OTLP/HTTP endpoint, host:port (e.g. "localhost:4318")
	TracesSampleRate  float64
	SentryEnabled     bool
	SentryDSN         string
	SentryEnv         string
	SentrySampleRate  float64
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		// .env file is optional, continue without it
	}

	return &Config{
		Server: ServerConfig{
			Port:           getEnv("SERVER_PORT", "8080"),
			Env:            getEnv("ENV", "development"),
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "mydb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:                 getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenExpiration:  getEnvAsInt("JWT_ACCESS_EXPIRATION", 1),
			RefreshTokenExpiration: getEnvAsInt("JWT_REFRESH_EXPIRATION", 168), // 7 days
		},
		Scheduler: SchedulerConfig{
			HealthInterval: getEnv("SCHEDULER_HEALTH_INTERVAL", "@every 10m"),
		},
		Mail: MailConfig{
			APIKey:           getEnv("MAILEROO_API_KEY", ""),
			FromAddress:      getEnv("MAIL_FROM_ADDRESS", "noreply@example.com"),
			FromName:         getEnv("MAIL_FROM_NAME", "No-Reply"),
			FrontendResetURL: getEnv("MAIL_FRONTEND_RESET_URL", "http://localhost:3000/reset-password"),
			Enabled:          getEnvAsBool("MAIL_ENABLED", false),
		},
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS", 20),
			Window:   time.Duration(getEnvAsInt("RATE_LIMIT_WINDOW_SECONDS", 1)) * time.Second,
		},
		Observability: ObservabilityConfig{
			ServiceName:      getEnv("OTEL_SERVICE_NAME", "base-repo-be-go"),
			TracingEnabled:   getEnvAsBool("OTEL_TRACING_ENABLED", false),
			OTLPEndpoint:     getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
			TracesSampleRate: getEnvAsFloat("OTEL_TRACES_SAMPLE_RATE", 1.0),
			SentryEnabled:    getEnvAsBool("SENTRY_ENABLED", false),
			SentryDSN:        getEnv("SENTRY_DSN", ""),
			SentryEnv:        getEnv("SENTRY_ENVIRONMENT", getEnv("ENV", "development")),
			SentrySampleRate: getEnvAsFloat("SENTRY_TRACES_SAMPLE_RATE", 0.1),
		},
	}, nil
}

func (c *Config) Validate() error {
	if c.Server.Env == "production" {
		if c.JWT.Secret == "your-secret-key-change-in-production" || len(c.JWT.Secret) < 32 {
			return errors.New("JWT_SECRET is insecure for production (must be at least 32 characters and non-default)")
		}
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		res := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				res = append(res, trimmed)
			}
		}
		if len(res) > 0 {
			return res
		}
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}
