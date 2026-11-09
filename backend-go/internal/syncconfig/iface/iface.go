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

// MemberSync is the Member POS sync bar in the back office. Time is the daily run in WIB ("HH:MM"),
// nil when the stored schedule is not a daily time.
type MemberSync struct {
	Enabled    bool       `json:"enabled"`
	Time       *string    `json:"time"`
	LastSyncAt *time.Time `json:"lastSyncAt"`
	LastResult *string    `json:"lastResult"`
	Count      int        `json:"count"`
	Source     *string    `json:"source"`
	Message    *string    `json:"message"`
}

type MemberScheduleRequest struct {
	Enabled bool   `json:"enabled"`
	Time    string `json:"time" binding:"required"`
}
