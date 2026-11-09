package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/device/iface"
	"github.com/krnskrll/event-mpos/internal/device/service"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
)

// Handler serves the cashier app's device routes and Perangkat in the back office.
type Handler struct {
	svc   *service.Service
	staff staffiface.Staff
}

func NewHandler(svc *service.Service, staff staffiface.Staff) *Handler {
	return &Handler{svc: svc, staff: staff}
}

func (h *Handler) Pair(c *gin.Context) {
	var req iface.PairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	resp, err := h.svc.Pair(req.Code, req.Model)
	respond(c, http.StatusCreated, resp, err)
}

func (h *Handler) Me(c *gin.Context) {
	p, _ := authiface.PrincipalFrom(c)
	resp, err := h.svc.Info(p.DeviceID)
	respond(c, http.StatusOK, resp, err)
}

// Staff lists the location's active staff with their PIN hashes, for signing in offline.
func (h *Handler) Staff(c *gin.Context) {
	p, _ := authiface.PrincipalFrom(c)
	list, err := h.staff.ForLocation(p.LocationID)
	respond(c, http.StatusOK, list, err)
}

func (h *Handler) List(c *gin.Context) {
	if id, ok := param(c, "id", "Invalid location id"); ok {
		list, err := h.svc.List(authiface.StoreOf(c), id)
		respond(c, http.StatusOK, list, err)
	}
}

func (h *Handler) NewCode(c *gin.Context) {
	if id, ok := param(c, "id", "Invalid location id"); ok {
		code, err := h.svc.NewCode(authiface.StoreOf(c), id)
		respond(c, http.StatusCreated, code, err)
	}
}

func (h *Handler) Rename(c *gin.Context) {
	id, ok := param(c, "id", "Invalid device id")
	if !ok {
		return
	}
	var req iface.RenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	resp, err := h.svc.Rename(authiface.StoreOf(c), id, req.Name)
	respond(c, http.StatusOK, resp, err)
}

func (h *Handler) Remove(c *gin.Context) {
	if id, ok := param(c, "id", "Invalid device id"); ok {
		if err := h.svc.Remove(authiface.StoreOf(c), id); err != nil {
			common.Fail(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func param(c *gin.Context, name, msg string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, msg)
	}
	return id, err == nil
}

func respond(c *gin.Context, status int, v any, err error) {
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(status, v)
}
