package erp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// RestSource calls a configured ERP endpoint returning product JSON and maps it into Product.
// Active when app.erp.source=rest; configure the URL with app.erp.rest-url.
type RestSource struct {
	URL string
}

func (r RestSource) Description() string { return "rest:" + r.URL }

type restProductJSON struct {
	SKU           string           `json:"sku"`
	Name          string           `json:"name"`
	Brand         *string          `json:"brand"`
	Price         *decimal.Decimal `json:"price"`
	ImageURL      *string          `json:"imageUrl"`
	StockQuantity *int             `json:"stockQuantity"`
	Barcodes      []struct {
		Code  string  `json:"code"`
		Label *string `json:"label"`
	} `json:"barcodes"`
}

func (r RestSource) FetchProducts() ([]Product, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(r.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ERP REST returned status %d", resp.StatusCode)
	}
	var raw []restProductJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]Product, 0, len(raw))
	for _, it := range raw {
		price := decimal.Zero
		if it.Price != nil {
			price = *it.Price
		}
		stock := 0
		if it.StockQuantity != nil {
			stock = *it.StockQuantity
		}
		barcodes := make([]Barcode, 0, len(it.Barcodes))
		for _, b := range it.Barcodes {
			barcodes = append(barcodes, Barcode{Code: b.Code, Label: b.Label})
		}
		out = append(out, Product{
			SKU: it.SKU, Name: it.Name, Brand: it.Brand, Price: price,
			ImageURL: it.ImageURL, StockQuantity: stock, Barcodes: barcodes,
		})
	}
	return out, nil
}

// NewSource selects the ERP source from config (mock | file | rest).
func NewSource(kind, filePath, restURL string) Source {
	switch kind {
	case "file":
		return FileSource{Path: filePath}
	case "rest":
		return RestSource{URL: restURL}
	default:
		return MockSource{}
	}
}
