package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear any relevant env vars
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("ENV")
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSLMODE")
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("REDIS_DB")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_ACCESS_EXPIRATION")
	os.Unsetenv("JWT_REFRESH_EXPIRATION")
	os.Unsetenv("SCHEDULER_HEALTH_INTERVAL")
	os.Unsetenv("MAILEROO_API_KEY")
	os.Unsetenv("MAIL_FROM_ADDRESS")
	os.Unsetenv("MAIL_FROM_NAME")
	os.Unsetenv("MAIL_FRONTEND_RESET_URL")
	os.Unsetenv("MAIL_ENABLED")
	os.Unsetenv("OTEL_SERVICE_NAME")
	os.Unsetenv("OTEL_TRACING_ENABLED")
	os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Unsetenv("OTEL_TRACES_SAMPLE_RATE")
	os.Unsetenv("SENTRY_ENABLED")
	os.Unsetenv("SENTRY_DSN")
	os.Unsetenv("SENTRY_ENVIRONMENT")
	os.Unsetenv("SENTRY_TRACES_SAMPLE_RATE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("expected 8080, got %s", cfg.Server.Port)
	}
	if cfg.Server.Env != "development" {
		t.Errorf("expected development, got %s", cfg.Server.Env)
	}
	if !reflect.DeepEqual(cfg.Server.AllowedOrigins, []string{"*"}) {
		t.Errorf("expected [*], got %v", cfg.Server.AllowedOrigins)
	}
	if cfg.Database.Host != "localhost" || cfg.Database.Port != "5432" || cfg.Database.DBName != "mydb" {
		t.Errorf("unexpected database config: %+v", cfg.Database)
	}
	if cfg.Redis.Host != "localhost" || cfg.Redis.Port != "6379" || cfg.Redis.DB != 0 {
		t.Errorf("unexpected redis config: %+v", cfg.Redis)
	}
	if cfg.JWT.AccessTokenExpiration != 1 || cfg.JWT.RefreshTokenExpiration != 168 {
		t.Errorf("unexpected jwt expiration config: %+v", cfg.JWT)
	}
	if cfg.Mail.Enabled != false {
		t.Errorf("expected mail enabled to be false")
	}
}

func TestLoadCustomEnv(t *testing.T) {
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("ENV", "production")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com, https://test.com")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("JWT_SECRET", "super-secret-key-that-is-at-least-32-chars-long!")
	t.Setenv("JWT_ACCESS_EXPIRATION", "2")
	t.Setenv("MAIL_ENABLED", "true")
	t.Setenv("OTEL_TRACES_SAMPLE_RATE", "0.5")
	t.Setenv("OTEL_TRACING_ENABLED", "invalid_bool_fallback")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Server.Port != "9090" {
		t.Errorf("expected 9090, got %s", cfg.Server.Port)
	}
	if cfg.Server.Env != "production" {
		t.Errorf("expected production, got %s", cfg.Server.Env)
	}
	expectedOrigins := []string{"https://example.com", "https://test.com"}
	if !reflect.DeepEqual(cfg.Server.AllowedOrigins, expectedOrigins) {
		t.Errorf("expected %v, got %v", expectedOrigins, cfg.Server.AllowedOrigins)
	}
	if cfg.Redis.DB != 2 {
		t.Errorf("expected redis db 2, got %d", cfg.Redis.DB)
	}
	if cfg.JWT.AccessTokenExpiration != 2 {
		t.Errorf("expected access expiration 2, got %d", cfg.JWT.AccessTokenExpiration)
	}
	if cfg.Mail.Enabled != true {
		t.Errorf("expected mail enabled to be true")
	}
	if cfg.Observability.TracesSampleRate != 0.5 {
		t.Errorf("expected traces sample rate 0.5, got %f", cfg.Observability.TracesSampleRate)
	}
	// invalid bool falls back to default false
	if cfg.Observability.TracingEnabled != false {
		t.Errorf("expected tracing enabled to fallback to false")
	}
}

func TestValidate(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Env: "development"},
		JWT:    JWTConfig{Secret: "your-secret-key-change-in-production"},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error in development, got %v", err)
	}

	cfg.Server.Env = "production"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for default secret in production, got nil")
	}

	cfg.JWT.Secret = "short"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for short secret in production, got nil")
	}

	cfg.JWT.Secret = "super-secret-key-that-is-at-least-32-chars-long!"
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error for valid secret in production, got %v", err)
	}
}

func TestGetEnvAsIntInvalid(t *testing.T) {
	t.Setenv("TEST_INT", "not_a_number")
	val := getEnvAsInt("TEST_INT", 42)
	if val != 42 {
		t.Errorf("expected default 42, got %d", val)
	}
}

func TestGetEnvAsFloatInvalid(t *testing.T) {
	t.Setenv("TEST_FLOAT", "not_a_float")
	val := getEnvAsFloat("TEST_FLOAT", 1.23)
	if val != 1.23 {
		t.Errorf("expected default 1.23, got %f", val)
	}
}

func TestGetEnvAsSliceEmpty(t *testing.T) {
	t.Setenv("TEST_SLICE", "   ,   ")
	val := getEnvAsSlice("TEST_SLICE", []string{"default"})
	if !reflect.DeepEqual(val, []string{"default"}) {
		t.Errorf("expected default, got %v", val)
	}
}
