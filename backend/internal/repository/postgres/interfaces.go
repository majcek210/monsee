package postgres

import "github.com/majcek210/monsee/internal/domain"

// Compile-time interface compliance checks.
var (
	_ domain.UserRepository        = (*UserRepo)(nil)
	_ domain.ServiceRepository     = (*ServiceRepo)(nil)
	_ domain.MonitorRepository     = (*MonitorRepo)(nil)
	_ domain.CheckResultRepository = (*CheckResultRepo)(nil)
	_ domain.IncidentRepository    = (*IncidentRepo)(nil)
	_ domain.APIKeyRepository      = (*APIKeyRepo)(nil)
)
