package job

import (
	"context"
	"sehatiku-notification-worker/internal/usecase"

	"go.uber.org/zap"
)

type ReminderJob struct {
	UseCase *usecase.DailyReminderUseCase
	RunType string // "noon" or "evening"
	Log     *zap.Logger
}

func (j *ReminderJob) Run() {
	j.Log.Info("cron triggered", zap.String("run_type", j.RunType))
	summary := j.UseCase.Run(context.Background(), j.RunType)
	j.Log.Info("cron finished",
		zap.String("run_type", j.RunType),
		zap.Int("sent", summary.Sent),
		zap.Int("failed", summary.Failed),
		zap.Int("skipped", summary.SkippedAlreadySent),
	)
}
