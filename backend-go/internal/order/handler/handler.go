package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/order/service"
	"github.com/krnskrll/event-mpos/internal/platform/common"
)

// WarehouseHandler exposes the warehouse order board under /api/warehouse.
type WarehouseHandler struct {
	svc *service.Service
}

func NewWarehouseHandler(svc *service.Service) *WarehouseHandler { return &WarehouseHandler{svc: svc} }

func (h *WarehouseHandler) Orders(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	var status *iface.OrderStatus
	if raw := c.Query("status"); raw != "" {
		st, ok := iface.ParseStatus(raw)
		if !ok {
			common.AbortError(c, http.StatusBadRequest, "status: must be one of NEW, PACKING, READY, COLLECTED, CANCELLED")
			return
		}
		status = &st
	}
	views, err := h.svc.WarehouseOrders(eventID, status)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, views)
}

func (h *WarehouseHandler) Advance(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid order id")
		return
	}
	var req iface.AdvanceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if _, ok := iface.ParseStatus(string(req.Status)); !ok {
		common.AbortError(c, http.StatusBadRequest, "status: must be one of NEW, PACKING, READY, COLLECTED")
		return
	}
	p, _ := authiface.PrincipalFrom(c)
	view, err := h.svc.AdvanceStatus(orderID, req.Status, &p.UserID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

// Cancel cancels an order from the board (warehouse/admin) or the BA's history; the service applies
// the per-role rules. Body: optional {"reason": "..."}.
func (h *WarehouseHandler) Cancel(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid order id")
		return
	}
	var req iface.CancelOrderRequest
	_ = c.ShouldBindJSON(&req) // body is optional
	p, _ := authiface.PrincipalFrom(c)
	view, err := h.svc.Cancel(orderID, p.UserID, p.Role, req.Reason)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}
