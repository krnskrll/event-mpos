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
	"github.com/krnskrll/event-mpos/internal/platform/database"
)

type Service struct {
	db        *gorm.DB
	providers iface.Providers
	onClose   func(tx database.Tx, locationID uuid.UUID, closedAt time.Time) error
}

var _ iface.Events = (*Service)(nil)

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) SetProviders(p iface.Providers) { s.providers = p }

// SetOnClose runs fn in the same transaction when a location closes.
func (s *Service) SetOnClose(fn func(tx database.Tx, locationID uuid.UUID, closedAt time.Time) error) {
	s.onClose = fn
}

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
	if err := assertTransition(e, target); err != nil {
		return nil, err
	}
	e.Status = target
	e.UpdatedAt = time.Now().UTC()
	if target == iface.StatusClosed {
		e.ClosedAt = &e.UpdatedAt
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(e).Error; err != nil {
			return err
		}
		if target == iface.StatusClosed && s.onClose != nil {
			return s.onClose(tx, e.ID, *e.ClosedAt)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

// assertTransition: an event runs Draf → Aktif → Tutup. The outlet is the Cabang, so it never closes
// here; it pauses (Aktif → Draf) and starts again.
func assertTransition(e *repository.Event, target iface.EventStatus) error {
	if e.Type != iface.TypeOutlet {
		return e.Status.AssertCanTransitionTo(target)
	}
	if (e.Status == iface.StatusDraft && target == iface.StatusActive) || (e.Status == iface.StatusActive && target == iface.StatusDraft) {
		return nil
	}
	return common.NewAppError(http.StatusConflict, "an outlet is the Cabang: pause selling instead of closing it")
}

// ---- snapshot ----

// SnapshotForEvent builds the cashier snapshot of a location. Selling (and therefore a snapshot) is only
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

func emptyIfNil(s []any) []any {
	if s == nil {
		return []any{}
	}
	return s
}
