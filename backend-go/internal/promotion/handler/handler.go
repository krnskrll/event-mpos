package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/promotion/iface"
	"github.com/krnskrll/event-mpos/internal/promotion/service"
)

// Handler exposes the promotion endpoints: per event under /api/admin, per store under /api/backoffice.
type Handler struct {
	svc    *service.Service
	events eventiface.Events
}

func NewHandler(svc *service.Service, events eventiface.Events) *Handler {
	return &Handler{svc: svc, events: events}
}

func (h *Handler) List(c *gin.Context) {
	eventID, ok := h.event(c)
	if !ok {
		return
	}
	reply(c, http.StatusOK)(h.svc.List(authiface.StoreOf(c), &eventID))
}

func (h *Handler) Create(c *gin.Context) {
	eventID, ok := h.event(c)
	if !ok {
		return
	}
	req, ok := bindPromotion(c)
	if !ok {
		return
	}
	req.LocationID = &eventID
	reply(c, http.StatusCreated)(h.svc.Create(authiface.StoreOf(c), req))
}

// Update keeps the promotion's location; the admin form does not send it.
func (h *Handler) Update(c *gin.Context) {
	id, req, ok := bindWithID(c)
	if !ok {
		return
	}
	cur, err := h.svc.Get(authiface.StoreOf(c), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	req.LocationID = cur.LocationID
	reply(c, http.StatusOK)(h.svc.Update(authiface.StoreOf(c), id, req))
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := promotionID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(authiface.StoreOf(c), id); err != nil {
		common.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// StoreList is the back-office list: ?locationId narrows it to what runs at one location.
func (h *Handler) StoreList(c *gin.Context) {
	var loc *uuid.UUID
	if raw := c.Query("locationId"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			common.AbortError(c, http.StatusBadRequest, "Invalid locationId")
			return
		}
		loc = &id
	}
	reply(c, http.StatusOK)(h.svc.List(authiface.StoreOf(c), loc))
}

func (h *Handler) StoreGet(c *gin.Context) {
	if id, ok := promotionID(c); ok {
		reply(c, http.StatusOK)(h.svc.Get(authiface.StoreOf(c), id))
	}
}

func (h *Handler) StoreCreate(c *gin.Context) {
	if req, ok := bindPromotion(c); ok {
		reply(c, http.StatusCreated)(h.svc.Create(authiface.StoreOf(c), req))
	}
}

func (h *Handler) StoreUpdate(c *gin.Context) {
	if id, req, ok := bindWithID(c); ok {
		reply(c, http.StatusOK)(h.svc.Update(authiface.StoreOf(c), id, req))
	}
}

func (h *Handler) event(c *gin.Context) (uuid.UUID, bool) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return eventID, false
	}
	if err := h.events.RequireInStore(authiface.StoreOf(c), eventID); err != nil {
		common.Fail(c, err)
		return eventID, false
	}
	return eventID, true
}

func promotionID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid promotion id")
	}
	return id, err == nil
}

func bindWithID(c *gin.Context) (uuid.UUID, iface.PromotionRequest, bool) {
	id, ok := promotionID(c)
	if !ok {
		return id, iface.PromotionRequest{}, false
	}
	req, ok := bindPromotion(c)
	return id, req, ok
}

func bindPromotion(c *gin.Context) (iface.PromotionRequest, bool) {
	var req iface.PromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return req, false
	}
	if _, ok := iface.ParseType(string(req.Type)); !ok {
		common.AbortError(c, http.StatusBadRequest, "type: must be one of PERCENTAGE, FIXED, BUNDLE")
		return req, false
	}
	return req, true
}

func reply(c *gin.Context, status int) func(any, error) {
	return func(v any, err error) {
		if err != nil {
			common.Fail(c, err)
			return
		}
		c.JSON(status, v)
	}
}
