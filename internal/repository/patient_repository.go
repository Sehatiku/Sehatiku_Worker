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

// FindPatientsWithNoLogToday returns all active patients with no health_log for "today".
// "Today" is computed in WIB (Asia/Jakarta), matching the main backend's date logic
// (record_usecase / patient_dashboard_repository use AT TIME ZONE 'Asia/Jakarta'). Using
// CURRENT_DATE (server/UTC) here previously caused near-midnight disagreement: the worker
// could nag a patient who already logged today (WIB) or skip one who had not.
func (r *PatientRepository) FindPatientsWithNoLogToday(ctx context.Context) ([]entity.Patient, error) {
	var patients []entity.Patient
	err := r.DB.WithContext(ctx).
		Where("status = 'active'").
		Where(`NOT EXISTS (
			SELECT 1 FROM health_logs hl
			WHERE hl.patient_id = patients.id
			  AND (hl.measured_at AT TIME ZONE 'Asia/Jakarta')::date
			      = (now() AT TIME ZONE 'Asia/Jakarta')::date
		)`).
		Find(&patients).Error
	return patients, err
}
