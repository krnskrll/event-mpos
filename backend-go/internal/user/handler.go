package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/common"
)

// Handler exposes the admin user-management endpoints under /api/admin/users.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List returns all users, optionally filtered by ?role=.
func (h *Handler) List(c *gin.Context) {
	if roleParam := c.Query("role"); roleParam != "" {
		role, ok := ParseRole(roleParam)
		if !ok {
			common.AbortError(c, http.StatusBadRequest, "role: must be one of ADMIN, BA, WAREHOUSE")
			return
		}
		users, err := h.svc.ListByRole(role)
		if err != nil {
			common.Fail(c, err)
			return
		}
		c.JSON(http.StatusOK, users)
		return
	}
	users, err := h.svc.List()
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if _, ok := ParseRole(string(req.Role)); !ok {
		common.AbortError(c, http.StatusBadRequest, "role: must be one of ADMIN, BA, WAREHOUSE")
		return
	}
	resp, err := h.svc.Create(req)
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
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if req.Role != nil {
		if _, ok := ParseRole(string(*req.Role)); !ok {
			common.AbortError(c, http.StatusBadRequest, "role: must be one of ADMIN, BA, WAREHOUSE")
			return
		}
	}
	resp, err := h.svc.Update(id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
