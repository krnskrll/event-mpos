package sync

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler exposes the admin member-sync endpoints under /api/admin/members.
type Handler struct {
	sync *Service
}

func NewHandler(sync *Service) *Handler { return &Handler{sync: sync} }

// SyncNow triggers an on-demand member sync.
func (h *Handler) SyncNow(c *gin.Context) {
	c.JSON(http.StatusOK, h.sync.Sync())
}

// Status returns the most recent member-sync outcome.
func (h *Handler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, h.sync.Status())
}
