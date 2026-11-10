package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/staff/iface"
)

type Staff struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	StoreID   uuid.UUID  `gorm:"column:store_id;type:uuid;not null"`
	Name      string     `gorm:"not null"`
	Role      iface.Role `gorm:"type:varchar(10);not null"`
	PinHash   string     `gorm:"column:pin_hash;not null"`
	Active    bool       `gorm:"not null"`
	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null"`
}

func (Staff) TableName() string { return "pos_staff" }

func (s *Staff) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type StaffLocation struct {
	StaffID    uuid.UUID `gorm:"column:staff_id;type:uuid;primaryKey"`
	LocationID uuid.UUID `gorm:"column:location_id;type:uuid;primaryKey"`
}

func (StaffLocation) TableName() string { return "pos_staff_locations" }
