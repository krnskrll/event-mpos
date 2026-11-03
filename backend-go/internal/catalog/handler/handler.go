package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/catalog/service"
	"github.com/krnskrll/event-mpos/internal/platform/common"
)

type Handler struct {
	svc *service.Service
	// pull runs the Produk Toko sync for one store now.
	pull func(storeID uuid.UUID) error
}

func NewHandler(svc *service.Service, pull func(storeID uuid.UUID) error) *Handler {
	return &Handler{svc: svc, pull: pull}
}

func (h *Handler) SyncStatus(c *gin.Context) {
	st, err := h.svc.Status(authiface.StoreOf(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, st)
}

func (h *Handler) Sync(c *gin.Context) {
	if err := h.pull(authiface.StoreOf(c)); err != nil {
		common.Fail(c, err)
		return
	}
	h.SyncStatus(c)
}
