package user

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/security"
)

// Service holds user business logic over the DB, mirroring the Kotlin UserService.
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// ---- internal finders (return the entity, used by auth) ----

// FindByEmail returns the user with the given (already-normalized) email, or nil if none exists.
func (s *Service) FindByEmail(email string) (*User, error) {
	var u User
	err := s.db.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID returns the user with the given id, or nil if none exists.
func (s *Service) FindByID(id uuid.UUID) (*User, error) {
	var u User
	err := s.db.First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ---- admin operations ----

func (s *Service) List() ([]UserResponse, error) {
	return s.query(s.db)
}

func (s *Service) ListByRole(role Role) ([]UserResponse, error) {
	return s.query(s.db.Where("role = ?", role))
}

func (s *Service) query(tx *gorm.DB) ([]UserResponse, error) {
	var users []User
	if err := tx.Order("created_at").Find(&users).Error; err != nil {
		return nil, err
	}
	out := make([]UserResponse, 0, len(users))
	for i := range users {
		out = append(out, users[i].ToResponse())
	}
	return out, nil
}

func (s *Service) Create(req CreateUserRequest) (*UserResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var count int64
	if err := s.db.Model(&User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, common.NewAppError(http.StatusConflict, "A user with email '"+email+"' already exists")
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	u := User{
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

func (s *Service) Update(id uuid.UUID, req UpdateUserRequest) (*UserResponse, error) {
	u, err := s.FindByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, common.NewAppError(http.StatusNotFound, "User not found")
	}
	if req.FullName != nil {
		u.FullName = strings.TrimSpace(*req.FullName)
	}
	if req.Role != nil {
		u.Role = *req.Role
	}
	if req.Active != nil {
		u.Active = *req.Active
	}
	if err := s.db.Save(u).Error; err != nil {
		return nil, err
	}
	resp := u.ToResponse()
	return &resp, nil
}
