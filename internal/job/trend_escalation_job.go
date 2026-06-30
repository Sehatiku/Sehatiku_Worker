package job

import (
	"context"
	"sehatiku-notification-worker/internal/usecase"

	"go.uber.org/zap"
)

type TrendEscalationJob struct {
	UseCase *usecase.TrendEscalationUseCase
	Log     *zap.Logger
}

func (j *TrendEscalationJob) Run() {
	j.Log.Info("cron triggered", zap.String("job", "trend_escalation"))
	s := j.UseCase.Run(context.Background())
	j.Log.Info("cron finished",
		zap.String("job", "trend_escalation"),
		zap.Int("created", s.Created),
		zap.Int("skipped", s.Skipped),
		zap.Int("failed", s.Failed),
	)
}
