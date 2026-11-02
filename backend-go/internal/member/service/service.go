package service

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/member/iface"
	"github.com/krnskrll/event-mpos/internal/member/repository"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/database"
)

type Service struct {
	db *gorm.DB
}

var _ iface.Members = (*Service)(nil)

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) List(storeID uuid.UUID) ([]iface.MemberResponse, error) {
	var members []repository.Member
	if err := s.db.Where("store_id = ?", storeID).Order("member_code").Find(&members).Error; err != nil {
		return nil, err
	}
	out := make([]iface.MemberResponse, 0, len(members))
	for i := range members {
		out = append(out, members[i].ToResponse())
	}
	return out, nil
}

func (s *Service) activeMembers(storeID uuid.UUID) ([]repository.Member, error) {
	var members []repository.Member
	if err := s.db.Where("store_id = ? AND active = ?", storeID, true).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (s *Service) ActiveSnapshotMembers(storeID uuid.UUID) ([]any, error) {
	members, err := s.activeMembers(storeID)
	if err != nil {
		return nil, err
	}
	out := make([]any, 0, len(members))
	for i := range members {
		out = append(out, members[i].ToSnapshotDto())
	}
	return out, nil
}

// ByQr is the exact QR lookup — the online fallback when a member isn't in the device cache.
func (s *Service) ByQr(storeID uuid.UUID, qrCode string) (*iface.MemberResponse, error) {
	var m repository.Member
	err := s.db.Where("store_id = ? AND qr_code = ?", storeID, strings.TrimSpace(qrCode)).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !m.Active) {
		return nil, common.NewAppError(http.StatusNotFound, "No active member for QR '"+qrCode+"'")
	}
	if err != nil {
		return nil, err
	}
	resp := m.ToResponse()
	return &resp, nil
}

// Fingerprint is folded into the event snapshot version so member changes trigger a re-sync.
// Content-based: "<count>:<javaHashCode(material)>" where material is sorted "qr:name:tier" joined.
func (s *Service) Fingerprint(storeID uuid.UUID) (string, error) {
	members, err := s.activeMembers(storeID)
	if err != nil {
		return "", err
	}
	sort.Slice(members, func(i, j int) bool { return members[i].QRCode < members[j].QRCode })
	parts := make([]string, 0, len(members))
	for i := range members {
		parts = append(parts, members[i].QRCode+":"+members[i].Name+":"+string(members[i].Tier))
	}
	material := strings.Join(parts, ",")
	return strconv.Itoa(len(members)) + ":" + strconv.FormatInt(int64(common.JavaStringHashCode(material)), 10), nil
}

func (s *Service) Upsert(tx database.Tx, storeID uuid.UUID, rec iface.MemberRecord) (bool, error) {
	var m repository.Member
	err := tx.Where("member_code = ?", rec.MemberCode).First(&m).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	created := err != nil
	m.MemberCode, m.QRCode, m.Name, m.Tier = rec.MemberCode, rec.QRCode, rec.Name, rec.Tier
	m.Region, m.Phone, m.Active = rec.Region, rec.Phone, rec.Active
	if created {
		m.StoreID = storeID
		return true, tx.Create(&m).Error
	}
	m.UpdatedAt = time.Now().UTC()
	return false, tx.Save(&m).Error
}
