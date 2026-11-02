package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/event/service"
	"github.com/krnskrll/event-mpos/internal/platform/common"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(c *gin.Context) {
	events, err := h.svc.ListEvents(authiface.StoreOf(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *Handler) Create(c *gin.Context) {
	var req iface.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	resp, err := h.svc.CreateEvent(authiface.StoreOf(c), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	var req iface.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	resp, err := h.svc.UpdateEvent(authiface.StoreOf(c), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ChangeStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	var req iface.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	status, ok := iface.ParseStatus(string(req.Status))
	if !ok {
		common.AbortError(c, http.StatusBadRequest, "status: must be one of DRAFT, ACTIVE, CLOSED")
		return
	}
	resp, err := h.svc.ChangeStatus(authiface.StoreOf(c), id, status)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) AssignBa(c *gin.Context) {
	var req iface.AssignBaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	resp, err := h.svc.AssignBaToEvent(authiface.StoreOf(c), req.UserID, req.EventID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
