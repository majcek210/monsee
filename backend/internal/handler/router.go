package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/majcek210/monsee/internal/config"
	"github.com/majcek210/monsee/internal/domain"
	v1 "github.com/majcek210/monsee/internal/handler/v1"
	"github.com/majcek210/monsee/internal/middleware"
	"github.com/majcek210/monsee/internal/repository/postgres"
	"github.com/majcek210/monsee/internal/service"
	"github.com/majcek210/monsee/internal/telemetry"
)

type Deps struct {
	Cfg           *config.Config
	Users         *service.UserService
	Services      *service.MonitoringService
	Monitors      *service.MonitorService
	Incidents     *service.IncidentService
	APIKeys       *service.APIKeyService
	Notifications *service.NotificationService
	Webhooks      *service.WebhookService
	Settings      *service.SettingsService
	Uptime        *service.UptimeService
	Maintenance   *service.MaintenanceService
	TwoFactor     *service.TwoFactorService
	AuditRepo     *postgres.AuditRepo
	Limiter       middleware.Limiter
	Audit         middleware.AuditLogger
	Metrics       *telemetry.Metrics
	Logger        *zap.Logger
}

// New creates and wires the Fiber app.
func New(d Deps) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})

	// ── Global middleware ────────────────────────────────────────────────────
	if d.Logger != nil {
		app.Use(middleware.RequestLogger(d.Logger))
	}
	if d.Metrics != nil {
		app.Use(middleware.PrometheusMiddleware(d.Metrics.HTTPDuration))
	}
	if d.Limiter != nil {
		var failOpenCounter prometheus.Counter
		if d.Metrics != nil {
			failOpenCounter = d.Metrics.RateLimiterFailOpen
		}
		app.Use(middleware.RateLimit(d.Limiter, d.Logger, failOpenCounter))
	}

	auth := middleware.RequireAuth(d.Cfg.JWTSecret)

	var auditMw fiber.Handler
	if d.Audit != nil {
		auditMw = middleware.Audit(d.Audit)
	}

	// ── Public ───────────────────────────────────────────────────────────────
	app.Get("/health", healthHandler)

	authGroup := app.Group("/auth")
	{
		uh := &UserHandler{users: d.Users, cfg: d.Cfg}
		authGroup.Post("/login", uh.Login)
		authGroup.Post("/logout", uh.Logout)
		authGroup.Get("/me", auth, uh.Me)

		if d.TwoFactor != nil {
			tfh := &TwoFactorHandler{tf: d.TwoFactor}
			authGroup.Post("/2fa/verify", tfh.Verify)
		}
	}

	// ── Admin (session-protected) ─────────────────────────────────────────────
	var admin fiber.Router
	if auditMw != nil {
		admin = app.Group("/admin", auth, auditMw)
	} else {
		admin = app.Group("/admin", auth)
	}
	{
		sh := &ServiceHandler{svc: d.Services}
		admin.Get("/services", sh.List)
		admin.Post("/services", middleware.RequireAdmin, sh.Create)
		admin.Get("/services/:id", sh.Get)
		admin.Patch("/services/:id", middleware.RequireAdmin, sh.Update)
		admin.Delete("/services/:id", middleware.RequireAdmin, sh.Archive)
		if d.Uptime != nil {
			admin.Get("/services/:id/uptime", func(c fiber.Ctx) error {
				result, err := d.Uptime.GetServiceUptime(c.Context(), c.Params("id"), nil)
				if err != nil {
					return err
				}
				return c.JSON(result)
			})
			admin.Get("/uptime", func(c fiber.Ctx) error {
				results, err := d.Uptime.GetAllServicesUptime(c.Context())
				if err != nil {
					return err
				}
				return c.JSON(results)
			})
		}

		mh := &MonitorHandler{monitors: d.Monitors}
		admin.Get("/monitors", mh.List)
		admin.Post("/monitors", middleware.RequireAdmin, mh.Create)
		admin.Get("/monitors/:id", mh.Get)
		admin.Patch("/monitors/:id", middleware.RequireAdmin, mh.Update)
		admin.Delete("/monitors/:id", middleware.RequireAdmin, mh.Archive)
		if d.Uptime != nil {
			admin.Get("/monitors/:id/latency", func(c fiber.Ctx) error {
				pts, err := d.Uptime.GetMonitorLatency(c.Context(), c.Params("id"), 24)
				if err != nil {
					return err
				}
				return c.JSON(pts)
			})
		}

		ih := &IncidentHandler{incidents: d.Incidents}
		admin.Get("/incidents", ih.List)
		admin.Post("/incidents", middleware.RequireAdmin, ih.Create)
		admin.Get("/incidents/:id", ih.Get)
		admin.Patch("/incidents/:id", middleware.RequireAdmin, ih.Update)
		admin.Post("/incidents/:id/resolve", middleware.RequireAdmin, ih.Resolve)
		admin.Get("/incidents/:id/updates", ih.ListUpdates)
		admin.Post("/incidents/:id/updates", middleware.RequireAdmin, ih.PostUpdate)

		akh := &APIKeyHandler{apikeys: d.APIKeys}
		admin.Get("/api-keys", akh.List)
		admin.Post("/api-keys", akh.Create)
		admin.Delete("/api-keys/:id", akh.Revoke)

		nh := &NotificationHandler{notifs: d.Notifications}
		admin.Get("/notifications", nh.List)
		admin.Post("/notifications", middleware.RequireAdmin, nh.Create)
		admin.Get("/notifications/:id", nh.Get)
		admin.Patch("/notifications/:id", middleware.RequireAdmin, nh.Update)
		admin.Delete("/notifications/:id", middleware.RequireAdmin, nh.Archive)

		wh := &WebhookHandler{webhooks: d.Webhooks}
		admin.Get("/webhooks", wh.List)
		admin.Post("/webhooks", middleware.RequireAdmin, wh.Create)
		admin.Get("/webhooks/:id", wh.Get)
		admin.Patch("/webhooks/:id", middleware.RequireAdmin, wh.Update)
		admin.Delete("/webhooks/:id", middleware.RequireAdmin, wh.Archive)
		admin.Get("/webhooks/:id/logs", wh.ListLogs)

		uh := &UserHandler{users: d.Users, cfg: d.Cfg}
		admin.Get("/users", middleware.RequireAdmin, uh.List)
		admin.Post("/users", middleware.RequireAdmin, uh.Create)
		admin.Patch("/users/:id", middleware.RequireAdmin, uh.UpdateRole)
		admin.Delete("/users/:id", middleware.RequireAdmin, uh.Archive)

		if d.Settings != nil {
			sgh := &SettingsHandler{settings: d.Settings}
			admin.Get("/settings", sgh.Get)
			admin.Patch("/settings", middleware.RequireAdmin, sgh.Update)
		}

		if d.Maintenance != nil {
			mwh := &MaintenanceHandler{maintenance: d.Maintenance}
			admin.Get("/maintenance-windows", mwh.List)
			admin.Post("/maintenance-windows", middleware.RequireAdmin, mwh.Create)
			admin.Get("/maintenance-windows/:id", mwh.Get)
			admin.Patch("/maintenance-windows/:id", middleware.RequireAdmin, mwh.Update)
			admin.Delete("/maintenance-windows/:id", middleware.RequireAdmin, mwh.Archive)
		}

		if d.TwoFactor != nil {
			tfh := &TwoFactorHandler{tf: d.TwoFactor}
			admin.Post("/2fa/setup", tfh.InitiateSetup)
			admin.Post("/2fa/confirm", tfh.ConfirmSetup)
			admin.Post("/2fa/disable", middleware.RequireAdmin, tfh.Disable)
		}

		if d.AuditRepo != nil {
			ah := &AuditHandler{repo: d.AuditRepo}
			admin.Get("/audit-log", middleware.RequireAdmin, ah.List)
		}
	}

	// ── Public REST API (/api/v1 always registered; settings guard per-handler) ──
	{
		api := app.Group("/api/v1")

		if d.Settings != nil {
			psh := v1.NewPublicSettingsHandler(d.Settings)
			api.Get("/settings", psh.GetSettings)
		}

		sh := v1.NewStatusHandler(d.Services, d.Monitors)
		api.Get("/status", sh.GetStatus)

		inh := v1.NewIncidentHandler(d.Incidents)
		api.Get("/incidents", inh.List)
		api.Get("/incidents/:id", inh.Get)
		api.Get("/incidents/:id/updates", inh.ListUpdates)

		if d.Uptime != nil {
			uth := v1.NewUptimeHandler(d.Uptime, d.Services)
			api.Get("/uptime", uth.GetAllUptime)
			api.Get("/services/:id/uptime", uth.GetServiceUptime)
			api.Get("/monitors/:id/latency", uth.GetMonitorLatency)
			api.Get("/pages/:slug", uth.GetPageBySlug)
			api.Get("/by-domain", uth.GetPageByDomain)
		}

		if d.Services != nil {
			bh := v1.NewBadgeHandler(d.Services)
			api.Get("/services/:id/badge.svg", bh.GetBadge)
		}

		if d.Settings != nil && d.Incidents != nil {
			rh := v1.NewRSSHandler(d.Incidents, d.Settings)
			api.Get("/rss", rh.GetFeed)
		}
	}

	return app
}

func healthHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// errorHandler maps domain sentinel errors to HTTP status codes.
func errorHandler(c fiber.Ctx, err error) error {
	statusMap := map[error]int{
		domain.ErrNotFound:     fiber.StatusNotFound,
		domain.ErrUnauthorized: fiber.StatusUnauthorized,
		domain.ErrForbidden:    fiber.StatusForbidden,
		domain.ErrConflict:     fiber.StatusConflict,
		domain.ErrValidation:   fiber.StatusUnprocessableEntity,
		domain.ErrArchivedOnly: fiber.StatusConflict,
	}

	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		status, ok := statusMap[appErr.Sentinel]
		if !ok {
			status = fiber.StatusInternalServerError
		}
		body := fiber.Map{"error": appErr.Message}
		if appErr.Field != "" {
			body["field"] = appErr.Field
		}
		return c.Status(status).JSON(body)
	}

	var fe *fiber.Error
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
}
