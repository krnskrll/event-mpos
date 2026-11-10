package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/krnskrll/event-mpos/internal/device/iface"
	"github.com/krnskrll/event-mpos/internal/device/repository"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/database"
)

var _ iface.Devices = (*Service)(nil)

type Service struct {
	db        *gorm.DB
	events    eventiface.Events
	storeName func(storeID uuid.UUID) string
}

func NewService(db *gorm.DB, events eventiface.Events, storeName func(uuid.UUID) string) *Service {
	return &Service{db: db, events: events, storeName: storeName}
}

// live keeps devices that still work: not removed, or a closed event's removal still ahead.
func live(tx *gorm.DB) *gorm.DB {
	return tx.Where("removed_at IS NULL OR removed_at > ?", time.Now())
}

// NewCode replaces the location's pairing code with a fresh one, unique among live codes.
func (s *Service) NewCode(storeID, locationID uuid.UUID) (*iface.PairingCode, error) {
	if _, err := s.openLocation(storeID, locationID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	pc := repository.PairingCode{StoreID: storeID, LocationID: locationID, ExpiresAt: now.Add(iface.CodeTTL), CreatedAt: now}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("location_id = ? OR expires_at <= ?", locationID, now).Delete(&repository.PairingCode{}).Error; err != nil {
			return err
		}
		for range 20 {
			n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
			if err != nil {
				return err
			}
			pc.Code = fmt.Sprintf("%06d", n.Int64())
			res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&pc)
			if res.Error != nil || res.RowsAffected == 1 {
				return res.Error
			}
		}
		return errors.New("no free pairing code")
	})
	if err != nil {
		return nil, err
	}
	return &iface.PairingCode{Code: pc.Code, ExpiresAt: pc.ExpiresAt}, nil
}

// Pair links a new device to the code's location as the next free "Kasir N".
func (s *Service) Pair(code, model string) (*iface.PairResponse, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	var d repository.Device
	var ev *eventiface.EventResponse
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var pc repository.PairingCode
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code = ? AND expires_at > ?", strings.TrimSpace(code), time.Now()).Take(&pc).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewAppError(http.StatusNotFound, "Pairing code not found or expired")
		}
		if err != nil {
			return err
		}
		if ev, err = s.openLocation(pc.StoreID, pc.LocationID); err != nil {
			return err
		}
		name, err := nextName(tx, pc.LocationID)
		if err != nil {
			return err
		}
		d = repository.Device{StoreID: pc.StoreID, LocationID: pc.LocationID, Name: name, TokenHash: hashToken(token), PairedAt: time.Now().UTC()}
		if m := strings.TrimSpace(model); m != "" {
			d.Model = &m
		}
		return tx.Create(&d).Error
	})
	if err != nil {
		return nil, err
	}
	return &iface.PairResponse{Token: token, DeviceResponse: s.response(&d, ev)}, nil
}

// nextName is the lowest "Kasir N" no live device at the location uses.
func nextName(tx *gorm.DB, locationID uuid.UUID) (string, error) {
	var names []string
	if err := live(tx.Model(&repository.Device{})).Where("location_id = ?", locationID).Pluck("name", &names).Error; err != nil {
		return "", err
	}
	taken := map[string]bool{}
	for _, n := range names {
		taken[n] = true
	}
	for n := 1; ; n++ {
		if name := fmt.Sprintf("Kasir %d", n); !taken[name] {
			return name, nil
		}
	}
}

