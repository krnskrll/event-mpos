package product

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/common"
)

// Handler exposes the admin product + per-event availability surface under /api/admin.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *gin.Context) {
	products, err := h.svc.List()
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
	resp, err := h.svc.Get(id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateProductRequest
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
	resp, err := h.svc.Create(req)
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
	var req UpdateProductRequest
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
	resp, err := h.svc.Update(id, req)
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
	resp, err := h.svc.EventProducts(eventID)
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
	var req SetAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	resp, err := h.svc.SetAvailability(eventID, productID, *req.Available)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func validateBarcodes(inputs []BarcodeInput) (string, bool) {
	for _, b := range inputs {
		if b.Barcode == "" {
			return "barcodes: barcode must not be blank", false
		}
	}
	return "", true
}
