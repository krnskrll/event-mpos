package stock

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/event"
)

// AdminHandler: view + set per-event allocations (under /api/admin).
type AdminHandler struct {
	svc *Service
}

func NewAdminHandler(svc *Service) *AdminHandler { return &AdminHandler{svc: svc} }

func (h *AdminHandler) Allocations(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	rows, err := h.svc.AllocationsForEvent(eventID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *AdminHandler) SetAllocation(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid product id")
		return
	}
	var req SetAllocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if req.AllocatedQty < 0 {
		common.AbortError(c, http.StatusBadRequest, "allocatedQty: must be greater than or equal to 0")
		return
	}
	row, err := h.svc.SetAllocation(eventID, productID, req.AllocatedQty)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, row)
}

// WarehouseHandler: pick an event + near-real-time remaining stock (under /api/warehouse).
type WarehouseHandler struct {
	svc    *Service
	events *event.Service
}

func NewWarehouseHandler(svc *Service, events *event.Service) *WarehouseHandler {
	return &WarehouseHandler{svc: svc, events: events}
}

func (h *WarehouseHandler) Events(c *gin.Context) {
	events, err := h.events.ListEvents()
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *WarehouseHandler) Levels(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	rows, err := h.svc.LevelsForEvent(eventID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}
