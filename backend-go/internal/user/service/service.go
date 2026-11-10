package service

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/database"
	"github.com/krnskrll/event-mpos/internal/platform/security"
	"github.com/krnskrll/event-mpos/internal/user/iface"
	"github.com/krnskrll/event-mpos/internal/user/repository"
)

type Service struct {
	db *gorm.DB
}

var _ iface.Users = (*Service)(nil)

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// FindByEmail expects an already-normalized email.
func (s *Service) FindByEmail(email string) (*iface.Account, error) {
	return s.account("email = ?", email)
}

func (s *Service) FindByID(id uuid.UUID) (*iface.Account, error) {
	return s.account("id = ?", id)
}

// NameOf reads through the caller's tx: a caller holding row locks must not need a second pool connection.
func (s *Service) NameOf(tx database.Tx, id uuid.UUID) (string, error) {
	var u repository.User
	err := tx.Select("full_name").Where("id = ?", id).Take(&u).Error
	return u.FullName, err
}

func (s *Service) account(where string, arg any) (*iface.Account, error) {
	u, err := s.find(where, arg)
	if u == nil {
		return nil, err
	}
	return u.ToAccount(), nil
}

func (s *Service) find(where string, args ...any) (*repository.User, error) {
	var u repository.User
	err := s.db.Where(where, args...).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Service) Create(storeID uuid.UUID, req iface.CreateUserRequest) (*iface.UserResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var count int64
	if err := s.db.Model(&repository.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, common.NewAppError(http.StatusConflict, "A user with email '"+email+"' already exists")
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	u := repository.User{
		StoreID:      storeID,
		Email:        email,
		PasswordHash: hash,
		FullName:     strings.TrimSpace(req.FullName),
		Role:         req.Role,
		Active:       true,
	}
	if err := s.db.Create(&u).Error; err != nil {
		return nil, err
	}
	resp := u.ToResponse()
	return &resp, nil
}
