package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/order/service"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
)

// WarehouseHandler serves the pickup board (/api/warehouse) and cancelling from the cashier's
// history. Routes sit behind device auth scoped to the device's location.
type WarehouseHandler struct {
	svc   *service.Service
	staff staffiface.Staff
}

func NewWarehouseHandler(svc *service.Service, staff staffiface.Staff) *WarehouseHandler {
	return &WarehouseHandler{svc: svc, staff: staff}
}

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

// CashierCancel cancels a NEW order from the cashier's history; BoardCancel also a PACKING one.
func (h *WarehouseHandler) CashierCancel(c *gin.Context) { h.cancel(c, false) }

func (h *WarehouseHandler) BoardCancel(c *gin.Context) { h.cancel(c, true) }

// cancel needs a supervisor of the location to approve with their PIN; the audit names both.
func (h *WarehouseHandler) cancel(c *gin.Context, packing bool) {
	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid order id")
		return
	}
	var req iface.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if req.SupervisorID == uuid.Nil || req.SupervisorPin == "" {
		common.AbortError(c, http.StatusBadRequest, "supervisorId, supervisorPin: required")
		return
	}
	p, _ := authiface.PrincipalFrom(c)
	supervisor, err := h.staff.Approve(p.StoreID, p.LocationID, req.SupervisorID, req.SupervisorPin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	view, err := h.svc.Cancel(orderID, p.UserID, supervisor.Name, req.Reason, packing)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}
