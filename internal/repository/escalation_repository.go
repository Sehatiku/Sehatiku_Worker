package repository

import (
	"context"
	"sehatiku-notification-worker/internal/entity"
	"time"

	"gorm.io/gorm"
)

// EscalationStore membuat eskalasi & mengecek dedup+cooldown (eskalasi terbuka/baru per
// pasien+tier).
type EscalationStore interface {
	ExistsActiveOrRecent(ctx context.Context, patientID, tier string, since time.Time) (bool, error)
	Create(ctx context.Context, e *entity.Escalation) error
}

type EscalationRepository struct {
	DB *gorm.DB
}

// ExistsActiveOrRecent = true bila ada eskalasi pasien+tier yang masih terbuka
// (status sent/viewed) ATAU baru dibuat (sent_at >= since). Mencegah duplikasi saat alert
// masih terbuka DAN re-nag harian setelah nakes menutup alert (cooldown).
func (r *EscalationRepository) ExistsActiveOrRecent(ctx context.Context, patientID, tier string, since time.Time) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&entity.Escalation{}).
		Where("patient_id = ? AND tier = ? AND (status IN ('sent','viewed') OR sent_at >= ?)", patientID, tier, since).
		Count(&count).Error
	return count > 0, err
}

func (r *EscalationRepository) Create(ctx context.Context, e *entity.Escalation) error {
	return r.DB.WithContext(ctx).Create(e).Error
}
