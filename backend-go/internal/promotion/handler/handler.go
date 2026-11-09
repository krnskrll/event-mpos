package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/promotion/iface"
	"github.com/krnskrll/event-mpos/internal/promotion/service"
)

// Handler exposes the store's promotions under /api/backoffice.
type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

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
