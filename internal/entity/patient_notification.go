package entity

import (
	"encoding/json"
	"time"
)

// PatientNotification memetakan tabel `patient_notifications` — inbox in-app yang dibaca
// Patient App. Worker hanya menulis baris bertipe daily_reminder; backend menulis tipe lain.
// Bukan baris transport: tidak ada status/retry/provider id di sini.
type PatientNotification struct {
	ID        string          `gorm:"column:id;primaryKey"`
	PatientID string          `gorm:"column:patient_id"`
	Type      string          `gorm:"column:type"`
	Title     string          `gorm:"column:title"`
	Body      string          `gorm:"column:body"`
	Payload   json.RawMessage `gorm:"column:payload;type:jsonb"`
	CreatedAt time.Time       `gorm:"column:created_at;autoCreateTime"`
}

func (PatientNotification) TableName() string { return "patient_notifications" }

const PatientNotifTypeDailyReminder = "daily_reminder"