func (s *Service) Authenticate(token string) (*iface.Device, error) {
	if token == "" {
		return nil, nil
	}
	var d repository.Device
	err := live(s.db).Where("token_hash = ?", hashToken(token)).Take(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if now := time.Now(); d.LastSeenAt == nil || now.Sub(*d.LastSeenAt) >= time.Minute {
		if err := s.db.Model(&d).UpdateColumn("last_seen_at", now).Error; err != nil {
			return nil, err
		}
	}
	return &iface.Device{ID: d.ID, StoreID: d.StoreID, LocationID: d.LocationID, Name: d.Name}, nil
}

// Info is GET /api/device for an authenticated device.
func (s *Service) Info(deviceID uuid.UUID) (*iface.DeviceResponse, error) {
	var d repository.Device
	if err := s.db.Take(&d, "id = ?", deviceID).Error; err != nil {
		return nil, err
	}
	ev, err := s.events.Get(d.LocationID)
	if err != nil {
		return nil, err
	}
	resp := s.response(&d, ev)
	return &resp, nil
}

func (s *Service) List(storeID, locationID uuid.UUID) ([]iface.DeviceView, error) {
	if err := s.events.RequireInStore(storeID, locationID); err != nil {
		return nil, err
	}
	var rows []repository.Device
	if err := live(s.db).Where("location_id = ?", locationID).Order("paired_at").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]iface.DeviceView, 0, len(rows))
	for i := range rows {
		out = append(out, view(&rows[i]))
	}
	return out, nil
}

func (s *Service) Rename(storeID, id uuid.UUID, name string) (*iface.DeviceView, error) {
	d, err := s.findIn(storeID, id)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 40 {
		return nil, common.NewAppError(http.StatusBadRequest, "name: 1 to 40 characters")
	}
	var clash int64
	if err := live(s.db.Model(&repository.Device{})).Where("location_id = ? AND name = ? AND id <> ?", d.LocationID, name, d.ID).Count(&clash).Error; err != nil {
		return nil, err
	}
	if clash > 0 {
		return nil, common.NewAppError(http.StatusConflict, "Another device here is already called "+name)
	}
	if err := s.db.Model(d).Update("name", name).Error; err != nil {
		return nil, err
	}
	v := view(d)
	return &v, nil
}

// Remove unlinks a device now; its token stops working.
func (s *Service) Remove(storeID, id uuid.UUID) error {
	d, err := s.findIn(storeID, id)
	if err != nil {
		return err
	}
	return s.db.Model(d).Update("removed_at", time.Now().UTC()).Error
}

func (s *Service) ScheduleRemoval(tx database.Tx, locationID uuid.UUID, at time.Time) error {
	return tx.Model(&repository.Device{}).Where("location_id = ? AND (removed_at IS NULL OR removed_at > ?)", locationID, at).
		Update("removed_at", at).Error
}

func (s *Service) findIn(storeID, id uuid.UUID) (*repository.Device, error) {
	var d repository.Device
	err := live(s.db).Where("id = ? AND store_id = ?", id, storeID).Take(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAppError(http.StatusNotFound, "Device not found")
	}
	return &d, err
}

// openLocation is a store's location that can still get devices (not closed).
func (s *Service) openLocation(storeID, locationID uuid.UUID) (*eventiface.EventResponse, error) {
	if err := s.events.RequireInStore(storeID, locationID); err != nil {
		return nil, err
	}
	ev, err := s.events.Get(locationID)
	if err != nil {
		return nil, err
	}
	if ev.Status == eventiface.StatusClosed {
		return nil, common.NewAppError(http.StatusConflict, ev.Name+" is closed; devices cannot be added")
	}
	return ev, nil
}

func (s *Service) response(d *repository.Device, ev *eventiface.EventResponse) iface.DeviceResponse {
	return iface.DeviceResponse{
		Device:    iface.DeviceRef{ID: d.ID, Name: d.Name, RemovesAt: d.RemovedAt},
		Location:  iface.Location{ID: ev.ID, Name: ev.Name, Type: ev.Type, Status: ev.Status, StartDate: ev.StartDate, EndDate: ev.EndDate},
		StoreName: s.storeName(d.StoreID),
	}
}

func view(d *repository.Device) iface.DeviceView {
	return iface.DeviceView{ID: d.ID, Name: d.Name, Model: d.Model, PairedAt: d.PairedAt, LastSeenAt: d.LastSeenAt, RemovesAt: d.RemovedAt}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
