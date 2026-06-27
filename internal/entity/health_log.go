package entity

import "time"

type HealthLog struct {
	ID         string    `gorm:"column:id;primaryKey"`
	PatientID  string    `gorm:"column:patient_id"`
	MeasuredAt time.Time `gorm:"column:measured_at"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (HealthLog) TableName() string { return "health_logs" }
