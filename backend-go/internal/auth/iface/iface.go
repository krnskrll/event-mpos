package iface

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Principal is the caller attached to the request context: the Bhoomy backend acting for a store,
// or a paired device (DeviceID, LocationID) with the staff member working on it as UserID/FullName.
type Principal struct {
	UserID   uuid.UUID
	StoreID  uuid.UUID
	FullName string

	DeviceID   uuid.UUID
	LocationID uuid.UUID
	// StaffActive: the staff member is active and assigned to the device's location.
	StaffActive bool
}

const principalKey = "authPrincipal"

func SetPrincipal(c *gin.Context, p Principal) { c.Set(principalKey, p) }

func PrincipalFrom(c *gin.Context) (Principal, bool) {
	v, exists := c.Get(principalKey)
	if !exists {
		return Principal{}, false
	}
	p, ok := v.(Principal)
	return p, ok
}

// StoreOf is the caller's store; device and back-office routes always have one.
func StoreOf(c *gin.Context) uuid.UUID {
	p, _ := PrincipalFrom(c)
	return p.StoreID
}
