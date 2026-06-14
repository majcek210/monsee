package postgres

import (
	"context"
	"encoding/json"

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
