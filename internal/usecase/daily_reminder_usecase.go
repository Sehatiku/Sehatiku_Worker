package usecase

import (
	"context"
	"encoding/json"
	"sehatiku-notification-worker/internal/entity"
	"sehatiku-notification-worker/internal/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WhatsAppSender is the outbound messaging interface used by the usecase.
// *whatsapp.WhatsAppGateway satisfies this interface via Go duck typing.
type WhatsAppSender interface {
	SendPatientReminder(ctx context.Context, phone, patientName, runType string) error
	SendCompanionReminder(ctx context.Context, phone, companionName, patientName string) error
}

type RunSummary struct {
	Sent               int
	Failed             int
	SkippedAlreadySent int
}

type DailyReminderUseCase struct {
	PatientRepo      repository.PatientQuerier
	NotificationRepo repository.NotificationStore
	InboxRepo        repository.InboxStore
	WhatsApp         WhatsAppSender
	Log              *zap.Logger
}

func (uc *DailyReminderUseCase) Run(ctx context.Context, runType string) RunSummary {
	start := time.Now()
	summary := RunSummary{}

	uc.Log.Info("reminder job started", zap.String("run", runType))

	patients, err := uc.PatientRepo.FindPatientsWithNoLogToday(ctx)
	if err != nil {
		uc.Log.Error("failed to query patients with no log", zap.Error(err))
		return summary
	}

	uc.Log.Info("patients to notify", zap.Int("count", len(patients)), zap.String("run", runType))

	for _, p := range patients {
		uc.processPatient(ctx, p, runType, &summary)
	}

	uc.Log.Info("reminder job finished",
		zap.String("run", runType),
		zap.Int("sent", summary.Sent),
		zap.Int("failed", summary.Failed),
		zap.Int("skipped_already_sent", summary.SkippedAlreadySent),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()),
	)
	return summary
}

func (uc *DailyReminderUseCase) processPatient(ctx context.Context, p entity.Patient, runType string, s *RunSummary) {
	exists, err := uc.NotificationRepo.ExistsForToday(ctx, p.ID)
	if err != nil {
		uc.Log.Warn("dedup check failed", zap.String("patient_id", p.ID), zap.Error(err))
		return
	}
	if exists {
		s.SkippedAlreadySent++
		return
	}

	// Catat pengingat ke inbox in-app pasien (patient_notifications). Independen dari WA:
	// pasien tetap melihat pengingat di app meski WA gagal. Best-effort — hanya pasien
	// (pendamping tidak punya akun app), gagal cukup di-log. Dedup harian sudah dijaga di atas.
	uc.createInboxReminder(ctx, p.ID, runType)

	patientNotif := &entity.Notification{
		ID:             uuid.New().String(),
		PatientID:      &p.ID,
		RecipientPhone: p.PhoneNumber,
		RecipientRole:  "patient",
		MessageType:    "daily_prompt",
		Channel:        "whatsapp",
		Payload:        mustMarshal(map[string]string{"run_type": runType}),
		Status:         "queued",
	}
	if err := uc.NotificationRepo.Create(ctx, patientNotif); err != nil {
		uc.Log.Warn("failed to create patient notification record",
			zap.String("patient_id", p.ID), zap.Error(err))
		return
	}

	if err := uc.WhatsApp.SendPatientReminder(ctx, p.PhoneNumber, p.FullName, runType); err != nil {
		errStr := err.Error()
		_ = uc.NotificationRepo.UpdateStatus(ctx, patientNotif.ID, "failed", &errStr)
		s.Failed++
		uc.Log.Warn("failed to send reminder to patient",
			zap.String("patient_id", p.ID), zap.Error(err))
		return
	}
	_ = uc.NotificationRepo.UpdateStatus(ctx, patientNotif.ID, "sent", nil)
	s.Sent++

	if p.CompanionPhone == nil || *p.CompanionPhone == "" {
		return
	}

	companionNotif := &entity.Notification{
		ID:             uuid.New().String(),
		PatientID:      &p.ID,
		RecipientPhone: *p.CompanionPhone,
		RecipientRole:  "companion",
		MessageType:    "daily_prompt",
		Channel:        "whatsapp",
		Payload:        mustMarshal(map[string]string{"run_type": runType}),
		Status:         "queued",
	}
	if err := uc.NotificationRepo.Create(ctx, companionNotif); err != nil {
		uc.Log.Warn("failed to create companion notification record",
			zap.String("patient_id", p.ID), zap.Error(err))
		return
	}

	companionName := ""
	if p.CompanionName != nil {
		companionName = *p.CompanionName
	}
	if err := uc.WhatsApp.SendCompanionReminder(ctx, *p.CompanionPhone, companionName, p.FullName); err != nil {
		errStr := err.Error()
		_ = uc.NotificationRepo.UpdateStatus(ctx, companionNotif.ID, "failed", &errStr)
		uc.Log.Warn("failed to send reminder to companion",
			zap.String("patient_id", p.ID), zap.Error(err))
		return
	}
	_ = uc.NotificationRepo.UpdateStatus(ctx, companionNotif.ID, "sent", nil)
}

func (uc *DailyReminderUseCase) createInboxReminder(ctx context.Context, patientID, runType string) {
	if uc.InboxRepo == nil {
		return
	}
	body := "Hari ini belum ada catatan kesehatanmu. Masih sempat, catat sekarang ya."
	if runType == "noon" {
		body = "Jangan lupa catat data kesehatanmu hari ini ya."
	}
	notif := &entity.PatientNotification{
		ID:        uuid.New().String(),
		PatientID: patientID,
		Type:      entity.PatientNotifTypeDailyReminder,
		Title:     "Pengingat catatan harian",
		Body:      body,
		Payload:   mustMarshal(map[string]string{"run_type": runType}),
	}
	if err := uc.InboxRepo.Create(ctx, notif); err != nil {
		uc.Log.Warn("failed to create in-app daily reminder",
			zap.String("patient_id", patientID), zap.Error(err))
	}
}

func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

