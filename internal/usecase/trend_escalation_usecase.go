package usecase

import (
	"context"
	"sehatiku-notification-worker/internal/entity"
	"sehatiku-notification-worker/internal/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TrendSummary struct {
	Created int
	Skipped int
	Failed  int
}

type TrendEscalationUseCase struct {
	TrendRepo      repository.TrendQuerier
	EscalationRepo repository.EscalationStore
	InboxRepo      repository.InboxStore
	WaswasDays     int // ambang hari berisiko dalam 7 hari; <=0 -> default 3
	CooldownDays   int // jeda sebelum eskalasi tren ulang utk pasien sama; <=0 -> default 7
	Log            *zap.Logger
}

func (uc *TrendEscalationUseCase) Run(ctx context.Context) TrendSummary {
	start := time.Now()
	s := TrendSummary{}
	uc.Log.Info("trend escalation job started")

	candidates, err := uc.TrendRepo.FindTrendCandidates(ctx)
	if err != nil {
		uc.Log.Error("failed to query trend candidates", zap.Error(err))
		return s
	}
	uc.Log.Info("trend candidates", zap.Int("count", len(candidates)))

	threshold := uc.WaswasDays
	if threshold <= 0 {
		threshold = 3
	}
	cooldownDays := uc.CooldownDays
	if cooldownDays <= 0 {
		cooldownDays = 7
	}
	cooldownSince := time.Now().AddDate(0, 0, -cooldownDays)

	for _, c := range candidates {
		uc.process(ctx, c, threshold, cooldownSince, &s)
	}

	uc.Log.Info("trend escalation job finished",
		zap.Int("created", s.Created),
		zap.Int("skipped", s.Skipped),
		zap.Int("failed", s.Failed),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()),
	)
	return s
}

func (uc *TrendEscalationUseCase) process(ctx context.Context, c repository.TrendCandidate, threshold int, cooldownSince time.Time, s *TrendSummary) {
	if c.RiskyDays7d < threshold {
		s.Skipped++
		return
	}

	exists, err := uc.EscalationRepo.ExistsActiveOrRecent(ctx, c.PatientID, entity.EscalationTierTrendThisWeek, cooldownSince)
	if err != nil {
		uc.Log.Warn("trend dedup check failed", zap.String("patient_id", c.PatientID), zap.Error(err))
		s.Failed++
		return
	}
	if exists {
		s.Skipped++
		return
	}

	esc := &entity.Escalation{
		ID:              uuid.New().String(),
		PatientID:       c.PatientID,
		RiskScoreID:     c.RiskScoreID,
		FaskesID:        c.FaskesID,
		AssignedNakesID: c.AssignedNakesID,
		Tier:            entity.EscalationTierTrendThisWeek,
		Channel:         entity.EscalationChannelWhatsApp, // kolom vestigial; trend tidak kirim WA
		Status:          entity.EscalationStatusSent,
	}
	if err := uc.EscalationRepo.Create(ctx, esc); err != nil {
		uc.Log.Warn("failed to create trend escalation", zap.String("patient_id", c.PatientID), zap.Error(err))
		s.Failed++
		return
	}
	s.Created++

	// Inbox in-app pasien. Dashboard nakes = baris escalation itu sendiri; trend TIDAK kirim WA.
	uc.createInbox(ctx, c.PatientID)
}

func (uc *TrendEscalationUseCase) createInbox(ctx context.Context, patientID string) {
	if uc.InboxRepo == nil {
		return
	}
	notif := &entity.PatientNotification{
		ID:        uuid.New().String(),
		PatientID: patientID,
		Type:      entity.PatientNotifTypeEscalation,
		Title:     "Kondisi Anda perlu perhatian",
		Body:      "Tren kesehatanmu beberapa hari terakhir menurun. Tim kesehatanmu sudah diberi tahu. Mohon jaga pola hidup dan hubungi faskes bila perlu.",
		Payload:   mustMarshal(map[string]string{"tier": entity.EscalationTierTrendThisWeek}),
	}
	if err := uc.InboxRepo.Create(ctx, notif); err != nil {
		uc.Log.Warn("failed to create trend inbox row", zap.String("patient_id", patientID), zap.Error(err))
	}
}
