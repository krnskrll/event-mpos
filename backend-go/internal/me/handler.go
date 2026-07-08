// Package me exposes the BA-facing /api/me surface: identity, assigned event(s), and the event
// snapshot. Extended by later phases (products/members/stock/promotions/orders/receipts).
package me

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/auth"
	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/product"
	"github.com/krnskrll/event-mpos/internal/user"
)

// MeResponse is the "who am I" payload.
type MeResponse struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	FullName string    `json:"fullName"`
	Role     user.Role `json:"role"`
}

// SnapshotHashResponse is the lightweight snapshot freshness check.
type SnapshotHashResponse struct {
	SnapshotHash string    `json:"snapshotHash"`
	CheckedAt    time.Time `json:"checkedAt"`
}

type Handler struct {
	events   *event.Service
	products *product.Service
}

func NewHandler(events *event.Service, products *product.Service) *Handler {
	return &Handler{events: events, products: products}
}

func (h *Handler) Me(c *gin.Context) {
	p, _ := auth.PrincipalFrom(c)
	c.JSON(http.StatusOK, MeResponse{ID: p.UserID, Email: p.Email, FullName: p.FullName, Role: p.Role})
}

func (h *Handler) MyEvent(c *gin.Context) {
	p, ok := h.requireBa(c)
	if !ok {
		return
	}
	ev, err := h.events.ActiveEventForUser(p.UserID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	if ev == nil {
		common.AbortError(c, http.StatusNotFound, "No event is assigned to your account yet")
		return
	}
	c.JSON(http.StatusOK, ev)
}

func (h *Handler) MyEvents(c *gin.Context) {
	p, ok := h.requireBa(c)
	if !ok {
		return
	}
	events, err := h.events.AssignedEventsForUser(p.UserID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *Handler) Snapshot(c *gin.Context) {
	p, ok := h.requireBa(c)
	if !ok {
		return
	}
	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	if err := h.events.RequireAssigned(p.UserID, eventID); err != nil {
		common.Fail(c, err)
		return
	}
	if version := c.Query("version"); version != "" {
		current, err := h.events.SnapshotVersion(eventID)
		if err != nil {
			common.Fail(c, err)
			return
		}
		if version == current {
			c.JSON(http.StatusOK, event.SnapshotStatus{Changed: false, Version: version})
			return
		}
	}
	snap, err := h.events.SnapshotForEvent(eventID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (h *Handler) SnapshotHash(c *gin.Context) {
	p, ok := h.requireBa(c)
	if !ok {
		return
	}
	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	if err := h.events.RequireAssigned(p.UserID, eventID); err != nil {
		common.Fail(c, err)
		return
	}
	version, err := h.events.SnapshotVersion(eventID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, SnapshotHashResponse{SnapshotHash: version, CheckedAt: time.Now().UTC()})
}

// ProductByBarcode resolves a scanned barcode to a product available for this event.
func (h *Handler) ProductByBarcode(c *gin.Context) {
	p, ok := h.requireBa(c)
	if !ok {
		return
	}
	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	if err := h.events.RequireAssigned(p.UserID, eventID); err != nil {
		common.Fail(c, err)
		return
	}
	resp, err := h.products.FindAvailableByBarcode(eventID, c.Param("barcode"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// requireBa enforces that the caller is a BA (matches Kotlin's 400 for non-BA callers).
func (h *Handler) requireBa(c *gin.Context) (auth.Principal, bool) {
	p, _ := auth.PrincipalFrom(c)
	if p.Role != user.RoleBA {
		common.AbortError(c, http.StatusBadRequest, "Only BA accounts are scoped to an event")
		return p, false
	}
	return p, true
}
