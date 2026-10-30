package syncconfig

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/common"
)

// Service owns the DB-stored sync schedules: seeds one row per type, validates cron on update.
type Service struct {
	db               *gorm.DB
	erpDefaultCron    string
	memberDefaultCron string
}

func NewService(db *gorm.DB, erpDefaultCron, memberDefaultCron string) *Service {
	return &Service{db: db, erpDefaultCron: erpDefaultCron, memberDefaultCron: memberDefaultCron}
}

func (s *Service) defaultCron(t SyncType) string {
	if t == TypeERP {
		return s.erpDefaultCron
	}
	return s.memberDefaultCron
}

// SeedDefaults inserts a disabled config row per type on startup if absent.
func (s *Service) SeedDefaults() error {
	for _, t := range []SyncType{TypeERP, TypeMember} {
		var count int64
		if err := s.db.Model(&SyncConfig{}).Where("sync_type = ?", string(t)).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := s.db.Create(&SyncConfig{SyncType: string(t), Enabled: false, Cron: s.defaultCron(t), UpdatedAt: time.Now().UTC()}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// Get returns the config for a type, creating a disabled default if missing.
func (s *Service) Get(t SyncType) (*SyncConfig, error) {
	var cfg SyncConfig
	err := s.db.First(&cfg, "sync_type = ?", string(t)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cfg = SyncConfig{SyncType: string(t), Enabled: false, Cron: s.defaultCron(t), UpdatedAt: time.Now().UTC()}
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

func (s *Service) Update(t SyncType, enabled bool, cronExpr string) (*SyncConfig, error) {
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
