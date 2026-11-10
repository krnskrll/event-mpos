package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/device/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
)

// Auth admits a paired device (Authorization: Device <token>). X-Staff-Id names the staff member
// signed in on it (the device checked the PIN); the request then acts as that person.
func Auth(devices iface.Devices, staff staffiface.Staff) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Device ")
		if !ok {
			common.AbortUnauthorized(c)
			return
		}
		d, err := devices.Authenticate(strings.TrimSpace(token))
		if err != nil {
			common.Fail(c, err)
			return
		}
		if d == nil {
			common.AbortUnauthorized(c)
			return
		}
		p := authiface.Principal{StoreID: d.StoreID, DeviceID: d.ID, LocationID: d.LocationID}
		if raw := c.GetHeader("X-Staff-Id"); raw != "" {
			id, err := uuid.Parse(raw)
			var person *staffiface.Person
			if err == nil {
				person, err = staff.Find(d.StoreID, id)
			}
			if err != nil || person == nil {
				common.AbortError(c, http.StatusForbidden, "Unknown staff member")
				return
			}
			p.UserID, p.FullName, p.StaffActive = person.ID, person.Name, person.WorksAt(d.LocationID)
		}
		authiface.SetPrincipal(c, p)
		c.Next()
	}
}

// RequireStaff guards routes that act for a person: the staff member must be active and assigned
// to the device's location.
func RequireStaff(c *gin.Context) {
	p, _ := authiface.PrincipalFrom(c)
	switch {
	case p.UserID == uuid.Nil:
		common.AbortError(c, http.StatusBadRequest, "X-Staff-Id: required")
	case !p.StaffActive:
		common.AbortError(c, http.StatusForbidden, "This staff member is inactive or not assigned to this location")
	default:
		c.Next()
	}
}

// Scope keeps a device on its own location: :eventId must be the device's location and :orderId
// an order sold there.
func Scope(orderLocation func(orderID uuid.UUID) (uuid.UUID, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, _ := authiface.PrincipalFrom(c)
		if raw := c.Param("eventId"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				common.AbortError(c, http.StatusBadRequest, "Invalid event id")
				return
			}
			if id != p.LocationID {
				common.AbortError(c, http.StatusForbidden, "This device belongs to another location")
				return
			}
		}
		if raw := c.Param("orderId"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				common.AbortError(c, http.StatusBadRequest, "Invalid order id")
				return
			}
			loc, err := orderLocation(id)
			if err == nil && loc != p.LocationID {
				err = common.NewAppError(http.StatusNotFound, "Order not found")
			}
			if err != nil {
				common.Fail(c, err)
				return
			}
		}
		c.Next()
	}
}
