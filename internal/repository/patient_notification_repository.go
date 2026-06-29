package repository

import (
	"context"
	"sehatiku-notification-worker/internal/entity"

	"gorm.io/gorm"
)

// InboxStore membuat baris inbox in-app pasien (tabel patient_notifications).
type InboxStore interface {
	Create(ctx context.Context, n *entity.PatientNotification) error
}

type PatientNotificationRepository struct {
	DB *gorm.DB
}

func (r *PatientNotificationRepository) Create(ctx context.Context, n *entity.PatientNotification) error {
	return r.DB.WithContext(ctx).Create(n).Error
}
