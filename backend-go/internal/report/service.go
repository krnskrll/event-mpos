// Package report produces end-of-event reporting: per-product units sold + revenue joined against
// stock allocations, as JSON or a CSV export.
package report

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/order"
	"github.com/krnskrll/event-mpos/internal/product"
	"github.com/krnskrll/event-mpos/internal/stock"
)

type ProductReportRow struct {
	ProductID uuid.UUID       `json:"productId"`
	SKU       string          `json:"sku"`
	Name      string          `json:"name"`
	UnitsSold int             `json:"unitsSold"`
	Revenue   decimal.Decimal `json:"revenue"`
	Allocated int             `json:"allocated"`
	Sold      int             `json:"sold"`
	Remaining int             `json:"remaining"`
}

type EndOfEventReport struct {
	EventID       uuid.UUID          `json:"eventId"`
	EventCode     string             `json:"eventCode"`
	EventName     string             `json:"eventName"`
	Status        event.EventStatus  `json:"status"`
	GeneratedAt   time.Time          `json:"generatedAt"`
	TotalOrders   int                `json:"totalOrders"`
	TotalItems    int                `json:"totalItems"`
	TotalRevenue  decimal.Decimal    `json:"totalRevenue"`
	TotalDiscount decimal.Decimal    `json:"totalDiscount"`
	Products      []ProductReportRow `json:"products"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) EndOfEvent(eventID uuid.UUID) (*EndOfEventReport, error) {
	var ev event.Event
	if err := s.db.First(&ev, "id = ?", eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewAppError(http.StatusNotFound, "Event not found")
		}
		return nil, err
	}
	// The final report is only meaningful once selling has stopped.
	if ev.Status != event.StatusClosed {
		return nil, common.NewAppError(http.StatusConflict,
			"The end-of-event report is available once "+ev.Code+" is CLOSED (it is "+string(ev.Status)+").")
	}
	var orders []order.Order
	if err := s.db.Preload("Items").Where("event_id = ?", eventID).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	totalRevenue, totalDiscount := decimal.Zero, decimal.Zero
	totalItems := 0
	type agg struct {
		sku, name string
		qty       int
		revenue   decimal.Decimal
	}
	sales := map[uuid.UUID]*agg{}
	var salesOrder []uuid.UUID // preserve first-seen order
	for i := range orders {
		o := &orders[i]
		totalRevenue = totalRevenue.Add(o.Total)
		totalDiscount = totalDiscount.Add(o.Discount)
		for _, it := range o.Items {
			totalItems += it.Qty
			a := sales[it.ProductID]
			if a == nil {
				a = &agg{sku: it.SKU, name: it.Name, revenue: decimal.Zero}
				sales[it.ProductID] = a
				salesOrder = append(salesOrder, it.ProductID)
			}
			a.qty += it.Qty
			a.revenue = a.revenue.Add(it.LineTotal)
		}
	}

	var allocs []stock.EventStockAllocation
	if err := s.db.Where("event_id = ?", eventID).Find(&allocs).Error; err != nil {
		return nil, err
	}
	allocByProduct := map[uuid.UUID]*stock.EventStockAllocation{}
	var allocOrder []uuid.UUID
	for i := range allocs {
		allocByProduct[allocs[i].ProductID] = &allocs[i]
		allocOrder = append(allocOrder, allocs[i].ProductID)
	}

	// Union: products that sold or had stock allocated (sales first, then allocation-only).
	seen := map[uuid.UUID]bool{}
	var ids []uuid.UUID
	for _, id := range salesOrder {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	for _, id := range allocOrder {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}

	rows := make([]ProductReportRow, 0, len(ids))
	for _, pid := range ids {
		a := sales[pid]
		al := allocByProduct[pid]
		row := ProductReportRow{ProductID: pid, Revenue: decimal.Zero.Round(2)}
		if a != nil {
			row.SKU, row.Name, row.UnitsSold, row.Revenue = a.sku, a.name, a.qty, a.revenue.Round(2)
		} else {
			var p product.Product
			if err := s.db.First(&p, "id = ?", pid).Error; err == nil {
				row.SKU, row.Name = p.SKU, p.Name
			} else {
				row.SKU, row.Name = "—", "(unknown)"
			}
		}
		if al != nil {
			row.Allocated, row.Sold, row.Remaining = al.AllocatedQty, al.SoldQty, al.Remaining()
		}
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].UnitsSold > rows[j].UnitsSold })

	return &EndOfEventReport{
		EventID: ev.ID, EventCode: ev.Code, EventName: ev.Name, Status: ev.Status,
		GeneratedAt: time.Now().UTC(), TotalOrders: len(orders), TotalItems: totalItems,
		TotalRevenue: totalRevenue.Round(2), TotalDiscount: totalDiscount.Round(2), Products: rows,
	}, nil
}

// ToCSV renders the report as CSV: a summary block, a blank line, then one row per product.
func (s *Service) ToCSV(r *EndOfEventReport) string {
	var sb strings.Builder
	line := func(s string) { sb.WriteString(s); sb.WriteString("\n") }
	line("End-of-Event Report," + esc(r.EventCode+" — "+r.EventName))
	line("Status," + string(r.Status))
	line("Generated," + r.GeneratedAt.Format(time.RFC3339))
	line("Total Orders," + itoa(r.TotalOrders))
	line("Total Items," + itoa(r.TotalItems))
	// StringFixed(2) matches Kotlin BigDecimal's fixed 2-decimal rendering (Round/String drops zeros).
	line("Total Revenue," + r.TotalRevenue.StringFixed(2))
	line("Total Discount," + r.TotalDiscount.StringFixed(2))
	line("")
	line("SKU,Product,Units Sold,Revenue,Allocated,Sold,Remaining")
	for _, p := range r.Products {
		line(esc(p.SKU) + "," + esc(p.Name) + "," + itoa(p.UnitsSold) + "," + p.Revenue.StringFixed(2) + "," +
			itoa(p.Allocated) + "," + itoa(p.Sold) + "," + itoa(p.Remaining))
	}
	return sb.String()
}

func esc(s string) string {
	if strings.ContainsAny(s, ",\"") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

func itoa(i int) string { return strconv.Itoa(i) }
