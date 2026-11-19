package repository

import "github.com/google/uuid"

// StoreSettings is the receipt part of a POS store row.
type StoreSettings struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Phone        string    `gorm:"column:receipt_phone"`
	Instagram    string    `gorm:"column:receipt_instagram"`
	Note         string    `gorm:"column:receipt_note"`
	ShowAddress  bool      `gorm:"column:receipt_show_address"`
	ShowStaff    bool      `gorm:"column:receipt_show_staff"`
	ShowCustomer bool      `gorm:"column:receipt_show_customer"`
}

func (StoreSettings) TableName() string { return "pos_stores" }
