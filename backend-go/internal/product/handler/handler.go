package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/product/iface"
	"github.com/krnskrll/event-mpos/internal/product/service"
)

type Handler struct {
	svc    *service.Service
	events eventiface.Events
}

func NewHandler(svc *service.Service, events eventiface.Events) *Handler {
	return &Handler{svc: svc, events: events}
}

func (h *Handler) List(c *gin.Context) {
	products, err := h.svc.List(authiface.StoreOf(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid product id")
		return
	}
	resp, err := h.svc.Get(authiface.StoreOf(c), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Create(c *gin.Context) {
	var req iface.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if req.Price.IsNegative() {
		common.AbortError(c, http.StatusBadRequest, "price: must be greater than or equal to 0")
		return
	}
	if msg, ok := validateBarcodes(req.Barcodes); !ok {
		common.AbortError(c, http.StatusBadRequest, msg)
		return
	}
	resp, err := h.svc.Create(authiface.StoreOf(c), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid product id")
		return
	}
	var req iface.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if req.Price != nil && req.Price.IsNegative() {
		common.AbortError(c, http.StatusBadRequest, "price: must be greater than or equal to 0")
		return
	}
	if req.Barcodes != nil {
		if msg, ok := validateBarcodes(*req.Barcodes); !ok {
			common.AbortError(c, http.StatusBadRequest, msg)
			return
		}
	}
	resp, err := h.svc.Update(authiface.StoreOf(c), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) EventProducts(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	if err := h.events.RequireInStore(authiface.StoreOf(c), eventID); err != nil {
		common.Fail(c, err)
		return
	}
	resp, err := h.svc.EventProducts(authiface.StoreOf(c), eventID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) SetAvailability(c *gin.Context) {
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
	var req iface.SetAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if err := h.events.RequireInStore(authiface.StoreOf(c), eventID); err != nil {
		common.Fail(c, err)
		return
	}
	resp, err := h.svc.SetAvailability(authiface.StoreOf(c), eventID, productID, *req.Available)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func validateBarcodes(inputs []iface.BarcodeInput) (string, bool) {
	for _, b := range inputs {
		if b.Barcode == "" {
			return "barcodes: barcode must not be blank", false
		}
	}
	return "", true
}
