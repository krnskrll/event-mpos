package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/staff/iface"
	"github.com/krnskrll/event-mpos/internal/staff/service"
)

// Handler serves Staf POS in the Bhoomy back office.
type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(c *gin.Context) {
	list, err := h.svc.List(authiface.StoreOf(c))
	respond(c, http.StatusOK, list, err)
}

func (h *Handler) Get(c *gin.Context) {
	if id, ok := staffID(c); ok {
		resp, err := h.svc.Get(authiface.StoreOf(c), id)
		respond(c, http.StatusOK, resp, err)
	}
}

func (h *Handler) Create(c *gin.Context) {
	if req, ok := bind(c); ok {
		resp, err := h.svc.Create(authiface.StoreOf(c), req)
		respond(c, http.StatusCreated, resp, err)
	}
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := staffID(c)
	if !ok {
		return
	}
	if req, ok := bind(c); ok {
		resp, err := h.svc.Update(authiface.StoreOf(c), id, req)
		respond(c, http.StatusOK, resp, err)
	}
}

func staffID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid staff id")
	}
	return id, err == nil
}

func bind(c *gin.Context) (iface.StaffRequest, bool) {
	var req iface.StaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return req, false
	}
	return req, true
}

func respond(c *gin.Context, status int, v any, err error) {
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(status, v)
}
