package sync

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/member"
)

// Status is a single-row (id=1) record of the most recent member sync.
type Status struct {
	ID          int        `gorm:"primaryKey" json:"id"`
	LastSyncAt  *time.Time `gorm:"column:last_sync_at" json:"lastSyncAt,omitempty"`
	LastResult  *string    `gorm:"column:last_result" json:"lastResult,omitempty"`
	MemberCount int        `gorm:"column:member_count" json:"memberCount"`
	Source      *string    `json:"source,omitempty"`
	Message     *string    `json:"message,omitempty"`
}

func (Status) TableName() string { return "member_sync_status" }

// Result is returned by POST /api/admin/members/sync.
type Result struct {
	Ok      bool    `json:"ok"`
	Created int     `json:"created"`
	Updated int     `json:"updated"`
	Total   int     `json:"total"`
	Source  string  `json:"source"`
	Message *string `json:"message,omitempty"`
}

// Service pulls members from the active Source and upserts them by memberCode (idempotent). An
// optional region filter narrows large datasets at sync time.
type Service struct {
	source       Source
	db           *gorm.DB
	regionFilter string
}

func NewService(source Source, db *gorm.DB, regionFilter string) *Service {
	return &Service{source: source, db: db, regionFilter: regionFilter}
}

func (s *Service) Sync() Result {
	log.Printf("Member sync starting from source: %s", s.source.Description())
	records, err := s.source.FetchMembers()
	if err != nil {
		return s.fail(err)
	}
	filtered := records[:0:0]
	for _, r := range records {
		if s.regionFilter == "" || (r.Region != nil && strings.EqualFold(*r.Region, s.regionFilter)) {
			filtered = append(filtered, r)
		}
	}

	created, updated := 0, 0
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		for _, rec := range filtered {
			var existing member.Member
			err := tx.Where("member_code = ?", rec.MemberCode).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				m := member.Member{
					MemberCode: rec.MemberCode, QRCode: rec.QRCode, Name: rec.Name,
					Tier: rec.Tier, Region: rec.Region, Phone: rec.Phone, Active: rec.Active,
				}
				if err := tx.Create(&m).Error; err != nil {
					return err
				}
				created++
				continue
			}
			if err != nil {
				return err
			}
			existing.QRCode = rec.QRCode
			existing.Name = rec.Name
			existing.Tier = rec.Tier
			existing.Region = rec.Region
			existing.Phone = rec.Phone
			existing.Active = rec.Active
			existing.UpdatedAt = time.Now().UTC()
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			updated++
		}
		return nil
	})
	if txErr != nil {
		return s.fail(txErr)
	}

	result := Result{Ok: true, Created: created, Updated: updated, Total: len(filtered), Source: s.source.Description()}
	s.saveStatus(result)
	log.Printf("Member sync done: %d created, %d updated, %d total", created, updated, len(filtered))
	return result
}

func (s *Service) Status() Status {
	var st Status
	if err := s.db.First(&st, 1).Error; err != nil {
		return Status{ID: 1}
	}
	return st
}

func (s *Service) fail(err error) Result {
	log.Printf("Member sync failed: %v", err)
	msg := err.Error()
	result := Result{Ok: false, Source: s.source.Description(), Message: &msg}
	s.saveStatus(result)
	return result
}

func (s *Service) saveStatus(result Result) {
	var st Status
	if err := s.db.First(&st, 1).Error; err != nil {
		st = Status{ID: 1}
	}
	now := time.Now().UTC()
	st.LastSyncAt = &now
	res := "SUCCESS"
	if !result.Ok {
		res = "FAILED"
	}
	st.LastResult = &res
	st.MemberCount = result.Total
	src := result.Source
	st.Source = &src
	if result.Message != nil {
		st.Message = result.Message
	} else {
		msg := "created=" + strconv.Itoa(result.Created) + ", updated=" + strconv.Itoa(result.Updated)
		st.Message = &msg
	}
	s.db.Save(&st)
}
