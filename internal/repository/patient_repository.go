package repository

import (
	"context"
	"sehatiku-notification-worker/internal/entity"

	"gorm.io/gorm"
)

type PatientQuerier interface {
	FindPatientsWithNoLogToday(ctx context.Context) ([]entity.Patient, error)
}

type PatientRepository struct {
	DB *gorm.DB
}

// FindPatientsWithNoLogToday returns all active patients with no health_log
// where measured_at::date = CURRENT_DATE (Postgres UTC date).
func (r *PatientRepository) FindPatientsWithNoLogToday(ctx context.Context) ([]entity.Patient, error) {
	var patients []entity.Patient
	err := r.DB.WithContext(ctx).
		Where("status = 'active'").
		Where(`NOT EXISTS (
			SELECT 1 FROM health_logs hl
			WHERE hl.patient_id = patients.id
			  AND hl.measured_at::date = CURRENT_DATE
		)`).
		Find(&patients).Error
	return patients, err
}
