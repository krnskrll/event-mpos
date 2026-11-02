package service

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/syncconfig/iface"
	"github.com/krnskrll/event-mpos/internal/syncconfig/repository"
)

// Service owns the DB-stored sync schedules: seeds one row per type, validates cron on update.
type Service struct {
	db                *gorm.DB
	erpDefaultCron    string
	memberDefaultCron string
}

func NewService(db *gorm.DB, erpDefaultCron, memberDefaultCron string) *Service {
	return &Service{db: db, erpDefaultCron: erpDefaultCron, memberDefaultCron: memberDefaultCron}
}

func (s *Service) defaultCron(t iface.SyncType) string {
	if t == iface.TypeERP {
		return s.erpDefaultCron
	}
	return s.memberDefaultCron
}

// SeedDefaults inserts a disabled config row per type on startup if absent.
func (s *Service) SeedDefaults() error {
	for _, t := range []iface.SyncType{iface.TypeERP, iface.TypeMember} {
		var count int64
		if err := s.db.Model(&repository.SyncConfig{}).Where("sync_type = ?", string(t)).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := s.db.Create(&repository.SyncConfig{SyncType: string(t), Enabled: false, Cron: s.defaultCron(t), UpdatedAt: time.Now().UTC()}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// Get returns the config for a type, creating a disabled default if missing.
func (s *Service) Get(t iface.SyncType) (*repository.SyncConfig, error) {
	var cfg repository.SyncConfig
	err := s.db.First(&cfg, "sync_type = ?", string(t)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cfg = repository.SyncConfig{SyncType: string(t), Enabled: false, Cron: s.defaultCron(t), UpdatedAt: time.Now().UTC()}
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
