package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/promotion/iface"
	"github.com/krnskrll/event-mpos/internal/promotion/service"
)

// Handler exposes the admin promotion endpoints under /api/admin.
type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
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
	if err := h.svc.Delete(id); err != nil {
		common.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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
