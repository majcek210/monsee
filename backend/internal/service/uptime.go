package service

import (
	"context"

	"github.com/majcek210/monsee/internal/domain"
)

type DailyUptimeStatus struct {
	Date          string  `json:"date"`
	Status        string  `json:"status"`
	UptimePercent float64 `json:"uptime_percent"`
}

type MonitorUptime struct {
	MonitorID   string              `json:"monitor_id"`
	MonitorName string              `json:"monitor_name"`
	Days        []DailyUptimeStatus `json:"days"`
}

type ServiceUptime struct {
	ServiceID string          `json:"service_id"`
	Monitors  []MonitorUptime `json:"monitors"`
}

type UptimeService struct {
	services     domain.ServiceRepository
	monitors     domain.MonitorRepository
	checkResults domain.CheckResultRepository
}

func NewUptimeService(services domain.ServiceRepository, monitors domain.MonitorRepository, checkResults domain.CheckResultRepository) *UptimeService {
	return &UptimeService{services: services, monitors: monitors, checkResults: checkResults}
}

func (s *UptimeService) GetServiceUptime(ctx context.Context, serviceID string, days *int32) (*ServiceUptime, error) {
	svc, err := s.services.GetByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	rangeDays := svc.UptimeRangeDays
	if days != nil {
		rangeDays = *days
	}
	if rangeDays <= 0 {
		rangeDays = 90
	}

	monitors, err := s.monitors.ListByService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	monitorUptimes := make([]MonitorUptime, 0, len(monitors))
	for _, m := range monitors {
		if m.ArchivedAt != nil {
			continue
		}
		rows, err := s.checkResults.ListDailyUptime(ctx, m.ID, rangeDays)
		if err != nil {
			rows = []*domain.DailyUptime{}
		}
		dayStatuses := make([]DailyUptimeStatus, len(rows))
		for i, row := range rows {
			status := "no_data"
			var pct float64
			if row.Total > 0 {
				if row.Down > 0 {
					status = "down"
				} else if row.Degraded > 0 {
					status = "degraded"
				} else {
					status = "up"
				}
				pct = float64(row.Up+row.Degraded) / float64(row.Total) * 100
			}
			dayStatuses[i] = DailyUptimeStatus{
				Date:          row.Day.Format("2006-01-02"),
				Status:        status,
				UptimePercent: pct,
			}
		}
		monitorUptimes = append(monitorUptimes, MonitorUptime{
			MonitorID:   m.ID,
			MonitorName: m.Name,
			Days:        dayStatuses,
		})
	}

	return &ServiceUptime{
		ServiceID: serviceID,
		Monitors:  monitorUptimes,
	}, nil
}

func (s *UptimeService) GetAllServicesUptime(ctx context.Context) ([]*ServiceUptime, error) {
	svcs, err := s.services.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*ServiceUptime, 0, len(svcs))
	for _, svc := range svcs {
		uptime, err := s.GetServiceUptime(ctx, svc.ID, nil)
		if err != nil {
			continue
		}
		out = append(out, uptime)
	}
	return out, nil
}

func (s *UptimeService) GetMonitorLatency(ctx context.Context, monitorID string, hours int32) ([]*domain.ResponseTimePoint, error) {
	return s.checkResults.ListResponseTimes(ctx, monitorID, hours)
}
