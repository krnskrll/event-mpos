package erp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler exposes the admin ERP endpoints under /api/admin/erp.
type Handler struct {
	sync *SyncService
}

func NewHandler(sync *SyncService) *Handler { return &Handler{sync: sync} }

// SyncNow triggers an on-demand ERP sync.
func (h *Handler) SyncNow(c *gin.Context) {
	c.JSON(http.StatusOK, h.sync.Sync())
}

// SyncStatus returns the most recent sync outcome.
func (h *Handler) SyncStatus(c *gin.Context) {
	c.JSON(http.StatusOK, h.sync.Status())
}
