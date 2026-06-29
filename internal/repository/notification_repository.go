package repository

import (
	"context"
	"sehatiku-notification-worker/internal/entity"
	"time"

	"gorm.io/gorm"
)

type NotificationStore interface {
	ExistsForToday(ctx context.Context, patientID string) (bool, error)
	Create(ctx context.Context, n *entity.Notification) error
	UpdateStatus(ctx context.Context, id, status string, errReason *string) error
}

type NotificationRepository struct {
	DB *gorm.DB
}

// ExistsForToday returns true if a daily_prompt notification was already queued or sent for
// this patient "today". "Today" is computed in WIB (Asia/Jakarta) to stay consistent with
// the no-log check and the main backend's date logic — never the server/UTC date.
func (r *NotificationRepository) ExistsForToday(ctx context.Context, patientID string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&entity.Notification{}).
		Where(`patient_id = ?
			AND message_type = 'daily_prompt'
			AND (queued_at AT TIME ZONE 'Asia/Jakarta')::date
			    = (now() AT TIME ZONE 'Asia/Jakarta')::date
			AND status IN ('queued', 'sent')`, patientID).
		Count(&count).Error
	return count > 0, err
}

func (r *NotificationRepository) Create(ctx context.Context, n *entity.Notification) error {
	return r.DB.WithContext(ctx).Create(n).Error
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id, status string, errReason *string) error {
	updates := map[string]interface{}{"status": status}
	if status == "sent" {
		updates["sent_at"] = time.Now()
	}
	if errReason != nil {
		updates["error_reason"] = errReason
	}
	return r.DB.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("id = ?", id).
		Updates(updates).Error
}
