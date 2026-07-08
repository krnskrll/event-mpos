package event

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

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/user"
)

// Service holds event business logic — the Go analogue of Kotlin EventService.
type Service struct {
	db        *gorm.DB
	users     *user.Service
	providers Providers
}

func NewService(db *gorm.DB, users *user.Service) *Service {
	return &Service{db: db, users: users}
}

// SetProviders wires the cross-domain snapshot contributors once their phases come online.
func (s *Service) SetProviders(p Providers) { s.providers = p }

// ---- admin operations ----

func (s *Service) ListEvents() ([]EventResponse, error) {
	var events []Event
	if err := s.db.Order("created_at").Find(&events).Error; err != nil {
		return nil, err
	}
	out := make([]EventResponse, 0, len(events))
	for i := range events {
		out = append(out, events[i].ToResponse())
	}
	return out, nil
}

func (s *Service) CreateEvent(req CreateEventRequest) (*EventResponse, error) {
	code := strings.TrimSpace(req.Code)
	var count int64
	if err := s.db.Model(&Event{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, common.NewAppError(http.StatusConflict, "An event with code '"+req.Code+"' already exists")
	}
	e := Event{
		Code:      code,
		Name:      strings.TrimSpace(req.Name),
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Status:    StatusDraft,
	}
	if err := s.db.Create(&e).Error; err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

func (s *Service) UpdateEvent(id uuid.UUID, req UpdateEventRequest) (*EventResponse, error) {
	e, err := s.find(id)
	if err != nil {
		return nil, err
	}
	if e.Status == StatusClosed {
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

func (s *Service) ChangeStatus(id uuid.UUID, target EventStatus) (*EventResponse, error) {
	e, err := s.find(id)
	if err != nil {
		return nil, err
	}
	if err := e.Status.AssertCanTransitionTo(target); err != nil {
		return nil, err
	}
	e.Status = target
	e.UpdatedAt = time.Now().UTC()
	if err := s.db.Save(e).Error; err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

// ---- BA assignment ----

func (s *Service) AssignBaToEvent(userID, eventID uuid.UUID) (*EventResponse, error) {
	u, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, common.NewAppError(http.StatusNotFound, "User not found")
	}
	if u.Role != user.RoleBA {
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
	if err := s.db.Create(&BaEventAssignment{UserID: userID, EventID: eventID, Active: true}).Error; err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

// ActiveEventForUser returns the BA's single active event, or nil if none.
func (s *Service) ActiveEventForUser(userID uuid.UUID) (*EventResponse, error) {
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
func (s *Service) AssignedEventsForUser(userID uuid.UUID) ([]EventResponse, error) {
	var assignments []BaEventAssignment
	if err := s.db.Where("user_id = ?", userID).Order("assigned_at").Find(&assignments).Error; err != nil {
		return nil, err
	}
	out := make([]EventResponse, 0, len(assignments))
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
	if err := s.db.Model(&BaEventAssignment{}).
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
func (s *Service) SnapshotForEvent(eventID uuid.UUID) (*EventSnapshot, error) {
	e, err := s.find(eventID)
	if err != nil {
		return nil, err
	}
	if e.Status != StatusActive {
		return nil, common.NewAppError(http.StatusConflict,
			"Event '"+e.Code+"' is "+string(e.Status)+" — selling is only allowed while ACTIVE")
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
	return &EventSnapshot{
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
func (s *Service) versionOf(e *Event) (string, error) {
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

func (s *Service) find(id uuid.UUID) (*Event, error) {
	var e Event
	err := s.db.First(&e, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAppError(http.StatusNotFound, "Event not found")
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Service) activeAssignment(userID uuid.UUID) (*BaEventAssignment, error) {
	var a BaEventAssignment
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
