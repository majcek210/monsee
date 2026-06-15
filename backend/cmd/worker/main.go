package main

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/majcek210/monsee/internal/config"
	"github.com/majcek210/monsee/internal/notifications"
	"github.com/majcek210/monsee/internal/repository/postgres"
	"github.com/majcek210/monsee/internal/service"
	"github.com/majcek210/monsee/internal/telemetry"
	"github.com/majcek210/monsee/internal/worker"
	whooks "github.com/majcek210/monsee/internal/webhooks"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger := telemetry.Must(telemetry.NewLogger(cfg.IsProd()))
	defer logger.Sync() //nolint:errcheck

	shutdownTracing, err := telemetry.InitTracing(context.Background(), cfg.IsProd())
	if err != nil {
		logger.Fatal("tracing init", zap.Error(err))
	}
	defer shutdownTracing(context.Background()) //nolint:errcheck

	metrics := telemetry.NewMetrics()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("db pool", zap.Error(err))
	}
	defer pool.Close()

	encKey := []byte(cfg.EncryptionKey)

	// ── Repositories ──────────────────────────────────────────────────────────
	monitorRepo     := postgres.NewMonitorRepo(pool)
	checkResultRepo := postgres.NewCheckResultRepo(pool)
	incidentRepo    := postgres.NewIncidentRepo(pool)
	notifRepo       := postgres.NewNotificationRepo(pool)
	webhookRepo     := postgres.NewWebhookRepo(pool)

	// ── Dispatchers ───────────────────────────────────────────────────────────
	smtpCfg := notifications.SMTPConfig{
		Host: cfg.SMTPHost,
		Port: cfg.SMTPPort,
		User: cfg.SMTPUser,
		Pass: cfg.SMTPPass,
		From: cfg.SMTPFrom,
	}
	notifDispatcher   := notifications.NewDispatcher(notifRepo, encKey, smtpCfg, cfg.FrontendURL)
	webhookDispatcher := whooks.NewDispatcher(webhookRepo, encKey)

	// ── Checker service ───────────────────────────────────────────────────────
	checkerSvc := service.NewCheckerService(
		monitorRepo,
		checkResultRepo,
		incidentRepo,
		notifDispatcher,
		webhookDispatcher,
	).WithMetrics(metrics).WithLogger(logger)

	// ── Asynq ─────────────────────────────────────────────────────────────────
	redisAddr := strings.TrimPrefix(cfg.RedisURL, "redis://")
	redisOpt := asynq.RedisClientOpt{Addr: redisAddr}

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		Queues:      map[string]int{"monitors": 8, "cleanup": 1},
	})

	mux := asynq.NewServeMux()
	mux.Handle(worker.TaskRunMonitorCheck, worker.NewCheckMonitorHandler(checkerSvc))

	client := asynq.NewClient(redisOpt)
	defer client.Close()

	scheduler := worker.NewScheduler(monitorRepo, client)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go scheduler.Run(ctx)

	logger.Info("worker starting")
	if err := srv.Run(mux); err != nil {
		logger.Fatal("worker", zap.Error(err))
	}
}
