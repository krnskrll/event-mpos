package service

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/receipt/iface"
	"github.com/krnskrll/event-mpos/internal/receipt/repository"
)

// Note fits two 58 mm lines; contact fields fit one.
const maxPhone, maxInstagram, maxNote = 20, 30, 64

// Settings serves Pengaturan Struk: read and saved in Bhoomy, sent to devices with their pairing data.
type Settings struct{ db *gorm.DB }

func NewSettings(db *gorm.DB) *Settings { return &Settings{db: db} }

func (s *Settings) Get(storeID uuid.UUID) (iface.Settings, error) {
	var r repository.StoreSettings
	if err := s.db.Take(&r, "id = ?", storeID).Error; err != nil {
		return iface.Settings{}, err
	}
	return iface.Settings{Phone: r.Phone, Instagram: r.Instagram, Note: r.Note, ShowAddress: r.ShowAddress, ShowStaff: r.ShowStaff, ShowCustomer: r.ShowCustomer}, nil
}

// Save trims the text (Instagram without "@") and answers 400 when a field is too long for the receipt.
func (s *Settings) Save(storeID uuid.UUID, in iface.Settings) (iface.Settings, error) {
	in.Phone, in.Note = strings.TrimSpace(in.Phone), strings.TrimSpace(in.Note)
	in.Instagram = strings.TrimPrefix(strings.TrimSpace(in.Instagram), "@")
	for _, f := range []struct {
		name, v string
		max     int
	}{{"phone", in.Phone, maxPhone}, {"instagram", in.Instagram, maxInstagram}, {"note", in.Note, maxNote}} {
		if utf8.RuneCountInString(f.v) > f.max {
			return iface.Settings{}, common.NewAppError(http.StatusBadRequest, f.name+" is longer than the receipt allows")
		}
	}
	err := s.db.Model(&repository.StoreSettings{}).Where("id = ?", storeID).Updates(map[string]any{
		"receipt_phone": in.Phone, "receipt_instagram": in.Instagram, "receipt_note": in.Note,
		"receipt_show_address": in.ShowAddress, "receipt_show_staff": in.ShowStaff, "receipt_show_customer": in.ShowCustomer,
	}).Error
	if err != nil {
		return iface.Settings{}, err
	}
	return s.Get(storeID)
}

// ForDevice is Get for the device's pairing data; a read error prints the defaults.
func (s *Settings) ForDevice(storeID uuid.UUID) iface.Settings {
	if v, err := s.Get(storeID); err == nil {
		return v
	}
	return iface.Settings{ShowAddress: true, ShowStaff: true, ShowCustomer: true}
}
