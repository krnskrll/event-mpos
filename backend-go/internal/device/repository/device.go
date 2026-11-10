package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Device struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	StoreID    uuid.UUID  `gorm:"column:store_id;type:uuid;not null"`
	LocationID uuid.UUID  `gorm:"column:location_id;type:uuid;not null"`
	Name       string     `gorm:"not null"`
	Model      *string    `gorm:"column:model"`
	TokenHash  string     `gorm:"column:token_hash;not null"`
	PairedAt   time.Time  `gorm:"column:paired_at;not null"`
	LastSeenAt *time.Time `gorm:"column:last_seen_at"`
	RemovedAt  *time.Time `gorm:"column:removed_at"`
}

func (Device) TableName() string { return "pos_devices" }

func (d *Device) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type PairingCode struct {
	Code       string    `gorm:"primaryKey"`
	StoreID    uuid.UUID `gorm:"column:store_id;type:uuid;not null"`
	LocationID uuid.UUID `gorm:"column:location_id;type:uuid;not null"`
	ExpiresAt  time.Time `gorm:"column:expires_at;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
}

func (PairingCode) TableName() string { return "pos_pairing_codes" }
