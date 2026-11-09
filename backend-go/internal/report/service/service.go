// Package service produces the back-office sales report of a location.
package service

import (
	"sort"
	"time"

	"github.com/google/uuid"

	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/report/iface"
)

type Service struct {
	events eventiface.Events
	orders orderiface.Orders
}

func NewService(events eventiface.Events, orders orderiface.Orders) *Service {
	return &Service{events: events, orders: orders}
}

// Day is midnight WIB of a YYYY-MM-DD date.
func Day(date string) (time.Time, error) { return time.ParseInLocation(time.DateOnly, date, wib) }

// Today is the current WIB date.
func Today(now time.Time) string { return now.In(wib).Format(time.DateOnly) }

var wib = time.FixedZone("WIB", 7*3600)

// Sales is a location's sales from the start of from to the end of to (WIB dates).
func (s *Service) Sales(locationID uuid.UUID, from, to time.Time, now time.Time) (*iface.SalesReport, error) {
	ev, err := s.events.Get(locationID)
	if err != nil {
		return nil, err
	}
	end := to.AddDate(0, 0, 1)
	orders, cancelled, err := s.orders.ForReport(locationID, from, end)
	if err != nil {
		return nil, err
	}
	r := &iface.SalesReport{
		LocationID: locationID, From: from.Format(time.DateOnly), To: to.Format(time.DateOnly),
		Final:       ev.Status == eventiface.StatusClosed || now.Sub(end) > 24*time.Hour,
		TotalOrders: len(orders), CancelledOrders: cancelled, Products: []iface.SalesRow{},
	}
	type key struct {
		id      uuid.UUID
		variant string
	}
	rows := map[key]*iface.SalesRow{}
	promos := map[uuid.UUID]bool{}
	for _, o := range orders {
		if o.LateSynced {
			r.LateSyncedOrders++
		}
		if o.PromotionID != nil {
			promos[*o.PromotionID] = true
		}
		r.TotalRevenue = r.TotalRevenue.Add(o.Total)
		r.TotalDiscount = r.TotalDiscount.Add(o.Discount)
		for _, it := range o.Items {
			k := key{it.ProductID, ""}
			name := it.Name
			if it.VariantName != nil && *it.VariantName != "" {
				k.variant, name = *it.VariantName, it.Name+" · "+*it.VariantName
			}
			row := rows[k]
			if row == nil {
				row = &iface.SalesRow{ProductID: it.ProductID, SKU: it.SKU, Name: name}
				rows[k] = row
			}
			row.UnitsSold += it.Qty
			row.Revenue = row.Revenue.Add(it.LineTotal)
			r.TotalItems += it.Qty
		}
	}
	items := map[uuid.UUID]bool{}
	for _, row := range rows {
		r.Products = append(r.Products, *row)
		items[row.ProductID] = true
	}
	sort.Slice(r.Products, func(i, j int) bool {
		a, b := r.Products[i], r.Products[j]
		if a.UnitsSold != b.UnitsSold {
			return a.UnitsSold > b.UnitsSold
		}
		return a.Name < b.Name
	})
	r.ProductCount, r.PromotionsUsed = len(items), len(promos)
	return r, nil
}
