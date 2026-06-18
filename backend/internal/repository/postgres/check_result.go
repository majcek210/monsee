package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/domain"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

type CheckResultRepo struct {
	q *sqlcdb.Queries
}

func NewCheckResultRepo(pool *pgxpool.Pool) *CheckResultRepo {
	return &CheckResultRepo{q: sqlcdb.New(pool)}
}

func (r *CheckResultRepo) Insert(ctx context.Context, p domain.InsertCheckResultParams) (*domain.CheckResult, error) {
	mid, err := parseUUID(p.MonitorID)
	if err != nil {
		return nil, domain.ValidationErr("monitor_id", "invalid monitor_id")
	}
	row, err := r.q.InsertCheckResult(ctx, sqlcdb.InsertCheckResultParams{
		MonitorID:      mid,
		Status:         p.Status,
		ResponseTimeMs: p.ResponseTimeMs,
		Error:          p.Error,
	})
	if err != nil {
		return nil, err
	}
	return checkResultToDomain(row), nil
}

func (r *CheckResultRepo) ListByMonitor(ctx context.Context, monitorID string, limit int32) ([]*domain.CheckResult, error) {
	mid, err := parseUUID(monitorID)
	if err != nil {
		return nil, domain.ValidationErr("monitor_id", "invalid monitor_id")
	}
	rows, err := r.q.ListCheckResultsByMonitor(ctx, sqlcdb.ListCheckResultsByMonitorParams{
		MonitorID: mid,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.CheckResult, len(rows))
	for i, row := range rows {
		out[i] = checkResultToDomain(row)
	}
	return out, nil
}

func (r *CheckResultRepo) ListDailyUptime(ctx context.Context, monitorID string, days int32) ([]*domain.DailyUptime, error) {
	mid, err := parseUUID(monitorID)
	if err != nil {
		return nil, domain.ValidationErr("monitor_id", "invalid monitor_id")
	}
	rows, err := r.q.GetDailyUptimeForMonitor(ctx, sqlcdb.GetDailyUptimeForMonitorParams{
		MonitorID: mid,
		Column2:   days,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.DailyUptime, len(rows))
	for i, row := range rows {
		var day time.Time
		if row.Day.Valid {
			day = time.Date(int(row.Day.Time.Year()), row.Day.Time.Month(), int(row.Day.Time.Day()), 0, 0, 0, 0, time.UTC)
		}
		out[i] = &domain.DailyUptime{
			Day:      day,
			Total:    row.Total,
			Up:       row.UpCount,
			Down:     row.DownCount,
			Degraded: row.DegradedCount,
		}
	}
	return out, nil
}

func (r *CheckResultRepo) ListResponseTimes(ctx context.Context, monitorID string, hours int32) ([]*domain.ResponseTimePoint, error) {
	mid, err := parseUUID(monitorID)
	if err != nil {
		return nil, domain.ValidationErr("monitor_id", "invalid monitor_id")
	}
	rows, err := r.q.ListResponseTimes(ctx, sqlcdb.ListResponseTimesParams{
		MonitorID: mid,
		Column2:   hours,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.ResponseTimePoint, len(rows))
	for i, row := range rows {
		var ms int32
		if row.ResponseTimeMs != nil {
			ms = *row.ResponseTimeMs
		}
		out[i] = &domain.ResponseTimePoint{
			CheckedAt:  tsToTime(row.CheckedAt),
			ResponseMs: ms,
			Status:     row.Status,
		}
	}
	return out, nil
}

func checkResultToDomain(c sqlcdb.CheckResult) *domain.CheckResult {
	return &domain.CheckResult{
		ID:             uuidStr(c.ID),
		MonitorID:      uuidStr(c.MonitorID),
		Status:         c.Status,
		ResponseTimeMs: c.ResponseTimeMs,
		Error:          c.Error,
		CheckedAt:      tsToTime(c.CheckedAt),
	}
}
