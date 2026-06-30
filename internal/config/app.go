package config

import (
	"sehatiku-notification-worker/internal/gateway/whatsapp"
	"sehatiku-notification-worker/internal/job"
	"sehatiku-notification-worker/internal/repository"
	"sehatiku-notification-worker/internal/usecase"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BootStrapConfig struct {
	DB       *gorm.DB
	Log      *zap.Logger
	Config   *viper.Viper
	WhatsApp *whatsapp.WhatsAppGateway
}

func BootStrap(cfg *BootStrapConfig) *job.Scheduler {
	patientRepo := &repository.PatientRepository{DB: cfg.DB}
	notifRepo := &repository.NotificationRepository{DB: cfg.DB}
	inboxRepo := &repository.PatientNotificationRepository{DB: cfg.DB}
	escalationRepo := &repository.EscalationRepository{DB: cfg.DB}
	trendRepo := &repository.TrendRepository{DB: cfg.DB}

	reminderUC := &usecase.DailyReminderUseCase{
		PatientRepo:      patientRepo,
		NotificationRepo: notifRepo,
		InboxRepo:        inboxRepo,
		WhatsApp:         cfg.WhatsApp,
		Log:              cfg.Log,
	}
	trendUC := &usecase.TrendEscalationUseCase{
		TrendRepo:      trendRepo,
		EscalationRepo: escalationRepo,
		InboxRepo:      inboxRepo,
		WaswasDays:     cfg.Config.GetInt("TREND_WASWAS_DAYS"),
		CooldownDays:   cfg.Config.GetInt("TREND_COOLDOWN_DAYS"),
		Log:            cfg.Log,
	}

	scheduler := job.NewScheduler(cfg.Config, cfg.Log, reminderUC, trendUC)
	scheduler.Start()
	return scheduler
}
