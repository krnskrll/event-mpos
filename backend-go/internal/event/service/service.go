package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/event/repository"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

type Service struct {
	db        *gorm.DB
	users     useriface.Users
	providers iface.Providers
}

var _ iface.Events = (*Service)(nil)

func NewService(db *gorm.DB, users useriface.Users) *Service {
	return &Service{db: db, users: users}
}

func (s *Service) SetProviders(p iface.Providers) { s.providers = p }

func (s *Service) Get(id uuid.UUID) (*iface.EventResponse, error) {
	e, err := s.find(id)
	if err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

func (s *Service) ListEvents() ([]iface.EventResponse, error) {
	var events []repository.Event
	if err := s.db.Order("created_at").Find(&events).Error; err != nil {
		return nil, err
	}
	out := make([]iface.EventResponse, 0, len(events))
	for i := range events {
		out = append(out, events[i].ToResponse())
	}
	return out, nil
}

func (s *Service) CreateEvent(req iface.CreateEventRequest) (*iface.EventResponse, error) {
	code := strings.TrimSpace(req.Code)
	var count int64
	if err := s.db.Model(&repository.Event{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, common.NewAppError(http.StatusConflict, "An event with code '"+req.Code+"' already exists")
	}
	e := repository.Event{
		Code:      code,
		Name:      strings.TrimSpace(req.Name),
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Status:    iface.StatusDraft,
	}
	if err := s.db.Create(&e).Error; err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

func (s *Service) UpdateEvent(id uuid.UUID, req iface.UpdateEventRequest) (*iface.EventResponse, error) {
	e, err := s.find(id)
	if err != nil {
		return nil, err
	}
	if e.Status == iface.StatusClosed {
		return nil, common.NewAppError(http.StatusConflict, "A closed event can no longer be edited")
	}
	if req.Name != nil {
		e.Name = strings.TrimSpace(*req.Name)
	}
	if req.StartDate != nil {
		e.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		e.EndDate = req.EndDate
	}
	e.UpdatedAt = time.Now().UTC()
	if err := s.db.Save(e).Error; err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

func (s *Service) ChangeStatus(id uuid.UUID, target iface.EventStatus) (*iface.EventResponse, error) {
	e, err := s.find(id)
	if err != nil {
		return nil, err
	}
	if err := e.Status.AssertCanTransitionTo(target); err != nil {
		return nil, err
	}
	e.Status = target
	e.UpdatedAt = time.Now().UTC()
	if target == iface.StatusClosed {
		e.ClosedAt = &e.UpdatedAt
	}
	if err := s.db.Save(e).Error; err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

// ---- BA assignment ----

func (s *Service) AssignBaToEvent(userID, eventID uuid.UUID) (*iface.EventResponse, error) {
	u, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, common.NewAppError(http.StatusNotFound, "User not found")
	}
	if u.Role != useriface.RoleBA {
		return nil, common.NewAppError(http.StatusBadRequest, "Only BA accounts can be assigned to an event")
	}
	e, err := s.find(eventID)
	if err != nil {
		return nil, err
	}
	current, err := s.activeAssignment(userID)
	if err != nil {
		return nil, err
	}
	if current != nil {
		if current.EventID == eventID {
			resp := e.ToResponse()
			return &resp, nil
		}
		current.Active = false
		if err := s.db.Save(current).Error; err != nil {
			return nil, err
		}
	}
	if err := s.db.Create(&repository.BaEventAssignment{UserID: userID, EventID: eventID, Active: true}).Error; err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

// ActiveEventForUser returns the BA's single active event, or nil if none.
func (s *Service) ActiveEventForUser(userID uuid.UUID) (*iface.EventResponse, error) {
	a, err := s.activeAssignment(userID)
	if err != nil || a == nil {
		return nil, err
	}
	e, err := s.find(a.EventID)
	if err != nil {
		// A dangling assignment (event deleted) is treated as "no active event".
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

// AssignedEventsForUser returns all events a BA has been assigned to (history + current).
func (s *Service) AssignedEventsForUser(userID uuid.UUID) ([]iface.EventResponse, error) {
	var assignments []repository.BaEventAssignment
	if err := s.db.Where("user_id = ?", userID).Order("assigned_at").Find(&assignments).Error; err != nil {
		return nil, err
	}
	out := make([]iface.EventResponse, 0, len(assignments))
	for _, a := range assignments {
		e, err := s.find(a.EventID)
		if err != nil {
			if isNotFound(err) {
				continue
			}
			return nil, err
		}
		out = append(out, e.ToResponse())
	}
	return out, nil
}

func (s *Service) RequireAssigned(userID, eventID uuid.UUID) error {
	var count int64
	if err := s.db.Model(&repository.BaEventAssignment{}).
		Where("user_id = ? AND event_id = ?", userID, eventID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return common.NewAppError(http.StatusForbidden, "You are not assigned to this event")
	}
	return nil
}

// ---- snapshot ----

// SnapshotForEvent builds the BA event snapshot. Selling (and therefore a snapshot) is only
// available while the event is ACTIVE.
func (s *Service) SnapshotForEvent(eventID uuid.UUID) (*iface.EventSnapshot, error) {
	e, err := s.find(eventID)
	if err != nil {
		return nil, err
	}
	if e.Status != iface.StatusActive {
		return nil, common.NewAppError(http.StatusConflict,
			"repository.Event '"+e.Code+"' is "+string(e.Status)+" — selling is only allowed while ACTIVE")
	}
	var products, members, promotions []any
	if s.providers.Products != nil {
		if products, err = s.providers.Products.AvailableSnapshotProducts(eventID); err != nil {
			return nil, err
		}
	}
	if s.providers.Members != nil {
		if members, err = s.providers.Members.ActiveSnapshotMembers(); err != nil {
			return nil, err
		}
	}
	if s.providers.Promotions != nil {
		if promotions, err = s.providers.Promotions.SnapshotPromotions(eventID); err != nil {
			return nil, err
		}
	}
	version, err := s.versionOf(e)
	if err != nil {
		return nil, err
	}
	return &iface.EventSnapshot{
		Event:       e.ToResponse(),
		Version:     version,
		Products:    emptyIfNil(products),
		Stock:       []any{},
		Members:     emptyIfNil(members),
		Promotions:  emptyIfNil(promotions),
		MemberCount: len(members),
	}, nil
}

// SnapshotVersion is the current content hash used by the device to decide whether to re-sync.
func (s *Service) SnapshotVersion(eventID uuid.UUID) (string, error) {
	e, err := s.find(eventID)
	if err != nil {
		return "", err
	}
	return s.versionOf(e)
}

// versionOf hashes everything that should trigger a device re-sync: the event's own mutable fields
// plus each domain's fingerprint (product availability, member directory, allocation config,
// promotion config). Mirrors Kotlin EventService.versionOf — SHA-256 hex, first 16 chars.
func (s *Service) versionOf(e *repository.Event) (string, error) {
	productFp, err := s.productFingerprint(e.ID)
	if err != nil {
		return "", err
	}
	memberFp, err := s.memberFingerprint()
	if err != nil {
		return "", err
	}
	stockFp, err := s.stockFingerprint(e.ID)
	if err != nil {
		return "", err
	}
	promoFp, err := s.promotionFingerprint(e.ID)
	if err != nil {
		return "", err
	}
	material := strings.Join([]string{
		e.ID.String(),
		e.Name,
		e.StartDate.String(),
		e.EndDate.String(),
		string(e.Status),
		strconv.FormatInt(e.UpdatedAt.UnixMilli(), 10),
		productFp,
		memberFp,
		stockFp,
		promoFp,
	}, "|")
	sum := sha256.Sum256([]byte(material))
	return hex.EncodeToString(sum[:])[:16], nil
}

func (s *Service) productFingerprint(eventID uuid.UUID) (string, error) {
	if s.providers.Products == nil {
		return "", nil
	}
	return s.providers.Products.AvailabilityFingerprint(eventID)
}

func (s *Service) memberFingerprint() (string, error) {
	if s.providers.Members == nil {
		return "", nil
	}
	return s.providers.Members.Fingerprint()
}

func (s *Service) stockFingerprint(eventID uuid.UUID) (string, error) {
	if s.providers.Stock == nil {
		return "", nil
	}
	return s.providers.Stock.AllocationFingerprint(eventID)
}

func (s *Service) promotionFingerprint(eventID uuid.UUID) (string, error) {
	if s.providers.Promotions == nil {
		return "", nil
	}
	return s.providers.Promotions.Fingerprint(eventID)
}

// ---- helpers ----

func (s *Service) find(id uuid.UUID) (*repository.Event, error) {
	var e repository.Event
	err := s.db.First(&e, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAppError(http.StatusNotFound, "repository.Event not found")
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Service) activeAssignment(userID uuid.UUID) (*repository.BaEventAssignment, error) {
	var a repository.BaEventAssignment
	err := s.db.Where("user_id = ? AND active = ?", userID, true).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func isNotFound(err error) bool {
	ae, ok := err.(*common.AppError)
	return ok && ae.Status == http.StatusNotFound
}

func emptyIfNil(s []any) []any {
	if s == nil {
		return []any{}
	}
	return s
}
