package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/domain"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

type WebhookRepo struct {
	q *sqlcdb.Queries
}

func NewWebhookRepo(pool *pgxpool.Pool) *WebhookRepo {
	return &WebhookRepo{q: sqlcdb.New(pool)}
}

func (r *WebhookRepo) Create(ctx context.Context, p domain.CreateWebhookParams) (*domain.Webhook, error) {
	row, err := r.q.CreateWebhook(ctx, sqlcdb.CreateWebhookParams{
		Name:   p.Name,
		Url:    p.URL,
		Secret: p.Secret,
		Events: p.Events,
	})
	if err != nil {
		return nil, err
	}
	return webhookToDomain(row), nil
}

func (r *WebhookRepo) GetByID(ctx context.Context, id string) (*domain.Webhook, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("webhook not found")
	}
	row, err := r.q.GetWebhookByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("webhook not found")
		}
		return nil, err
	}
	return webhookToDomain(row), nil
}

func (r *WebhookRepo) List(ctx context.Context) ([]*domain.Webhook, error) {
	rows, err := r.q.ListWebhooks(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Webhook, len(rows))
	for i, row := range rows {
		out[i] = webhookToDomain(row)
	}
	return out, nil
}

func (r *WebhookRepo) ListByEvent(ctx context.Context, event string) ([]*domain.Webhook, error) {
	rows, err := r.q.ListWebhooksByEvent(ctx, event)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Webhook, len(rows))
	for i, row := range rows {
		out[i] = webhookToDomain(row)
	}
	return out, nil
}

func (r *WebhookRepo) Update(ctx context.Context, id string, p domain.UpdateWebhookParams) (*domain.Webhook, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("webhook not found")
	}
	row, err := r.q.UpdateWebhook(ctx, sqlcdb.UpdateWebhookParams{
		ID:      uid,
		Name:    p.Name,
		Url:     p.URL,
		Secret:  p.Secret,
		Events:  p.Events,
		Enabled: p.Enabled,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("webhook not found")
		}
		return nil, err
	}
	return webhookToDomain(row), nil
}

func (r *WebhookRepo) Archive(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("webhook not found")
	}
	return r.q.ArchiveWebhook(ctx, uid)
}

func (r *WebhookRepo) InsertLog(ctx context.Context, p domain.InsertWebhookLogParams) (*domain.WebhookLog, error) {
	wid, err := parseUUID(p.WebhookID)
	if err != nil {
		return nil, domain.ValidationErr("webhook_id", "invalid webhook_id")
	}
	row, err := r.q.InsertWebhookLog(ctx, sqlcdb.InsertWebhookLogParams{
		WebhookID:  wid,
		Event:      p.Event,
		StatusCode: p.StatusCode,
		Error:      p.Error,
		DurationMs: p.DurationMs,
	})
	if err != nil {
		return nil, err
	}
	return webhookLogToDomain(row), nil
}

func (r *WebhookRepo) ListLogs(ctx context.Context, webhookID string, limit int32) ([]*domain.WebhookLog, error) {
	wid, err := parseUUID(webhookID)
	if err != nil {
		return nil, domain.ValidationErr("webhook_id", "invalid webhook_id")
	}
	rows, err := r.q.ListWebhookLogs(ctx, sqlcdb.ListWebhookLogsParams{
		WebhookID: wid,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.WebhookLog, len(rows))
	for i, row := range rows {
		out[i] = webhookLogToDomain(row)
	}
	return out, nil
}

func webhookToDomain(w sqlcdb.Webhook) *domain.Webhook {
	return &domain.Webhook{
		ID:         uuidStr(w.ID),
		Name:       w.Name,
		URL:        w.Url,
		Secret:     w.Secret,
		Events:     w.Events,
		Enabled:    w.Enabled,
		CreatedAt:  tsToTime(w.CreatedAt),
		ArchivedAt: tsToTimePtr(w.ArchivedAt),
	}
}

func webhookLogToDomain(l sqlcdb.WebhookLog) *domain.WebhookLog {
	return &domain.WebhookLog{
		ID:          uuidStr(l.ID),
		WebhookID:   uuidStr(l.WebhookID),
		Event:       l.Event,
		StatusCode:  l.StatusCode,
		Error:       l.Error,
		DurationMs:  l.DurationMs,
		DeliveredAt: tsToTime(l.DeliveredAt),
	}
}
