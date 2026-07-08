package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/krnskrll/event-mpos/internal/common"
)

// Handler exposes the /api/auth/* endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	resp, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	resp, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Logout revokes all refresh tokens for the caller; the access token expires on its own shortly.
func (h *Handler) Logout(c *gin.Context) {
	p, ok := PrincipalFrom(c)
	if !ok {
		common.AbortUnauthorized(c)
		return
	}
	if err := h.svc.Logout(p.UserID); err != nil {
		common.Fail(c, err)
		return
	}
	c.Status(http.StatusOK)
}
