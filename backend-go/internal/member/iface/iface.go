package iface

import (
	"time"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/platform/database"
)

type MemberTier string

const (
	TierStandard MemberTier = "STANDARD"
	TierSilver   MemberTier = "SILVER"
	TierGold     MemberTier = "GOLD"
	TierPlatinum MemberTier = "PLATINUM"
)

// MemberResponse is the admin directory view. It never includes the phone number.
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

// MemberSnapshotDto is the compact device shape.
type MemberSnapshotDto struct {
	QRCode string     `json:"qrCode"`
	Name   string     `json:"name"`
	Tier   MemberTier `json:"tier"`
}

// MemberRecord is a member as delivered by the external member system.
type MemberRecord struct {
	MemberCode string
	QRCode     string
	Name       string
	Tier       MemberTier
	Region     *string
	Active     bool
	Phone      *string
}

type Members interface {
	ByQr(storeID uuid.UUID, qrCode string) (*MemberResponse, error)
	ActiveSnapshotMembers(storeID uuid.UUID) ([]any, error)
	Fingerprint(storeID uuid.UUID) (string, error)
	Upsert(tx database.Tx, storeID uuid.UUID, rec MemberRecord) (created bool, err error)
}
