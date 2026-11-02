// Package erp is the pluggable ERP boundary: swapping between API-pull, file-import, or the dev mock
// is just a different Source implementation — the sync service and everything downstream are unchanged.
package erp

import (
	"github.com/shopspring/decimal"
)

// Barcode is a scannable code with an optional variant label.
type Barcode struct {
	Code  string
	Label *string
}

// Product is a normalized product record as it comes out of the ERP, regardless of transport.
type Product struct {
	SKU           string
	Name          string
	Brand         *string
	Price         decimal.Decimal
	ImageURL      *string
	StockQuantity int
	Barcodes      []Barcode
}

// Source pulls the full current product list from the ERP. Selected by config app.erp.source:
// mock (dev default) | file (CSV) | rest (REST API).
type Source interface {
	// Description is a human label for logs / sync status (e.g. "mock", "file:products.csv").
	Description() string
	// FetchProducts pulls the full current product list; returns an error on transport/parse failure.
	FetchProducts() ([]Product, error)
}

func strptr(s string) *string { return &s }
