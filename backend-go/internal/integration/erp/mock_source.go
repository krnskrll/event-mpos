package erp

import "github.com/shopspring/decimal"

// MockSource returns a small fixed catalogue so the whole sync pipeline works out of the box
// without a real ERP. Active when app.erp.source is unset or "mock".
type MockSource struct{}

func (MockSource) Description() string { return "mock" }

// Prices are in Rupiah (whole thousands). imageUrl is nil — the apps render a branded placeholder
// rather than an unrepresentative stock photo.
func (MockSource) FetchProducts() ([]Product, error) {
	d := func(v string) decimal.Decimal { return decimal.RequireFromString(v) }
	return []Product{
		{SKU: "SKU-1001", Name: "Hydra Glow Serum 30ml", Brand: strptr("LumièreCo"), Price: d("129000"),
			StockQuantity: 120, Barcodes: []Barcode{{Code: "8990001001", Label: strptr("single")}, {Code: "18990001001", Label: strptr("box of 6")}}},
		{SKU: "SKU-1002", Name: "Velvet Matte Lipstick — Rose", Brand: strptr("LumièreCo"), Price: d("79000"),
			StockQuantity: 300, Barcodes: []Barcode{{Code: "8990001002"}}},
		{SKU: "SKU-1003", Name: "Daily Mineral Sunscreen SPF50", Brand: strptr("DermaPure"), Price: d("99000"),
			StockQuantity: 80, Barcodes: []Barcode{{Code: "8990001003", Label: strptr("single")}, {Code: "28990001003", Label: strptr("twin pack")}}},
		{SKU: "SKU-1004", Name: "Nourishing Hair Oil 100ml", Brand: strptr("Botané"), Price: d("65000"),
			StockQuantity: 45, Barcodes: []Barcode{{Code: "8990001004"}}},
		{SKU: "SKU-1005", Name: "Gentle Foaming Cleanser 150ml", Brand: strptr("DermaPure"), Price: d("55000"),
			StockQuantity: 200, Barcodes: []Barcode{{Code: "8990001005"}}},
	}, nil
}
