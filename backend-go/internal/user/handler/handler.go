package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/user/iface"
	"github.com/krnskrll/event-mpos/internal/user/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// List returns all users, optionally filtered by ?role=.
func (h *Handler) List(c *gin.Context) {
	if roleParam := c.Query("role"); roleParam != "" {
		role, ok := iface.ParseRole(roleParam)
		if !ok {
			common.AbortError(c, http.StatusBadRequest, "role: must be one of ADMIN, BA, WAREHOUSE")
			return
		}
		users, err := h.svc.ListByRole(authiface.StoreOf(c), role)
		if err != nil {
			common.Fail(c, err)
			return
		}
		c.JSON(http.StatusOK, users)
		return
	}
	users, err := h.svc.List(authiface.StoreOf(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) Create(c *gin.Context) {
	var req iface.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if _, ok := iface.ParseRole(string(req.Role)); !ok {
		common.AbortError(c, http.StatusBadRequest, "role: must be one of ADMIN, BA, WAREHOUSE")
		return
	}
	resp, err := h.svc.Create(authiface.StoreOf(c), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid user id")
		return
	}
	var req iface.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if req.Role != nil {
		if _, ok := iface.ParseRole(string(*req.Role)); !ok {
			common.AbortError(c, http.StatusBadRequest, "role: must be one of ADMIN, BA, WAREHOUSE")
			return
		}
	}
	resp, err := h.svc.Update(authiface.StoreOf(c), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
