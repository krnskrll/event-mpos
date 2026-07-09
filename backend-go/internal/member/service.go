package member

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/common"
)

// Service holds member directory logic — the Go analogue of Kotlin MemberService.
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) List() ([]MemberResponse, error) {
	var members []Member
	if err := s.db.Order("member_code").Find(&members).Error; err != nil {
		return nil, err
	}
	out := make([]MemberResponse, 0, len(members))
	for i := range members {
		out = append(out, members[i].ToResponse())
	}
	return out, nil
}

func (s *Service) activeMembers() ([]Member, error) {
	var members []Member
	if err := s.db.Where("active = ?", true).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// ActiveSnapshotMembers returns active members as compact snapshot DTOs (event.MemberProvider).
func (s *Service) ActiveSnapshotMembers() ([]any, error) {
	members, err := s.activeMembers()
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
func (s *Service) ByQr(qrCode string) (*MemberResponse, error) {
	var m Member
	err := s.db.Where("qr_code = ?", strings.TrimSpace(qrCode)).First(&m).Error
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
func (s *Service) Fingerprint() (string, error) {
	members, err := s.activeMembers()
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

// FindByMemberCode returns the member with the given code, or nil (used by sync).
func (s *Service) FindByMemberCode(code string) (*Member, error) {
	var m Member
	err := s.db.Where("member_code = ?", code).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}
