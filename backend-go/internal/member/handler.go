package member

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/krnskrll/event-mpos/internal/common"
)

// Handler exposes the admin member directory under /api/admin/members.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(c *gin.Context) {
	members, err := h.svc.List()
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, members)
}
