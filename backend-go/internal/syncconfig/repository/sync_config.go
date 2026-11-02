package repository

import "time"

// SyncConfig is a runtime-editable schedule row.
type SyncConfig struct {
	SyncType  string    `gorm:"column:sync_type;primaryKey"`
	Enabled   bool      `gorm:"not null"`
	Cron      string    `gorm:"not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (SyncConfig) TableName() string { return "sync_config" }
