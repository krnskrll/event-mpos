// Package handler serves the cashier app's /api/me surface on a paired device: its location, the
// snapshot, stock, quotes, orders and receipts. Device auth puts the device, its location and the
// signed-in staff member on the request, and has already checked :eventId and :orderId against the
// device's location (so they parse).
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	memberiface "github.com/krnskrll/event-mpos/internal/member/iface"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	promotioniface "github.com/krnskrll/event-mpos/internal/promotion/iface"
	receiptiface "github.com/krnskrll/event-mpos/internal/receipt/iface"
	stockiface "github.com/krnskrll/event-mpos/internal/stock/iface"
)

// SnapshotHashResponse is the lightweight snapshot freshness check.
type SnapshotHashResponse struct {
	SnapshotHash string    `json:"snapshotHash"`
	CheckedAt    time.Time `json:"checkedAt"`
}

type Handler struct {
	events     eventiface.Events
	products   productiface.Products
	members    memberiface.Members
	stock      stockiface.Stock
	promotions promotioniface.Promotions
	orders     orderiface.Orders
	receipts   receiptiface.Receipts
}

func NewHandler(events eventiface.Events, products productiface.Products, members memberiface.Members, stk stockiface.Stock, promotions promotioniface.Promotions, orders orderiface.Orders, receipts receiptiface.Receipts) *Handler {
	return &Handler{events: events, products: products, members: members, stock: stk, promotions: promotions, orders: orders, receipts: receipts}
}

// MyEvent is the device's location.
func (h *Handler) MyEvent(c *gin.Context) {
	p, _ := authiface.PrincipalFrom(c)
	ev, err := h.events.Get(p.LocationID)
	respond(c, http.StatusOK, ev, err)
}

// MyEvents lists the device's location, for apps that still pick from a list.
func (h *Handler) MyEvents(c *gin.Context) {
	p, _ := authiface.PrincipalFrom(c)
	ev, err := h.events.Get(p.LocationID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, []eventiface.EventResponse{*ev})
}

func (h *Handler) Snapshot(c *gin.Context) {
	eventID := uuid.MustParse(c.Param("eventId"))
	if version := c.Query("version"); version != "" {
		current, err := h.events.SnapshotVersion(eventID)
		if err != nil {
			common.Fail(c, err)
			return
		}
		if version == current {
			c.JSON(http.StatusOK, eventiface.SnapshotStatus{Changed: false, Version: version})
			return
		}
	}
	snap, err := h.events.SnapshotForEvent(eventID)
	respond(c, http.StatusOK, snap, err)
}

func (h *Handler) SnapshotHash(c *gin.Context) {
	version, err := h.events.SnapshotVersion(uuid.MustParse(c.Param("eventId")))
	respond(c, http.StatusOK, SnapshotHashResponse{SnapshotHash: version, CheckedAt: time.Now().UTC()}, err)
}

// ProductByBarcode resolves a scanned barcode to a product sold at the location.
func (h *Handler) ProductByBarcode(c *gin.Context) {
	resp, err := h.products.FindAvailableByBarcode(authiface.StoreOf(c), uuid.MustParse(c.Param("eventId")), c.Param("barcode"))
	respond(c, http.StatusOK, resp, err)
}

// SubmitOrder creates an order atomically (stock + promotion + daily number). Idempotent per key.
// An order the device already confirmed offline is accepted even if its cashier has since been
// deactivated or moved, so a printed sale is never lost.
func (h *Handler) SubmitOrder(c *gin.Context) {
	var req orderiface.SubmitOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	if len(req.Lines) == 0 {
		common.AbortError(c, http.StatusBadRequest, "lines: must not be empty")
		return
	}
	p, _ := authiface.PrincipalFrom(c)
	if p.UserID == uuid.Nil {
		common.AbortError(c, http.StatusBadRequest, "X-Staff-Id: required")
		return
	}
	if !p.StaffActive && req.OrderNumber == nil {
		common.AbortError(c, http.StatusForbidden, "This staff member is inactive or not assigned to this location")
		return
	}
	device := p.DeviceID.String()
	req.DeviceID = &device
	conf, err := h.orders.Submit(uuid.MustParse(c.Param("eventId")), p.UserID, req)
	respond(c, http.StatusOK, conf, err)
}

// OrderHistory returns the orders this device sold at the location.
func (h *Handler) OrderHistory(c *gin.Context) {
	p, _ := authiface.PrincipalFrom(c)
	summaries, err := h.orders.History(uuid.MustParse(c.Param("eventId")), p.DeviceID.String())
	respond(c, http.StatusOK, summaries, err)
}

// Order returns full order detail / reprint payload.
func (h *Handler) Order(c *gin.Context) {
	conf, err := h.orders.Get(uuid.MustParse(c.Param("orderId")))
	respond(c, http.StatusOK, conf, err)
}

// LogReceipt records that a receipt was printed (or reprinted) for an order. Returns 204.
func (h *Handler) LogReceipt(c *gin.Context) {
	p, _ := authiface.PrincipalFrom(c)
	if err := h.receipts.LogPrint(uuid.MustParse(c.Param("orderId")), p.LocationID, p.UserID); err != nil {
		common.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Quote returns the authoritative cart quote — the server picks + applies the single best promotion.
func (h *Handler) Quote(c *gin.Context) {
	var req promotioniface.QuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	resp, err := h.promotions.Quote(uuid.MustParse(c.Param("eventId")), req.Lines)
	respond(c, http.StatusOK, resp, err)
}

// Stock returns live remaining stock per product at the location (the device polls this).
func (h *Handler) Stock(c *gin.Context) {
	rows, err := h.stock.BaLevelsForEvent(uuid.MustParse(c.Param("eventId")))
	respond(c, http.StatusOK, rows, err)
}

// Members returns the active member directory for the device to cache.
func (h *Handler) Members(c *gin.Context) {
	members, err := h.members.ActiveSnapshotMembers(authiface.StoreOf(c))
	respond(c, http.StatusOK, members, err)
}

// MemberByQr is the online QR fallback when a scanned member isn't in the device cache.
func (h *Handler) MemberByQr(c *gin.Context) {
	resp, err := h.members.ByQr(authiface.StoreOf(c), c.Param("qrCode"))
	respond(c, http.StatusOK, resp, err)
}

// IssueNumberBlock reserves a block of order numbers for the calling device, so it can confirm
// orders offline with final numbers.
func (h *Handler) IssueNumberBlock(c *gin.Context) {
	p, _ := authiface.PrincipalFrom(c)
	block, err := h.orders.IssueNumberBlock(uuid.MustParse(c.Param("eventId")), p.DeviceID.String())
	respond(c, http.StatusCreated, block, err)
}

func respond(c *gin.Context, status int, v any, err error) {
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(status, v)
}
