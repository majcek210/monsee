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

		mh := &MonitorHandler{monitors: d.Monitors}
		admin.Get("/monitors", mh.List)
		admin.Post("/monitors", middleware.RequireAdmin, mh.Create)
		admin.Get("/monitors/:id", mh.Get)
		admin.Patch("/monitors/:id", middleware.RequireAdmin, mh.Update)
		admin.Delete("/monitors/:id", middleware.RequireAdmin, mh.Archive)

		ih := &IncidentHandler{incidents: d.Incidents}
		admin.Get("/incidents", ih.List)
		admin.Post("/incidents", middleware.RequireAdmin, ih.Create)
		admin.Get("/incidents/:id", ih.Get)
		admin.Patch("/incidents/:id", middleware.RequireAdmin, ih.Update)
		admin.Post("/incidents/:id/resolve", middleware.RequireAdmin, ih.Resolve)

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
	}

	// ── Public REST API (no auth — read-only status page data) ──────────────
	if d.Cfg.PublicStatusEnabled {
		api := app.Group("/api/v1")

		sh := v1.NewStatusHandler(d.Services, d.Monitors)
		api.Get("/status", sh.GetStatus)

		inh := v1.NewIncidentHandler(d.Incidents)
		api.Get("/incidents", inh.List)
		api.Get("/incidents/:id", inh.Get)
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
