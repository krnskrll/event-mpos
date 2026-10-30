package member

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MemberTier — loyalty tier.
type MemberTier string

const (
	TierStandard MemberTier = "STANDARD"
	TierSilver   MemberTier = "SILVER"
	TierGold     MemberTier = "GOLD"
	TierPlatinum MemberTier = "PLATINUM"
)

// Member is a customer synced from the external member system. Deliberately minimal (no financial
// data). phone is stored for back-office use and never exposed to clients.
type Member struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	MemberCode string     `gorm:"column:member_code;uniqueIndex;not null"`
	QRCode     string     `gorm:"column:qr_code;uniqueIndex;not null"`
	Name       string     `gorm:"not null"`
	Tier       MemberTier `gorm:"type:varchar(20);not null"`
	Region     *string    ``
	Phone      *string    ``
	Active     bool       `gorm:"not null"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;not null"`
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

// ---- DTOs ----

// MemberResponse is the full member view (admin directory). Note: no phone.
type MemberResponse struct {
	ID         uuid.UUID  `json:"id"`
	MemberCode string     `json:"memberCode"`
	QRCode     string     `json:"qrCode"`
	Name       string     `json:"name"`
	Tier       MemberTier `json:"tier"`
	Region     *string    `json:"region,omitempty"`
	Active     bool       `json:"active"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// MemberSnapshotDto is the compact device/snapshot shape — display + linking only.
type MemberSnapshotDto struct {
	QRCode string     `json:"qrCode"`
	Name   string     `json:"name"`
	Tier   MemberTier `json:"tier"`
}

func (m *Member) ToResponse() MemberResponse {
	return MemberResponse{
		ID:         m.ID,
		MemberCode: m.MemberCode,
		QRCode:     m.QRCode,
		Name:       m.Name,
		Tier:       m.Tier,
		Region:     m.Region,
		Active:     m.Active,
		UpdatedAt:  m.UpdatedAt,
	}
}

func (m *Member) ToSnapshotDto() MemberSnapshotDto {
	return MemberSnapshotDto{QRCode: m.QRCode, Name: m.Name, Tier: m.Tier}
}
