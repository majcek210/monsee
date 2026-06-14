package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/majcek210/monsee/internal/config"
	"github.com/majcek210/monsee/internal/handler"
	"github.com/majcek210/monsee/internal/repository/postgres"
	redisrepo "github.com/majcek210/monsee/internal/repository/redis"
	"github.com/majcek210/monsee/internal/service"
	"github.com/majcek210/monsee/internal/telemetry"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── Telemetry ─────────────────────────────────────────────────────────────
	logger := telemetry.Must(telemetry.NewLogger(cfg.IsProd()))
	defer logger.Sync() //nolint:errcheck

	shutdownTracing, err := telemetry.InitTracing(context.Background(), cfg.IsProd())
	if err != nil {
		logger.Fatal("tracing init", zap.Error(err))
	}
	defer shutdownTracing(context.Background()) //nolint:errcheck

	metrics := telemetry.NewMetrics()

	// ── Prometheus metrics server on :2112 ────────────────────────────────────
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		logger.Info("metrics server starting on :2112")
		if err := http.ListenAndServe(":2112", mux); err != nil {
			logger.Error("metrics server", zap.Error(err))
		}
	}()

	// ── Migrations ────────────────────────────────────────────────────────────
	migrateURL := "pgx5://" + strings.TrimPrefix(cfg.DatabaseURL, "postgres://")
	m, err := migrate.New("file://db/migrations", migrateURL)
	if err != nil {
		logger.Fatal("migrate init", zap.Error(err))
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatal("migrate up", zap.Error(err))
	}
	logger.Info("migrations applied")

	// ── Database pool ─────────────────────────────────────────────────────────
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("db pool", zap.Error(err))
	}
	defer pool.Close()

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisAddr := strings.TrimPrefix(cfg.RedisURL, "redis://")
	rdb := redisrepo.NewRedisClient(redisAddr)
	defer rdb.Close()
	limiter := redisrepo.NewRateLimiter(rdb, 100, time.Minute)

	encKey := []byte(cfg.EncryptionKey)

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo     := postgres.NewUserRepo(pool)
	serviceRepo  := postgres.NewServiceRepo(pool)
	monitorRepo  := postgres.NewMonitorRepo(pool)
	incidentRepo := postgres.NewIncidentRepo(pool)
	apikeyRepo   := postgres.NewAPIKeyRepo(pool)
	notifRepo    := postgres.NewNotificationRepo(pool)
	webhookRepo  := postgres.NewWebhookRepo(pool)
	auditRepo    := postgres.NewAuditRepo(pool)

	// ── Services ──────────────────────────────────────────────────────────────
	userSvc     := service.NewUserService(userRepo)
	serviceSvc  := service.NewMonitoringService(serviceRepo)
	monitorSvc  := service.NewMonitorService(monitorRepo, serviceRepo)
	incidentSvc := service.NewIncidentService(incidentRepo, serviceRepo)
	apikeySvc   := service.NewAPIKeyService(apikeyRepo)
	notifSvc    := service.NewNotificationService(notifRepo, encKey)
	webhookSvc  := service.NewWebhookService(webhookRepo, encKey)

	// ── HTTP app ──────────────────────────────────────────────────────────────
	app := handler.New(handler.Deps{
		Cfg:           cfg,
		Users:         userSvc,
		Services:      serviceSvc,
		Monitors:      monitorSvc,
		Incidents:     incidentSvc,
		APIKeys:       apikeySvc,
		Notifications: notifSvc,
		Webhooks:      webhookSvc,
		Limiter:       limiter,
		Audit:         auditRepo,
		Metrics:       metrics,
		Logger:        logger,
	})

	logger.Info("server starting", zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		logger.Fatal("server", zap.Error(err))
	}
}
