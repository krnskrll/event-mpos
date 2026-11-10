// Package iface is the staff module's contract: cashiers and supervisors who sign in on a paired
// device with their name and PIN.
package iface

import (
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/platform/database"
)

type Role string

const (
	RoleKasir      Role = "KASIR"
	RoleSupervisor Role = "SUPERVISOR"
)

// StaffRequest creates or updates a staff member. Pin is required on create; on update it resets
// the PIN. Active defaults to true on create and is left as is when omitted on update.
type StaffRequest struct {
	Name        string      `json:"name"`
	Role        Role        `json:"role"`
	Pin         *string     `json:"pin"`
	LocationIDs []uuid.UUID `json:"locationIds"`
	Active      *bool       `json:"active"`
}

// StaffResponse never carries the PIN.
type StaffResponse struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Role        Role        `json:"role"`
	LocationIDs []uuid.UUID `json:"locationIds"`
	Active      bool        `json:"active"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// DeviceStaff is what a cashier device keeps to check PINs offline.
type DeviceStaff struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Role    Role      `json:"role"`
	PinHash string    `json:"pinHash"`
}

// Person is a staff member as other modules see it.
type Person struct {
	ID          uuid.UUID
	StoreID     uuid.UUID
	Name        string
	Role        Role
	Active      bool
	LocationIDs []uuid.UUID
}

// WorksAt reports whether the person is active and assigned to the location.
func (p *Person) WorksAt(locationID uuid.UUID) bool {
	return p.Active && slices.Contains(p.LocationIDs, locationID)
}

type Staff interface {
	// Find returns nil when the staff member is not in the store.
	Find(storeID, id uuid.UUID) (*Person, error)
	ForLocation(locationID uuid.UUID) ([]DeviceStaff, error)
	// Approve checks a supervisor's PIN for an action at a location (403 when not allowed).
	Approve(storeID, locationID, supervisorID uuid.UUID, pin string) (*Person, error)
	NameOf(tx database.Tx, id uuid.UUID) (string, error)
}
