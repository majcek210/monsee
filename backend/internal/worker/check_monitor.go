package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/majcek210/monsee/internal/service"
)

type CheckMonitorHandler struct {
	checker *service.CheckerService
}

func NewCheckMonitorHandler(checker *service.CheckerService) *CheckMonitorHandler {
	return &CheckMonitorHandler{checker: checker}
}

func (h *CheckMonitorHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p MonitorCheckPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	return h.checker.RunCheck(ctx, p.MonitorID)
}
