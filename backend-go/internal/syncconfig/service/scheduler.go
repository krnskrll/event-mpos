package service

import (
	"log"
	"sync"

	"github.com/robfig/cron/v3"

	"github.com/krnskrll/event-mpos/internal/syncconfig/iface"
)

// Scheduler drives the ERP + member syncs from the DB-stored config instead of a fixed cron. Each
// job re-reads `enabled` at fire time (so toggling never needs a restart), and a config update
// reschedules the entry with the new cron. Mirrors the Kotlin DynamicSyncScheduler.
type Scheduler struct {
	c       *cron.Cron
	cfg     *Service
	jobs    map[iface.SyncType]func()
	entries map[iface.SyncType]cron.EntryID
	mu      sync.Mutex
}

func NewScheduler(cfg *Service, erpJob, memberJob func()) *Scheduler {
	return &Scheduler{
		c:       cron.New(cron.WithParser(cronParser)),
		cfg:     cfg,
		jobs:    map[iface.SyncType]func(){iface.TypeERP: erpJob, iface.TypeMember: memberJob},
		entries: map[iface.SyncType]cron.EntryID{},
	}
}

// Start schedules both syncs from their current config and starts the cron loop.
func (s *Scheduler) Start() {
	for _, t := range []iface.SyncType{iface.TypeERP, iface.TypeMember} {
		cfg, err := s.cfg.Get(t)
		if err != nil {
			log.Printf("scheduler: cannot load %s config: %v", t, err)
			continue
		}
		s.schedule(t, cfg.Cron)
	}
	s.c.Start()
}

// Reschedule replaces a type's cron entry (called after a config update).
// Stop halts scheduling and waits for a running sync to finish.
func (s *Scheduler) Stop() { <-s.c.Stop().Done() }

func (s *Scheduler) Reschedule(t iface.SyncType, cronExpr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.entries[t]; ok {
		s.c.Remove(id)
		delete(s.entries, t)
	}
	s.schedule(t, cronExpr)
}

func (s *Scheduler) schedule(t iface.SyncType, cronExpr string) {
	job := s.jobs[t]
	id, err := s.c.AddFunc(cronExpr, func() {
		cfg, err := s.cfg.Get(t)
		if err != nil || !cfg.Enabled {
			return
		}
		log.Printf("Scheduled %s sync triggered", t)
		job()
	})
	if err != nil {
		log.Printf("scheduler: invalid cron %q for %s: %v", cronExpr, t, err)
		return
	}
	s.entries[t] = id
}

// cronParser matches Spring's 6-field cron (seconds first), for validation and scheduling.
var cronParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

func IsValidCron(expr string) bool {
	_, err := cronParser.Parse(expr)
	return err == nil
}
