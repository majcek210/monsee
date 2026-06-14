package postgres

import (
	"context"

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
