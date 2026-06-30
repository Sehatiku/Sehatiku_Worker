package job

import (
	"sehatiku-notification-worker/internal/usecase"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Scheduler struct {
	cron *cron.Cron
	log  *zap.Logger
}

// appLocation memuat zona waktu aplikasi dari APP_TIMEZONE (default Asia/Jakarta), dengan
// fallback statis UTC+7 bila tzdata tidak tersedia di host. Pola ini identik dengan main
// backend (patient_dashboard_usecase.go) agar worker & backend memakai "hari ini" yang sama.
func appLocation(cfg *viper.Viper) *time.Location {
	name := cfg.GetString("APP_TIMEZONE")
	if name == "" {
		name = "Asia/Jakarta"
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}

func NewScheduler(cfg *viper.Viper, log *zap.Logger, uc *usecase.DailyReminderUseCase, trendUC *usecase.TrendEscalationUseCase) *Scheduler {
	loc := appLocation(cfg)
	c := cron.New(cron.WithLocation(loc))

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

	trendJob := &TrendEscalationJob{UseCase: trendUC, Log: log}
	trendExpr := cfg.GetString("TREND_ESCALATION_CRON")
	if trendExpr == "" {
		trendExpr = "0 6 * * *" // default 06:00 WIB harian
	}
	if _, err := c.AddJob(trendExpr, trendJob); err != nil {
		log.Fatal("invalid trend escalation cron expression",
			zap.String("expr", trendExpr), zap.Error(err))
	}

	log.Info("scheduler configured",
		zap.String("noon", noonExpr),
		zap.String("evening", eveningExpr),
		zap.String("trend_escalation", trendExpr),
		zap.String("timezone", loc.String()),
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
