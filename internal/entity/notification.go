package entity

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Notification struct {
	ID             string          `gorm:"column:id;primaryKey"`
	PatientID      *string         `gorm:"column:patient_id"`
	RecipientPhone string          `gorm:"column:recipient_phone"`
	RecipientRole  string          `gorm:"column:recipient_role"`
	MessageType    string          `gorm:"column:message_type"`
	Channel        string          `gorm:"column:channel"`
	Payload        json.RawMessage `gorm:"column:payload;type:jsonb"`
	Status         string          `gorm:"column:status"`
	ErrorReason    *string         `gorm:"column:error_reason"`
	RetryCount     int             `gorm:"column:retry_count"`
	QueuedAt       time.Time       `gorm:"column:queued_at;autoCreateTime"`
	SentAt         sql.NullTime    `gorm:"column:sent_at"`
}

func (Notification) TableName() string { return "notifications" }
