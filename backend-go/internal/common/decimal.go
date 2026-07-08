package common

import "github.com/shopspring/decimal"

// Serialize decimals as JSON numbers (e.g. 129000.00), not quoted strings, to match Jackson's
// BigDecimal output. Set once at package init; common is imported everywhere.
func init() {
	decimal.MarshalJSONWithoutQuotes = true
}
