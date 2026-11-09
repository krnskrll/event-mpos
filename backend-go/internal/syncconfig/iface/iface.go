package iface

import "time"

// SyncType is one schedulable sync; also the primary key of its config row. The ERP product sync
// is retired: Bhoomy owns products.
type SyncType string

const TypeMember SyncType = "MEMBER"

// LastRun is the latest outcome of a sync job.
type LastRun struct {
	At      *time.Time
	Result  *string
	Count   int
	Source  *string
	Message *string
}

// StatusReporter is implemented by the member sync.
type StatusReporter interface {
	LastRun() LastRun
}

type UpdateSyncConfigRequest struct {
	Enabled bool   `json:"enabled"`
	Cron    string `json:"cron" binding:"required"`
}

// SyncTypeView combines schedule and last-run status for one sync type.
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
	Member SyncTypeView `json:"member"`
}
