package worker

const (
	TaskRunMonitorCheck   = "monitor:check"
	TaskCleanupOldResults = "cleanup:results"
)

type MonitorCheckPayload struct {
	MonitorID string `json:"monitor_id"`
}
