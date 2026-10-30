// Package syncconfig owns the DB-stored, runtime-editable ERP/member sync schedules and the dynamic
// cron scheduler that drives them — the Go analogue of the Kotlin sync package.
package syncconfig

import (
	"time"

	"github.com/robfig/cron/v3"
)

// SyncType — the two schedulable syncs (also the PK, one config row each).
type SyncType string

const (
	TypeERP    SyncType = "ERP"
	TypeMember SyncType = "MEMBER"
)

func ParseType(s string) (SyncType, bool) {
	switch SyncType(s) {
	case TypeERP, TypeMember:
		return SyncType(s), true
	}
	return "", false
}

// SyncConfig is a runtime-editable schedule row.
type SyncConfig struct {
	SyncType  string    `gorm:"column:sync_type;primaryKey"`
	Enabled   bool      `gorm:"not null"`
	Cron      string    `gorm:"not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (SyncConfig) TableName() string { return "sync_config" }

// cronParser matches Spring's 6-field cron (seconds first), used for both validation and scheduling.
var cronParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// IsValidCron reports whether expr is a valid 6-field cron expression.
func IsValidCron(expr string) bool {
	_, err := cronParser.Parse(expr)
	return err == nil
}

// ---- DTOs ----

type UpdateSyncConfigRequest struct {
	Enabled bool   `json:"enabled"`
	Cron    string `json:"cron" binding:"required"`
}

// SyncTypeView combines schedule + last-run status for one sync type.
type SyncTypeView struct {
	Type       SyncType   `json:"type"`
	Enabled    bool       `json:"enabled"`
	Cron       string     `json:"cron"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	LastSyncAt *time.Time `json:"lastSyncAt,omitempty"`
	LastResult *string    `json:"lastResult,omitempty"`
	Count      int        `json:"count"`
	Source     *string    `json:"source,omitempty"`
	Message    *string    `json:"message,omitempty"`
}

type SyncOverview struct {
	ERP    SyncTypeView `json:"erp"`
	Member SyncTypeView `json:"member"`
}
