// Package iface is the device module's contract: cashier devices paired to one location with a
// pairing code, authenticated by an opaque token.
package iface

import (
	"time"

	"github.com/google/uuid"

	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/database"
)

// CodeTTL is how long a pairing code works; any number of devices may use it meanwhile.
const CodeTTL = 10 * time.Minute

type PairRequest struct {
	Code  string `json:"code" binding:"required"`
	Model string `json:"model"`
}

type DeviceRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	// RemovesAt is set once the device's event closed: it stops working then.
	RemovesAt *time.Time `json:"removesAt,omitempty"`
}

type Location struct {
	ID        uuid.UUID               `json:"id"`
	Name      string                  `json:"name"`
	Type      eventiface.LocationType `json:"type"`
	Status    eventiface.EventStatus  `json:"status"`
	StartDate *common.Date            `json:"startDate,omitempty"`
	EndDate   *common.Date            `json:"endDate,omitempty"`
}

// DeviceResponse is GET /api/device; PairResponse adds the token, shown only once.
type DeviceResponse struct {
	Device    DeviceRef `json:"device"`
	Location  Location  `json:"location"`
	StoreName string    `json:"storeName"`
}

type PairResponse struct {
	Token string `json:"token"`
	DeviceResponse
}

// DeviceView is a device in the back office.
type DeviceView struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Model      *string    `json:"model,omitempty"`
	PairedAt   time.Time  `json:"pairedAt"`
	LastSeenAt *time.Time `json:"lastSeenAt"`
	RemovesAt  *time.Time `json:"removesAt,omitempty"`
}

type PairingCode struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type RenameRequest struct {
	Name string `json:"name" binding:"required"`
}

// Device is an authenticated device.
type Device struct {
	ID         uuid.UUID
	StoreID    uuid.UUID
	LocationID uuid.UUID
	Name       string
}

type Devices interface {
	// Authenticate returns nil for an unknown or removed device.
	Authenticate(token string) (*Device, error)
	// ScheduleRemoval makes a location's devices stop working at the given time.
	ScheduleRemoval(tx database.Tx, locationID uuid.UUID, at time.Time) error
}
