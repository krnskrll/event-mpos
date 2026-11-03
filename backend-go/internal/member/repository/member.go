package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/member/iface"
)

// Member is a customer synced from the member system. Phone is back-office only.
type Member struct {
	ID         uuid.UUID        `gorm:"type:uuid;primaryKey"`
	StoreID    uuid.UUID        `gorm:"column:store_id;type:uuid;not null"`
	MemberCode string           `gorm:"column:member_code;not null"`
	QRCode     string           `gorm:"column:qr_code;not null"`
	Name       string           `gorm:"not null"`
	Tier       iface.MemberTier `gorm:"type:varchar(20);not null"`
	Region     *string          ``
	Phone      *string          ``
	Active     bool             `gorm:"not null"`
	UpdatedAt  time.Time        `gorm:"column:updated_at;not null"`
}

func (Member) TableName() string { return "members" }

func (m *Member) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (m *Member) ToResponse() iface.MemberResponse {
	return iface.MemberResponse{
		ID: m.ID, MemberCode: m.MemberCode, QRCode: m.QRCode, Name: m.Name,
		Tier: m.Tier, Region: m.Region, Active: m.Active, UpdatedAt: m.UpdatedAt,
	}
}

func (m *Member) ToSnapshotDto() iface.MemberSnapshotDto {
	return iface.MemberSnapshotDto{QRCode: m.QRCode, Name: m.Name, Tier: m.Tier}
}
