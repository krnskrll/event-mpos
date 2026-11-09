package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/syncconfig/iface"
	"github.com/krnskrll/event-mpos/internal/syncconfig/repository"
)

// Service owns the DB-stored member sync schedule and validates cron on update.
type Service struct {
	db          *gorm.DB
	defaultCron string
}

func NewService(db *gorm.DB, memberDefaultCron string) *Service {
	return &Service{db: db, defaultCron: memberDefaultCron}
}

// SeedDefaults inserts a disabled member config row on startup if absent.
func (s *Service) SeedDefaults() error {
	_, err := s.Get(iface.TypeMember)
	return err
}

// Get returns the config for a type, creating a disabled default if missing.
func (s *Service) Get(t iface.SyncType) (*repository.SyncConfig, error) {
	var cfg repository.SyncConfig
	err := s.db.First(&cfg, "sync_type = ?", string(t)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cfg = repository.SyncConfig{SyncType: string(t), Enabled: false, Cron: s.defaultCron, UpdatedAt: time.Now().UTC()}
		if err := s.db.Create(&cfg).Error; err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Service) Update(t iface.SyncType, enabled bool, cronExpr string) (*repository.SyncConfig, error) {
	cronExpr = strings.TrimSpace(cronExpr)
	if !IsValidCron(cronExpr) {
		return nil, common.NewAppError(http.StatusBadRequest, "Invalid cron expression: '"+cronExpr+"'")
	}
	cfg, err := s.Get(t)
	if err != nil {
		return nil, err
	}
	cfg.Enabled = enabled
	cfg.Cron = cronExpr
	cfg.UpdatedAt = time.Now().UTC()
	if err := s.db.Save(cfg).Error; err != nil {
		return nil, err
	}
	return cfg, nil
}

// Schedules set from the back office run daily at a WIB time, whatever the server's time zone.
const dailyPrefix = "CRON_TZ=Asia/Jakarta "

// DailyCron turns "HH:MM" (WIB) into the stored cron.
func DailyCron(hhmm string) (string, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(hhmm))
	if err != nil {
		return "", common.NewAppError(http.StatusBadRequest, "time: must be HH:MM")
	}
	return fmt.Sprintf("%s0 %d %d * * *", dailyPrefix, t.Minute(), t.Hour()), nil
}

// DailyTime reads "HH:MM" back from a cron DailyCron made; nil for any other schedule.
func DailyTime(cron string) *string {
	var m, h int
	if _, err := fmt.Sscanf(cron, dailyPrefix+"0 %d %d * * *", &m, &h); err != nil || h > 23 || m > 59 {
		return nil
	}
	t := fmt.Sprintf("%02d:%02d", h, m)
	return &t
}
