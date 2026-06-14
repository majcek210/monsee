package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/majcek210/monsee/internal/checks"
	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/telemetry"
)

var tracer = otel.Tracer("status-monitor")

// AlertDispatcher is implemented by both the notification and webhook dispatchers.
type AlertDispatcher interface {
	Dispatch(ctx context.Context, event domain.AlertEvent, data map[string]any)
}

// CheckerService runs a single monitor check end-to-end.
type CheckerService struct {
	monitors     domain.MonitorRepository
	checkResults domain.CheckResultRepository
	incidents    domain.IncidentRepository
	dispatchers  []AlertDispatcher
	metrics      *telemetry.Metrics
	log          *zap.Logger
}

func NewCheckerService(
	monitors domain.MonitorRepository,
	checkResults domain.CheckResultRepository,
	incidents domain.IncidentRepository,
	dispatchers ...AlertDispatcher,
) *CheckerService {
	return &CheckerService{
		monitors:    monitors,
		checkResults: checkResults,
		incidents:   incidents,
		dispatchers: dispatchers,
		log:         zap.NewNop(),
	}
}

func (s *CheckerService) WithMetrics(m *telemetry.Metrics) *CheckerService {
	s.metrics = m
	return s
}

func (s *CheckerService) WithLogger(log *zap.Logger) *CheckerService {
	s.log = log
	return s
}

// RunCheck is the single entry point for all check logic.
func (s *CheckerService) RunCheck(ctx context.Context, monitorID string) error {
	ctx, span := tracer.Start(ctx, "RunCheck")
	defer span.End()

	mon, err := s.monitors.GetByID(ctx, monitorID)
	if err != nil {
		return fmt.Errorf("get monitor: %w", err)
	}

	span.SetAttributes(
		attribute.String("monitor.id", mon.ID),
		attribute.String("monitor.name", mon.Name),
		attribute.String("monitor.type", mon.Type),
	)

	// ── Execute check with timing ─────────────────────────────────────────────
	start := time.Now()
	result := checks.Run(ctx, mon)
	elapsed := time.Since(start).Seconds()

	// ── Prometheus metrics ────────────────────────────────────────────────────
	if s.metrics != nil {
		s.metrics.CheckDuration.WithLabelValues(mon.Type).Observe(elapsed)
		s.metrics.MonitorStatusTotal.WithLabelValues(result.Status).Inc()
	}

	s.log.Info("check completed",
		zap.String("monitor_id", mon.ID),
		zap.String("monitor_name", mon.Name),
		zap.String("status", result.Status),
		zap.Int("response_ms", result.ResponseTimeMs),
	)

	// ── Persist result ────────────────────────────────────────────────────────
	var errStr *string
	if result.Error != "" {
		errStr = &result.Error
	}
	responseMs := int32(result.ResponseTimeMs)
	if _, err = s.checkResults.Insert(ctx, domain.InsertCheckResultParams{
		MonitorID:      monitorID,
		Status:         result.Status,
		ResponseTimeMs: &responseMs,
		Error:          errStr,
	}); err != nil {
		s.log.Error("insert check result", zap.Error(err))
	}

	// ── Outage / recovery logic ───────────────────────────────────────────────
	if result.Status == "down" {
		newCount, err := s.incrementAndCheck(ctx, mon)
		if err != nil {
			s.log.Error("increment failures", zap.Error(err))
		} else if newCount >= mon.RetryCount {
			if err := s.handleOutage(ctx, mon); err != nil {
				s.log.Error("handle outage", zap.Error(err))
			}
		}
	} else {
		if mon.ConsecutiveFailures > 0 {
			if err := s.monitors.ResetFailures(ctx, monitorID); err != nil {
				s.log.Error("reset failures", zap.Error(err))
			}
			if err := s.handleRecovery(ctx, mon); err != nil {
				s.log.Error("handle recovery", zap.Error(err))
			}
		}
	}

	// ── Always advance schedule ───────────────────────────────────────────────
	if err := s.monitors.SetNextCheckAt(ctx, monitorID, time.Now()); err != nil {
		s.log.Error("set next check", zap.Error(err))
	}

	return nil
}

func (s *CheckerService) incrementAndCheck(ctx context.Context, mon *domain.Monitor) (int32, error) {
	if err := s.monitors.IncrementFailures(ctx, mon.ID); err != nil {
		return 0, err
	}
	updated, err := s.monitors.GetByID(ctx, mon.ID)
	if err != nil {
		return 0, err
	}
	return updated.ConsecutiveFailures, nil
}

func (s *CheckerService) handleOutage(ctx context.Context, mon *domain.Monitor) error {
	_, err := s.incidents.GetOpenForMonitor(ctx, mon.ID)
	if err == nil {
		return nil // already open — idempotent
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	monID := mon.ID
	if _, err = s.incidents.Create(ctx, domain.CreateIncidentParams{
		ServiceID: mon.ServiceID,
		MonitorID: &monID,
		Title:     fmt.Sprintf("%s is down", mon.Name),
		Severity:  "high",
	}); err != nil {
		return err
	}

	if s.metrics != nil {
		s.metrics.ActiveIncidents.Inc()
	}

	event := domain.AlertEvent{Event: "monitor.down", MonitorID: mon.ID, MonitorName: mon.Name, ServiceID: mon.ServiceID}
	data := map[string]any{"monitor_id": mon.ID, "monitor_name": mon.Name, "service_id": mon.ServiceID}
	for _, d := range s.dispatchers {
		d.Dispatch(ctx, event, data)
	}
	return nil
}

func (s *CheckerService) handleRecovery(ctx context.Context, mon *domain.Monitor) error {
	inc, err := s.incidents.GetOpenForMonitor(ctx, mon.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	if _, err = s.incidents.Resolve(ctx, inc.ID, time.Now()); err != nil {
		return err
	}

	if s.metrics != nil {
		s.metrics.ActiveIncidents.Dec()
	}

	event := domain.AlertEvent{Event: "monitor.recovered", MonitorID: mon.ID, MonitorName: mon.Name, ServiceID: mon.ServiceID}
	data := map[string]any{"monitor_id": mon.ID, "monitor_name": mon.Name, "service_id": mon.ServiceID}
	for _, d := range s.dispatchers {
		d.Dispatch(ctx, event, data)
	}
	return nil
}
