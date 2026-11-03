package common

import (
	"net/http"

	"github.com/shopspring/decimal"
)

// Serialize decimals as JSON numbers (e.g. 129000), not quoted strings, to match Jackson's
// BigDecimal output. Set once at package init; common is imported everywhere.
func init() {
	decimal.MarshalJSONWithoutQuotes = true
}

// Rupiah rounds money to whole rupiah, .5 up; money is stored as BIGINT.
func Rupiah(d decimal.Decimal) decimal.Decimal { return d.Round(0) }

// CheckRupiah refuses an amount with decimals; field is the JSON field name.
func CheckRupiah(field string, d *decimal.Decimal) error {
	if d != nil && !d.Equal(d.Round(0)) {
		return NewAppError(http.StatusBadRequest, field+": must be whole rupiah")
	}
	return nil
}
