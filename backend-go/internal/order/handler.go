package order

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/auth"
	"github.com/krnskrll/event-mpos/internal/common"
)

// WarehouseHandler exposes the warehouse order board under /api/warehouse.
type WarehouseHandler struct {
	svc *Service
}

func NewWarehouseHandler(svc *Service) *WarehouseHandler { return &WarehouseHandler{svc: svc} }

func (h *WarehouseHandler) Orders(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	var status *OrderStatus
	if raw := c.Query("status"); raw != "" {
		st, ok := ParseStatus(raw)
		if !ok {
			common.AbortError(c, http.StatusBadRequest, "status: must be one of NEW, PACKING, READY, COLLECTED")
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
	var req AdvanceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if _, ok := ParseStatus(string(req.Status)); !ok {
		common.AbortError(c, http.StatusBadRequest, "status: must be one of NEW, PACKING, READY, COLLECTED")
		return
	}
	p, _ := auth.PrincipalFrom(c)
	view, err := h.svc.AdvanceStatus(orderID, req.Status, &p.UserID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}
