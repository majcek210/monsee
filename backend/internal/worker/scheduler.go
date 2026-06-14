package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"github.com/majcek210/monsee/internal/domain"
)

// Scheduler polls for due monitors every 15 seconds and enqueues check tasks.
type Scheduler struct {
	monitors domain.MonitorRepository
	client   *asynq.Client
}

func NewScheduler(monitors domain.MonitorRepository, client *asynq.Client) *Scheduler {
	return &Scheduler{monitors: monitors, client: client}
}

// Run blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Enqueue immediately on startup
	s.enqueue(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.enqueue(ctx)
		}
	}
}

func (s *Scheduler) enqueue(ctx context.Context) {
	due, err := s.monitors.ListDue(ctx)
	if err != nil {
		log.Printf("scheduler list due: %v", err)
		return
	}

	for _, m := range due {
		payload, _ := json.Marshal(MonitorCheckPayload{MonitorID: m.ID})
		task := asynq.NewTask(TaskRunMonitorCheck, payload,
			asynq.TaskID("check:"+m.ID),
			asynq.MaxRetry(3),
			asynq.Timeout(30*time.Second),
			asynq.Queue("monitors"),
		)
		_, err := s.client.EnqueueContext(ctx, task)
		if err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) {
			log.Printf("enqueue monitor %s: %v", m.ID, err)
		}
	}
}
