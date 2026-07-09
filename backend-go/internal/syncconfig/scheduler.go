package syncconfig

import (
	"log"
	"sync"

	"github.com/robfig/cron/v3"
)

// Scheduler drives the ERP + member syncs from the DB-stored config instead of a fixed cron. Each
// job re-reads `enabled` at fire time (so toggling never needs a restart), and a config update
// reschedules the entry with the new cron. Mirrors the Kotlin DynamicSyncScheduler.
type Scheduler struct {
	c       *cron.Cron
	cfg     *Service
	jobs    map[SyncType]func()
	entries map[SyncType]cron.EntryID
	mu      sync.Mutex
}

func NewScheduler(cfg *Service, erpJob, memberJob func()) *Scheduler {
	return &Scheduler{
		c:       cron.New(cron.WithParser(cronParser)),
		cfg:     cfg,
		jobs:    map[SyncType]func(){TypeERP: erpJob, TypeMember: memberJob},
		entries: map[SyncType]cron.EntryID{},
	}
}

// Start schedules both syncs from their current config and starts the cron loop.
func (s *Scheduler) Start() {
	for _, t := range []SyncType{TypeERP, TypeMember} {
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
func (s *Scheduler) Reschedule(t SyncType, cronExpr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.entries[t]; ok {
		s.c.Remove(id)
		delete(s.entries, t)
	}
	s.schedule(t, cronExpr)
}

func (s *Scheduler) schedule(t SyncType, cronExpr string) {
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
