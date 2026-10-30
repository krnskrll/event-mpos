package erp

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// FileSource reads a CSV export — the most common ERP file format. Active when app.erp.source=file.
// Expected columns: sku,name,brand,price,imageUrl,stock,barcodes
// where `barcodes` is "code|label;code|label" (label optional).
type FileSource struct {
	Path string
}

func (f FileSource) Description() string { return "file:" + f.Path }

func (f FileSource) FetchProducts() ([]Product, error) {
	path := strings.TrimPrefix(f.Path, "classpath:")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ERP file not found at %s: %w", f.Path, err)
	}
	defer file.Close()

	var out []Product
	scanner := bufio.NewScanner(file)
	first := true
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if first { // skip header row
			first = false
			continue
		}
		cols := strings.Split(line, ",")
		for i := range cols {
			cols[i] = strings.TrimSpace(cols[i])
		}
		if len(cols) < 7 {
			return nil, fmt.Errorf("malformed CSV row (need 7 columns): %s", line)
		}
		price, err := decimal.NewFromString(cols[3])
		if err != nil {
			price = decimal.Zero
		}
		stock, _ := strconv.Atoi(cols[5])
		out = append(out, Product{
			SKU:           cols[0],
			Name:          cols[1],
			Brand:         blankToNil(cols[2]),
			Price:         price,
			ImageURL:      blankToNil(cols[4]),
			StockQuantity: stock,
			Barcodes:      parseBarcodes(cols[6]),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func parseBarcodes(raw string) []Barcode {
	var out []Barcode
	for _, entry := range strings.Split(raw, ";") {
		parts := strings.Split(entry, "|")
		code := strings.TrimSpace(parts[0])
		if code == "" {
			continue
		}
		var label *string
		if len(parts) > 1 {
			label = blankToNil(strings.TrimSpace(parts[1]))
		}
		out = append(out, Barcode{Code: code, Label: label})
	}
	return out
}

func blankToNil(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}
