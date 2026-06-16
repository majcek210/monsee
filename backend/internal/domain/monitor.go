package domain

import (
	"context"
	"time"
)

type Monitor struct {
	ID                       string     `json:"id"`
	ServiceID                string     `json:"service_id"`
	Name                     string     `json:"name"`
	Type                     string     `json:"type"`
	URL                      *string    `json:"url"`
	Host                     *string    `json:"host"`
	Port                     *int32     `json:"port"`
	IntervalSeconds          int32      `json:"interval_seconds"`
	TimeoutMs                int32      `json:"timeout_ms"`
	RetryCount               int32      `json:"retry_count"`
	ConsecutiveFailures      int32      `json:"consecutive_failures"`
	DegradedThresholdMs      *int32     `json:"degraded_threshold_ms"`
	HTTPMethod               *string    `json:"http_method"`
	HTTPExpectedStatus       *int32     `json:"http_expected_status"`
	SSLExpiryThresholdDays   int32      `json:"ssl_expiry_threshold_days"`
	KeywordMatch             *string    `json:"keyword_match"`
	KeywordShouldExist       bool       `json:"keyword_should_exist"`
	DNSRecordType            *string    `json:"dns_record_type"`
	DNSExpectedValue         *string    `json:"dns_expected_value"`
	Enabled                  bool       `json:"enabled"`
	NextCheckAt              time.Time  `json:"next_check_at"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	ArchivedAt               *time.Time `json:"archived_at"`
}

type CreateMonitorParams struct {
	ServiceID              string
	Name                   string
	Type                   string
	URL                    *string
	Host                   *string
	Port                   *int32
	IntervalSeconds        int32
	TimeoutMs              int32
	RetryCount             int32
	DegradedThresholdMs    *int32
	HTTPMethod             *string
	HTTPExpectedStatus     *int32
	SSLExpiryThresholdDays int32
	KeywordMatch           *string
	KeywordShouldExist     bool
	DNSRecordType          *string
	DNSExpectedValue       *string
}

type UpdateMonitorParams struct {
	Name                   string
	URL                    *string
	Host                   *string
	Port                   *int32
	IntervalSeconds        int32
	TimeoutMs              int32
	RetryCount             int32
	DegradedThresholdMs    *int32
	HTTPMethod             *string
	HTTPExpectedStatus     *int32
	Enabled                bool
	SSLExpiryThresholdDays int32
	KeywordMatch           *string
	KeywordShouldExist     bool
	DNSRecordType          *string
	DNSExpectedValue       *string
}

type DailyUptime struct {
	Day      time.Time `json:"day"`
	Total    int64     `json:"total"`
	Up       int64     `json:"up"`
	Down     int64     `json:"down"`
	Degraded int64     `json:"degraded"`
}

type ResponseTimePoint struct {
	CheckedAt  time.Time `json:"checked_at"`
	ResponseMs int32     `json:"response_ms"`
	Status     string    `json:"status"`
}

type CheckResult struct {
	ID             string    `json:"id"`
	MonitorID      string    `json:"monitor_id"`
	Status         string    `json:"status"`
	ResponseTimeMs *int32    `json:"response_time_ms"`
	Error          *string   `json:"error"`
	CheckedAt      time.Time `json:"checked_at"`
}

type InsertCheckResultParams struct {
	MonitorID      string
	Status         string
	ResponseTimeMs *int32
	Error          *string
}

type MonitorRepository interface {
	Create(ctx context.Context, p CreateMonitorParams) (*Monitor, error)
	GetByID(ctx context.Context, id string) (*Monitor, error)
	ListByService(ctx context.Context, serviceID string) ([]*Monitor, error)
	ListDue(ctx context.Context) ([]*Monitor, error)
	Update(ctx context.Context, id string, p UpdateMonitorParams) (*Monitor, error)
	IncrementFailures(ctx context.Context, id string) error
	ResetFailures(ctx context.Context, id string) error
	SetNextCheckAt(ctx context.Context, id string, t time.Time) error
	Archive(ctx context.Context, id string) error
}

type CheckResultRepository interface {
	Insert(ctx context.Context, p InsertCheckResultParams) (*CheckResult, error)
	ListByMonitor(ctx context.Context, monitorID string, limit int32) ([]*CheckResult, error)
	ListDailyUptime(ctx context.Context, monitorID string, days int32) ([]*DailyUptime, error)
	ListResponseTimes(ctx context.Context, monitorID string, hours int32) ([]*ResponseTimePoint, error)
}
