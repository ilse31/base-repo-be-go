package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"

	_ "github.com/ilse31/base-repo-be-go/docs"
	"github.com/ilse31/base-repo-be-go/internal/container"
	"github.com/ilse31/base-repo-be-go/internal/infrastructure/database"
	"github.com/ilse31/base-repo-be-go/internal/infrastructure/health"
	maileroo "github.com/ilse31/base-repo-be-go/internal/infrastructure/maileroo"
	"github.com/ilse31/base-repo-be-go/internal/infrastructure/redis"
	authinfra "github.com/ilse31/base-repo-be-go/internal/modules/auth/infrastructure"
	userdb "github.com/ilse31/base-repo-be-go/internal/modules/user/infrastructure"
	mw "github.com/ilse31/base-repo-be-go/internal/presentation/middleware"
	"github.com/ilse31/base-repo-be-go/internal/shared/response"
	"github.com/ilse31/base-repo-be-go/internal/shared/validation"
	"github.com/ilse31/base-repo-be-go/pkg/config"
	"github.com/ilse31/base-repo-be-go/pkg/jwt"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
	"github.com/ilse31/base-repo-be-go/pkg/mailer"
	"github.com/ilse31/base-repo-be-go/pkg/observability"
	"github.com/ilse31/base-repo-be-go/pkg/scheduler"
)

// @title Go Clean Architecture API
// @version 1.0
// @description Reusable Go Echo REST API Template following Clean Architecture.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Initialize logger
	if err := logger.Init(cfg.Server.Env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting application")

	// Initialize observability: tracing + error tracking. Must run before DB
	// and Redis clients are built so they can register instrumentation hooks.
	traceShutdown, err := observability.SetupTracing(cfg.Observability)
	if err != nil {
		logger.Fatal("Failed to initialize tracing")
	}
	if err := observability.SetupSentry(cfg.Observability); err != nil {
		logger.Fatal("Failed to initialize Sentry")
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		logger.Fatal("Failed to ping database")
	}
	logger.Info("Database connected successfully")

	// Initialize Redis
	redisClient, err := redis.NewClient(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()
	logger.Info("Redis connected successfully")

	// Initialize JWT manager
	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenExpiration, cfg.JWT.RefreshTokenExpiration)

	// Initialize repositories
	userRepo := userdb.NewUserRepository(db.DB)
	authRepo := authinfra.NewAuthRepository(redisClient)

	// Initialize shared validator
	v := validation.New()

	// Initialize mailer (Maileroo when enabled, no-op otherwise)
	var mailSender mailer.Mailer = mailer.NewNoop()
	if cfg.Mail.Enabled {
		sender, err := maileroo.New(cfg.Mail.APIKey, cfg.Mail.FromAddress, cfg.Mail.FromName)
		if err != nil {
			logger.Fatal("Failed to initialize mailer")
		}
		mailSender = sender
		logger.Info("Mailer initialized (Maileroo)")
	} else {
		logger.Info("Mailer disabled, using no-op mailer")
	}

	// Initialize Echo
	e := echo.New()
	// OpenTelemetry: create a span per HTTP request (outermost, wraps Sentry).
	e.Use(otelecho.Middleware(cfg.Observability.ServiceName))
	// Sentry: recover from panics and capture them; also scopes per-request hub.
	if cfg.Observability.SentryEnabled {
		e.Use(sentryecho.New(sentryecho.Options{}))
	}
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:     true,
		LogStatus:  true,
		LogMethod:  true,
		LogLatency: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("HTTP Request",
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.Duration("latency", v.Latency),
			)
			return nil
		},
	}))
	e.Use(middleware.Recover())

	// Security Middlewares: CORS, Secure Headers, Rate Limiter, and CSRF Protection
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.Server.AllowedOrigins,
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXCSRFToken},
		AllowCredentials: true,
	}))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:;",
	}))

	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieSecure:   cfg.Server.Env == "production",
		CookieHTTPOnly: false, // Accessible by frontend JS to set X-CSRF-Token header
		CookieSameSite: http.SameSiteStrictMode,
		TokenLookup:    "header:" + echo.HeaderXCSRFToken,
		Skipper: func(c echo.Context) bool {
			// Skip CSRF for safe HTTP methods, health check, or swagger docs
			if c.Request().Method == http.MethodGet ||
				c.Request().Method == http.MethodHead ||
				c.Request().Method == http.MethodOptions ||
				c.Path() == "/health" ||
				strings.HasPrefix(c.Path(), "/swagger") {
				return true
			}
			// Skip CSRF if request uses Bearer token authorization
			authHeader := c.Request().Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				return true
			}
			return false
		},
	}))

	e.HTTPErrorHandler = mw.HTTPErrorHandler

	e.GET("/health", func(c echo.Context) error {
		return response.OKWithMessage(c, "ok", map[string]string{"status": "ok"})
	})

	// Swagger UI route
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Compose modules and mount routes
	appContainer := container.New(container.Deps{
		UserRepo:         userRepo,
		AuthRepo:         authRepo,
		JWTManager:       jwtManager,
		Mailer:           mailSender,
		FrontendResetURL: cfg.Mail.FrontendResetURL,
		Validator:        v,
		IsProduction:     cfg.Server.Env == "production",
	})
	appContainer.RegisterRoutes(e)

	// Initialize health checker and scheduler
	healthChecker := health.NewChecker()
	healthChecker.Register("postgres", db)
	healthChecker.Register("redis", redisClient)

	sched := scheduler.New()
	if err := sched.AddJob("health-check", cfg.Scheduler.HealthInterval, healthChecker.Check); err != nil {
		logger.Fatal("Failed to register health check job")
	}
	sched.Start()

	// Start server
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		logger.Info("Server starting")
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Stop the scheduler and wait for in-flight jobs to complete.
	sched.Stop()

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown")
	}

	// Flush pending telemetry: traces to the collector, events to Sentry.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := traceShutdown(shutdownCtx); err != nil {
		logger.Error("Failed to flush traces")
	}
	observability.FlushSentry(2 * time.Second)

	logger.Info("Server exited")
}
