package iface

import "github.com/google/uuid"

type Receipts interface {
	LogPrint(orderID, locationID, staffID uuid.UUID) error
}

// Settings is the store's Pengaturan Struk, printed by the cashier app on every receipt.
type Settings struct {
	Phone        string `json:"phone"`
	Instagram    string `json:"instagram"`
	Note         string `json:"note"`
	ShowAddress  bool   `json:"showAddress"`
	ShowStaff    bool   `json:"showStaff"`
	ShowCustomer bool   `json:"showCustomer"`
}
