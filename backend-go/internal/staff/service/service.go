package service

import (
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"

	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/database"
	"github.com/krnskrll/event-mpos/internal/platform/security"
	"github.com/krnskrll/event-mpos/internal/staff/iface"
	"github.com/krnskrll/event-mpos/internal/staff/repository"
)

var _ iface.Staff = (*Service)(nil)

var pinPattern = regexp.MustCompile(`^[0-9]{4,6}$`)

type Service struct {
	db     *gorm.DB
	events eventiface.Events
}

func NewService(db *gorm.DB, events eventiface.Events) *Service {
	return &Service{db: db, events: events}
}

func (s *Service) List(storeID uuid.UUID) ([]iface.StaffResponse, error) {
	var rows []repository.Staff
	if err := s.db.Where("store_id = ?", storeID).Order("name").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]iface.StaffResponse, 0, len(rows))
	for i := range rows {
		r, err := s.response(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, nil
}

func (s *Service) Get(storeID, id uuid.UUID) (*iface.StaffResponse, error) {
	row, err := s.findIn(storeID, id)
	if err != nil {
		return nil, err
	}
	return s.response(row)
}

func (s *Service) Create(storeID uuid.UUID, req iface.StaffRequest) (*iface.StaffResponse, error) {
	if req.Pin == nil {
		return nil, common.NewAppError(http.StatusBadRequest, "pin: required")
	}
	row := &repository.Staff{StoreID: storeID, Active: req.Active == nil || *req.Active, CreatedAt: time.Now().UTC()}
	return s.save(row, req, nil)
}

func (s *Service) Update(storeID, id uuid.UUID, req iface.StaffRequest) (*iface.StaffResponse, error) {
	row, err := s.findIn(storeID, id)
	if err != nil {
		return nil, err
	}
	if req.Active != nil {
		row.Active = *req.Active
	}
	current, err := s.locationsOf(id)
	if err != nil {
		return nil, err
	}
	return s.save(row, req, current)
}

// save validates the request onto row and replaces its locations. A closed location can stay
// assigned but not be newly added.
func (s *Service) save(row *repository.Staff, req iface.StaffRequest, current []uuid.UUID) (*iface.StaffResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || utf8.RuneCountInString(name) > 40 {
		return nil, common.NewAppError(http.StatusBadRequest, "name: 1 to 40 characters")
	}
	if req.Role != iface.RoleKasir && req.Role != iface.RoleSupervisor {
		return nil, common.NewAppError(http.StatusBadRequest, "role: must be KASIR or SUPERVISOR")
	}
	if req.Pin != nil {
		if !pinPattern.MatchString(*req.Pin) {
			return nil, common.NewAppError(http.StatusBadRequest, "pin: 4 to 6 digits")
		}
		hash, err := security.HashPin(*req.Pin)
		if err != nil {
			return nil, err
		}
		row.PinHash = hash
	}
	var locations []uuid.UUID
	for _, id := range req.LocationIDs {
		if slices.Contains(locations, id) {
			continue
		}
		ev, err := s.events.Get(id)
		if err != nil || ev.StoreID != row.StoreID {
			return nil, common.NewAppError(http.StatusNotFound, "Location not found")
		}
		if ev.Status == eventiface.StatusClosed && !slices.Contains(current, id) {
			return nil, common.NewAppError(http.StatusBadRequest, "locationIds: "+ev.Name+" is closed")
		}
		locations = append(locations, id)
	}
	row.Name, row.Role, row.UpdatedAt = name, req.Role, time.Now().UTC()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(row).Error; err != nil {
			return err
		}
		if err := tx.Where("staff_id = ?", row.ID).Delete(&repository.StaffLocation{}).Error; err != nil {
			return err
		}
		for _, l := range locations {
			if err := tx.Create(&repository.StaffLocation{StaffID: row.ID, LocationID: l}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.response(row)
}

func (s *Service) Find(storeID, id uuid.UUID) (*iface.Person, error) {
	row, err := s.findIn(storeID, id)
	if err != nil {
		if common.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	locations, err := s.locationsOf(id)
	if err != nil {
		return nil, err
	}
	return &iface.Person{ID: row.ID, StoreID: row.StoreID, Name: row.Name, Role: row.Role, Active: row.Active, LocationIDs: locations}, nil
}

func (s *Service) ForLocation(locationID uuid.UUID) ([]iface.DeviceStaff, error) {
	out := []iface.DeviceStaff{}
	err := s.db.Model(&repository.Staff{}).Select("pos_staff.id, pos_staff.name, pos_staff.role, pos_staff.pin_hash").
		Joins("JOIN pos_staff_locations sl ON sl.staff_id = pos_staff.id").
		Where("sl.location_id = ? AND pos_staff.active", locationID).Order("pos_staff.name").Scan(&out).Error
	return out, err
}

func (s *Service) Approve(storeID, locationID, supervisorID uuid.UUID, pin string) (*iface.Person, error) {
	p, err := s.Find(storeID, supervisorID)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Role != iface.RoleSupervisor || !p.WorksAt(locationID) {
		return nil, common.NewAppError(http.StatusForbidden, "Only an active supervisor of this location can approve")
	}
	var hash string
	if err := s.db.Model(&repository.Staff{}).Where("id = ?", p.ID).Pluck("pin_hash", &hash).Error; err != nil {
		return nil, err
	}
	if !security.CheckPassword(hash, pin) {
		return nil, common.NewAppError(http.StatusForbidden, "Wrong supervisor PIN")
	}
	return p, nil
}

// NameOf reads through the caller's tx: a caller holding row locks must not need a second pool connection.
func (s *Service) NameOf(tx database.Tx, id uuid.UUID) (string, error) {
	var row repository.Staff
	err := tx.Select("name").Where("id = ?", id).Take(&row).Error
	return row.Name, err
}

func (s *Service) findIn(storeID, id uuid.UUID) (*repository.Staff, error) {
	var row repository.Staff
	err := s.db.Where("id = ? AND store_id = ?", id, storeID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAppError(http.StatusNotFound, "Staff not found")
	}
	return &row, err
}

func (s *Service) locationsOf(staffID uuid.UUID) ([]uuid.UUID, error) {
	ids := []uuid.UUID{}
	err := s.db.Model(&repository.StaffLocation{}).Where("staff_id = ?", staffID).Order("location_id").Pluck("location_id", &ids).Error
	return ids, err
}

func (s *Service) response(row *repository.Staff) (*iface.StaffResponse, error) {
	locations, err := s.locationsOf(row.ID)
	if err != nil {
		return nil, err
	}
	return &iface.StaffResponse{ID: row.ID, Name: row.Name, Role: row.Role, LocationIDs: locations, Active: row.Active,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}
