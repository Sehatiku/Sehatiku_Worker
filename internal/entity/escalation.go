package entity

import "time"

const (
	EscalationTierTrendThisWeek = "trend_this_week"
	EscalationStatusSent        = "sent"
	EscalationStatusViewed      = "viewed"
	EscalationChannelWhatsApp   = "whatsapp"
)

// Escalation memetakan tabel `escalations` (dibagi dengan main backend). Worker hanya
// MEMBUAT baris tier trend_this_week + cek dedup; lifecycle/feedback ditangani backend.
type Escalation struct {
	ID              string    `gorm:"column:id;primaryKey"`
	PatientID       string    `gorm:"column:patient_id"`
	RiskScoreID     string    `gorm:"column:risk_score_id"`
	FaskesID        string    `gorm:"column:faskes_id"`
	AssignedNakesID string    `gorm:"column:assigned_nakes_id"`
	Tier            string    `gorm:"column:tier"`
	Channel         string    `gorm:"column:channel"`
	Status          string    `gorm:"column:status"`
	SentAt          time.Time `gorm:"column:sent_at;autoCreateTime"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Escalation) TableName() string { return "escalations" }
