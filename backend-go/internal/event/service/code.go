package service

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/event/repository"
	"github.com/krnskrll/event-mpos/internal/platform/common"
)

var (
	skipWords = map[string]bool{"outlet": true, "event": true, "pop-up": true, "popup": true, "bazaar": true, "toko": true}
	notAlnum  = regexp.MustCompile(`[^A-Z0-9]`)
)

// CreateLocation creates a draft event with a generated code. The outlet comes with the Cabang
// (EnsureOutlet), so it cannot be added here.
func (s *Service) CreateLocation(storeID uuid.UUID, req iface.CreateLocationRequest) (*iface.EventResponse, error) {
	if req.Type == iface.TypeOutlet {
		return nil, common.NewAppError(http.StatusConflict, "each Cabang has one outlet, created with it")
	}
	if req.Type == iface.TypeEvent && (req.StartDate == nil || req.EndDate == nil) {
		return nil, common.NewAppError(http.StatusBadRequest, "startDate, endDate: required for an event")
	}
	if req.Type == iface.TypeEvent && req.EndDate.Before(req.StartDate.Time) {
		return nil, common.NewAppError(http.StatusBadRequest, "endDate: must not be before startDate")
	}
	code, err := s.nextCode(storeID, req)
	if err != nil {
		return nil, err
	}
	return s.CreateEvent(storeID, iface.CreateEventRequest{Code: code, Name: req.Name, Type: req.Type,
		Address: req.Address, StartDate: req.StartDate, EndDate: req.EndDate})
}

// EnsureOutlet gives the store its one open outlet, a draft named and addressed after the Cabang, when
// it has none; an outlet without an address takes the Cabang's.
func (s *Service) EnsureOutlet(storeID uuid.UUID, name, address string) error {
	var n int64
	if err := s.db.Model(&repository.Event{}).Where("store_id = ? AND type = ? AND status <> ?",
		storeID, iface.TypeOutlet, iface.StatusClosed).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		if address == "" {
			return nil
		}
		return s.db.Model(&repository.Event{}).Where("store_id = ? AND type = ? AND status <> ? AND COALESCE(address, '') = ''",
			storeID, iface.TypeOutlet, iface.StatusClosed).Update("address", address).Error
	}
	code, err := s.nextCode(storeID, iface.CreateLocationRequest{Type: iface.TypeOutlet, Name: name})
	if err != nil {
		return err
	}
	e := repository.Event{StoreID: storeID, Type: iface.TypeOutlet, Code: code, Name: strings.TrimSpace(name), Address: nonEmpty(address), Status: iface.StatusDraft}
	// A concurrent visit may have created it: the open-outlet index makes this a no-op.
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&e).Error
}

// Outlet is the store's open outlet.
func (s *Service) Outlet(storeID uuid.UUID) (*iface.EventResponse, error) {
	var e repository.Event
	err := s.db.Where("store_id = ? AND type = ? AND status <> ?", storeID, iface.TypeOutlet, iface.StatusClosed).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAppError(http.StatusNotFound, "this Cabang has no outlet")
	}
	if err != nil {
		return nil, err
	}
	resp := e.ToResponse()
	return &resp, nil
}

// nextCode is OUT- or EV- plus the initials of the name (event: plus the start year), with -2, -3…
// when taken: "Outlet Kelapa Gading" → OUT-KG, "Pop-up Senayan City" in 2027 → EV-SC27.
func (s *Service) nextCode(storeID uuid.UUID, req iface.CreateLocationRequest) (string, error) {
	initials := ""
	for _, w := range strings.Fields(req.Name) {
		if a := notAlnum.ReplaceAllString(strings.ToUpper(w), ""); a != "" && !skipWords[strings.ToLower(w)] && len(initials) < 4 {
			initials += a[:1]
		}
	}
	if initials == "" {
		initials = "X"
	}
	base := "OUT-" + initials
	if req.Type == iface.TypeEvent {
		base = "EV-" + initials + strconv.Itoa(req.StartDate.Year()%100)
	}
	var taken []string
	if err := s.db.Model(&repository.Event{}).Where("store_id = ? AND (code = ? OR code LIKE ?)", storeID, base, base+"-%").
		Pluck("code", &taken).Error; err != nil {
		return "", err
	}
	used := map[string]bool{}
	for _, c := range taken {
		used[c] = true
	}
	code := base
	for n := 2; used[code]; n++ {
		code = base + "-" + strconv.Itoa(n)
	}
	return code, nil
}

func nonEmpty(s string) *string {
	if s = strings.TrimSpace(s); s == "" {
		return nil
	}
	return &s
}
