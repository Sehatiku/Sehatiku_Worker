package job

import (
	"sehatiku-notification-worker/internal/usecase"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Scheduler struct {
	cron *cron.Cron
	log  *zap.Logger
}

func NewScheduler(cfg *viper.Viper, log *zap.Logger, uc *usecase.DailyReminderUseCase) *Scheduler {
	c := cron.New()

	noonJob := &ReminderJob{UseCase: uc, RunType: "noon", Log: log}
	eveningJob := &ReminderJob{UseCase: uc, RunType: "evening", Log: log}

	noonExpr := cfg.GetString("REMINDER_CRON_NOON")
	eveningExpr := cfg.GetString("REMINDER_CRON_EVENING")

	if _, err := c.AddJob(noonExpr, noonJob); err != nil {
		log.Fatal("invalid noon cron expression",
			zap.String("expr", noonExpr), zap.Error(err))
	}
	if _, err := c.AddJob(eveningExpr, eveningJob); err != nil {
		log.Fatal("invalid evening cron expression",
			zap.String("expr", eveningExpr), zap.Error(err))
	}

	log.Info("scheduler configured",
		zap.String("noon", noonExpr),
		zap.String("evening", eveningExpr),
	)

	return &Scheduler{cron: c, log: log}
}

func (s *Scheduler) Start() {
	s.cron.Start()
	s.log.Info("scheduler started")
}

// Stop waits for any running job to finish before returning.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.log.Info("scheduler stopped")
}
