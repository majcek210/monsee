package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/domain"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

type MonitorRepo struct {
	q *sqlcdb.Queries
}

func NewMonitorRepo(pool *pgxpool.Pool) *MonitorRepo {
	return &MonitorRepo{q: sqlcdb.New(pool)}
}

func (r *MonitorRepo) Create(ctx context.Context, p domain.CreateMonitorParams) (*domain.Monitor, error) {
	sid, err := parseUUID(p.ServiceID)
	if err != nil {
		return nil, domain.ValidationErr("service_id", "invalid service_id")
	}
	row, err := r.q.CreateMonitor(ctx, sqlcdb.CreateMonitorParams{
		ServiceID:              sid,
		Name:                   p.Name,
		Type:                   p.Type,
		Url:                    p.URL,
		Host:                   p.Host,
		Port:                   p.Port,
		IntervalSeconds:        p.IntervalSeconds,
		TimeoutMs:              p.TimeoutMs,
		RetryCount:             p.RetryCount,
		DegradedThresholdMs:    p.DegradedThresholdMs,
		HttpMethod:             p.HTTPMethod,
		HttpExpectedStatus:     p.HTTPExpectedStatus,
		SslExpiryThresholdDays: p.SSLExpiryThresholdDays,
		KeywordMatch:           p.KeywordMatch,
		KeywordShouldExist:     p.KeywordShouldExist,
		DnsRecordType:          p.DNSRecordType,
		DnsExpectedValue:       p.DNSExpectedValue,
	})
	if err != nil {
		return nil, err
	}
	return monitorToDomain(row), nil
}

func (r *MonitorRepo) GetByID(ctx context.Context, id string) (*domain.Monitor, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("monitor not found")
	}
	row, err := r.q.GetMonitorByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("monitor not found")
		}
		return nil, err
	}
	return monitorToDomain(row), nil
}

func (r *MonitorRepo) ListByService(ctx context.Context, serviceID string) ([]*domain.Monitor, error) {
	sid, err := parseUUID(serviceID)
	if err != nil {
		return nil, domain.ValidationErr("service_id", "invalid service_id")
	}
	rows, err := r.q.ListMonitorsByService(ctx, sid)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Monitor, len(rows))
	for i, row := range rows {
		out[i] = monitorToDomain(row)
	}
	return out, nil
}

func (r *MonitorRepo) ListDue(ctx context.Context) ([]*domain.Monitor, error) {
	rows, err := r.q.ListDueMonitors(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Monitor, len(rows))
	for i, row := range rows {
		out[i] = monitorToDomain(row)
	}
	return out, nil
}

func (r *MonitorRepo) Update(ctx context.Context, id string, p domain.UpdateMonitorParams) (*domain.Monitor, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("monitor not found")
	}
	row, err := r.q.UpdateMonitor(ctx, sqlcdb.UpdateMonitorParams{
		ID:                     uid,
		Name:                   p.Name,
		Url:                    p.URL,
		Host:                   p.Host,
		Port:                   p.Port,
		IntervalSeconds:        p.IntervalSeconds,
		TimeoutMs:              p.TimeoutMs,
		RetryCount:             p.RetryCount,
		DegradedThresholdMs:    p.DegradedThresholdMs,
		HttpMethod:             p.HTTPMethod,
		HttpExpectedStatus:     p.HTTPExpectedStatus,
		Enabled:                p.Enabled,
		SslExpiryThresholdDays: p.SSLExpiryThresholdDays,
		KeywordMatch:           p.KeywordMatch,
		KeywordShouldExist:     p.KeywordShouldExist,
		DnsRecordType:          p.DNSRecordType,
		DnsExpectedValue:       p.DNSExpectedValue,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("monitor not found")
		}
		return nil, err
	}
	return monitorToDomain(row), nil
}

func (r *MonitorRepo) IncrementFailures(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("monitor not found")
	}
	_, err = r.q.IncrementConsecutiveFailures(ctx, uid)
	return err
}

func (r *MonitorRepo) ResetFailures(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("monitor not found")
	}
	return r.q.ResetConsecutiveFailures(ctx, uid)
}

func (r *MonitorRepo) SetNextCheckAt(ctx context.Context, id string, _ time.Time) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("monitor not found")
	}
	return r.q.SetNextCheckAt(ctx, uid)
}

func (r *MonitorRepo) Archive(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("monitor not found")
	}
	return r.q.ArchiveMonitor(ctx, uid)
}

func monitorToDomain(m sqlcdb.Monitor) *domain.Monitor {
	return &domain.Monitor{
		ID:                     uuidStr(m.ID),
		ServiceID:              uuidStr(m.ServiceID),
		Name:                   m.Name,
		Type:                   m.Type,
		URL:                    m.Url,
		Host:                   m.Host,
		Port:                   m.Port,
		IntervalSeconds:        m.IntervalSeconds,
		TimeoutMs:              m.TimeoutMs,
		RetryCount:             m.RetryCount,
		ConsecutiveFailures:    m.ConsecutiveFailures,
		DegradedThresholdMs:    m.DegradedThresholdMs,
		HTTPMethod:             m.HttpMethod,
		HTTPExpectedStatus:     m.HttpExpectedStatus,
		SSLExpiryThresholdDays: m.SslExpiryThresholdDays,
		KeywordMatch:           m.KeywordMatch,
		KeywordShouldExist:     m.KeywordShouldExist,
		DNSRecordType:          m.DnsRecordType,
		DNSExpectedValue:       m.DnsExpectedValue,
		Enabled:                m.Enabled,
		NextCheckAt:            tsToTime(m.NextCheckAt),
		CreatedAt:              tsToTime(m.CreatedAt),
		UpdatedAt:              tsToTime(m.UpdatedAt),
		ArchivedAt:             tsToTimePtr(m.ArchivedAt),
	}
}
