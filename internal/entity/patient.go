package entity

import "time"

type Patient struct {
	ID             string    `gorm:"column:id;primaryKey"`
	FaskesID       string    `gorm:"column:faskes_id"`
	FullName       string    `gorm:"column:full_name"`
	PhoneNumber    string    `gorm:"column:phone_number"`
	CompanionPhone *string   `gorm:"column:companion_phone"`
	CompanionName  *string   `gorm:"column:companion_name"`
	Status         string    `gorm:"column:status"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Patient) TableName() string { return "patients" }
