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

// Handler exposes the admin promotion endpoints under /api/admin.
type Handler struct {
	svc    *service.Service
	events eventiface.Events
}

func NewHandler(svc *service.Service, events eventiface.Events) *Handler {
	return &Handler{svc: svc, events: events}
}

func (h *Handler) List(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	if err := h.events.RequireInStore(authiface.StoreOf(c), eventID); err != nil {
		common.Fail(c, err)
		return
	}
	promos, err := h.svc.ListForEvent(eventID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, promos)
}

func (h *Handler) Create(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	req, ok := bindPromotion(c)
	if !ok {
		return
	}
	if err := h.events.RequireInStore(authiface.StoreOf(c), eventID); err != nil {
		common.Fail(c, err)
		return
	}
	resp, err := h.svc.Create(eventID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid promotion id")
		return
	}
	req, ok := bindPromotion(c)
	if !ok {
		return
	}
	if !h.promotionInStore(c, id) {
		return
	}
	resp, err := h.svc.Update(id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid promotion id")
		return
	}
	if !h.promotionInStore(c, id) {
		return
	}
	if err := h.svc.Delete(id); err != nil {
		common.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// promotionInStore answers 404 when the promotion's event is in another store.
func (h *Handler) promotionInStore(c *gin.Context, id uuid.UUID) bool {
	eventID, err := h.svc.EventOf(id)
	if err == nil {
		err = h.events.RequireInStore(authiface.StoreOf(c), eventID)
	}
	if err != nil {
		common.Fail(c, err)
		return false
	}
	return true
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
