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

func (s *Service) RequireInStore(storeID, eventID uuid.UUID) error {
	_, err := s.findIn(storeID, eventID)
	return err
}

func (s *Service) ListEvents(storeID uuid.UUID) ([]iface.EventResponse, error) {
	var events []repository.Event
	if err := s.db.Where("store_id = ?", storeID).Order("created_at").Find(&events).Error; err != nil {
		return nil, err
	}
	out := make([]iface.EventResponse, 0, len(events))
	for i := range events {
		out = append(out, events[i].ToResponse())
	}
	return out, nil
}

func (s *Service) CreateEvent(storeID uuid.UUID, req iface.CreateEventRequest) (*iface.EventResponse, error) {
	outletType, ok := iface.ParseLocationType(string(req.Type))
	if !ok {
		return nil, common.NewAppError(http.StatusBadRequest, "type: must be one of event, outlet")
	}
	if err := validateOutlet(outletType, req.Address, req.EndDate); err != nil {
		return nil, err
	}
	code := strings.TrimSpace(req.Code)
	var count int64
	if err := s.db.Model(&repository.Event{}).Where("store_id = ? AND code = ?", storeID, code).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, common.NewAppError(http.StatusConflict, "An event with code '"+req.Code+"' already exists")
	}
	e := repository.Event{
		StoreID:   storeID,
		Type:      outletType,
		Address:   trimmed(req.Address),
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

func (s *Service) UpdateEvent(storeID, id uuid.UUID, req iface.UpdateEventRequest) (*iface.EventResponse, error) {
	e, err := s.findIn(storeID, id)
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
	if req.Address != nil {
		e.Address = trimmed(req.Address)
	}
	if err := validateOutlet(e.Type, e.Address, e.EndDate); err != nil {
		return nil, err
	}
	e.UpdatedAt = time.Now().UTC()
	if err := s.db.Save(e).Error; err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

func (s *Service) ChangeStatus(storeID, id uuid.UUID, target iface.EventStatus) (*iface.EventResponse, error) {
	e, err := s.findIn(storeID, id)
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

func (s *Service) AssignBaToEvent(storeID, userID, eventID uuid.UUID) (*iface.EventResponse, error) {
	u, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil || u.StoreID != storeID {
		return nil, common.NewAppError(http.StatusNotFound, "User not found")
	}
	if u.Role != useriface.RoleBA {
		return nil, common.NewAppError(http.StatusBadRequest, "Only BA accounts can be assigned to an event")
	}
	e, err := s.findIn(storeID, eventID)
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
			"Event '"+e.Code+"' is "+string(e.Status)+" — selling is only allowed while ACTIVE")
	}
	var products, members, promotions []any
	if s.providers.Products != nil {
		if products, err = s.providers.Products.AvailableSnapshotProducts(eventID); err != nil {
			return nil, err
		}
	}
	if s.providers.Members != nil {
		if members, err = s.providers.Members.ActiveSnapshotMembers(e.StoreID); err != nil {
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
	memberFp, err := s.memberFingerprint(e.StoreID)
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

func (s *Service) memberFingerprint(storeID uuid.UUID) (string, error) {
	if s.providers.Members == nil {
		return "", nil
	}
	return s.providers.Members.Fingerprint(storeID)
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
		return nil, common.NewAppError(http.StatusNotFound, "Event not found")
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Service) findIn(storeID, id uuid.UUID) (*repository.Event, error) {
	e, err := s.find(id)
	if err != nil {
		return nil, err
	}
	if e.StoreID != storeID {
		return nil, common.NewAppError(http.StatusNotFound, "Event not found")
	}
	return e, nil
}

// validateOutlet: an outlet is a permanent shop, so it needs an address and has no end date.
func validateOutlet(t iface.LocationType, address *string, endDate *common.Date) error {
	if t != iface.TypeOutlet {
		return nil
	}
	if address == nil || strings.TrimSpace(*address) == "" {
		return common.NewAppError(http.StatusBadRequest, "address: required for an outlet")
	}
	if endDate != nil {
		return common.NewAppError(http.StatusBadRequest, "endDate: an outlet has no end date")
	}
	return nil
}

func trimmed(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	return &t
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
