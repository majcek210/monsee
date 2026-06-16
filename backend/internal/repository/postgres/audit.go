package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/middleware"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

// AuditRepo implements middleware.AuditLogger backed by Postgres.
type AuditRepo struct {
	q *sqlcdb.Queries
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{q: sqlcdb.New(pool)}
}

// AuditLogEntry is a read-model for the audit log viewer.
type AuditLogEntry struct {
	ID         string         `json:"id"`
	UserID     *string        `json:"user_id"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	ResourceID *string        `json:"resource_id"`
	IP         *string        `json:"ip"`
	UserAgent  *string        `json:"user_agent"`
	Diff       map[string]any `json:"diff"`
	CreatedAt  time.Time      `json:"created_at"`
}

// AuditListParams for paginated queries.
type AuditListParams struct {
	Resource *string
	UserID   *string
	Limit    int32
	Offset   int32
}

// List returns a paginated page of audit log entries.
func (r *AuditRepo) List(ctx context.Context, p AuditListParams) ([]AuditLogEntry, int64, error) {
	if p.Limit == 0 {
		p.Limit = 50
	}
	filterUID := parseUUIDPtr(p.UserID)
	rows, err := r.q.ListAuditLog(ctx, sqlcdb.ListAuditLogParams{
		Resource:     p.Resource,
		FilterUserID: filterUID,
		Limit:        p.Limit,
		Offset:       p.Offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.q.CountAuditLog(ctx, sqlcdb.CountAuditLogParams{
		Resource:     p.Resource,
		FilterUserID: filterUID,
	})
	if err != nil {
		return nil, 0, err
	}
	entries := make([]AuditLogEntry, len(rows))
	for i, row := range rows {
		entries[i] = auditRowToEntry(row)
	}
	return entries, total, nil
}

func auditRowToEntry(row sqlcdb.AuditLog) AuditLogEntry {
	e := AuditLogEntry{
		ID:         uuidStr(row.ID),
		Action:     row.Action,
		Resource:   row.Resource,
		ResourceID: row.ResourceID,
		IP:         row.Ip,
		UserAgent:  row.UserAgent,
	}
	if row.UserID.Valid {
		s := uuidStr(row.UserID)
		e.UserID = &s
	}
	if row.CreatedAt.Valid {
		e.CreatedAt = row.CreatedAt.Time
	}
	if len(row.Diff) > 0 {
		_ = json.Unmarshal(row.Diff, &e.Diff)
	}
	return e
}

func (r *AuditRepo) Log(ctx context.Context, entry middleware.AuditEntry) error {
	var diff []byte
	if entry.Diff != nil {
		diff, _ = json.Marshal(entry.Diff)
	}

	var userID *string
	if entry.UserID != "" {
		userID = &entry.UserID
	}
	var resourceID *string
	if entry.ResourceID != "" {
		resourceID = &entry.ResourceID
	}
	var ip *string
	if entry.IP != "" {
		ip = &entry.IP
	}
	var ua *string
	if entry.UserAgent != "" {
		ua = &entry.UserAgent
	}

	uid := parseUUIDPtr(userID)

	_, err := r.q.InsertAuditLog(ctx, sqlcdb.InsertAuditLogParams{
		UserID:     uid,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: resourceID,
		Ip:         ip,
		UserAgent:  ua,
		Diff:       diff,
	})
	return err
}
