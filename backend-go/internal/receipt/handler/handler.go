package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/receipt/iface"
	"github.com/krnskrll/event-mpos/internal/receipt/service"
)

// Handler serves Pengaturan Struk in the Bhoomy back office.
type Handler struct{ settings *service.Settings }

func NewHandler(settings *service.Settings) *Handler { return &Handler{settings: settings} }

func (h *Handler) Get(c *gin.Context) {
	v, err := h.settings.Get(authiface.StoreOf(c))
	respond(c, v, err)
}

func (h *Handler) Save(c *gin.Context) {
	var in iface.Settings
	if err := c.ShouldBindJSON(&in); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	v, err := h.settings.Save(authiface.StoreOf(c), in)
	respond(c, v, err)
}

func respond(c *gin.Context, v any, err error) {
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}
