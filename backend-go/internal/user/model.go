package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role is one of the three system roles. Stored as a string (matches Kotlin @Enumerated(STRING)),
// serialized as e.g. "ADMIN".
type Role string

const (
	RoleAdmin     Role = "ADMIN"
	RoleBA        Role = "BA"
	RoleWarehouse Role = "WAREHOUSE"
)

// ParseRole validates a raw role string.
func ParseRole(s string) (Role, bool) {
	switch Role(s) {
	case RoleAdmin, RoleBA, RoleWarehouse:
		return Role(s), true
	}
	return "", false
}

// User maps to the "users" table.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	FullName     string    `gorm:"column:full_name;not null" json:"fullName"`
	Role         Role      `gorm:"type:varchar(20);not null" json:"role"`
	Active       bool      `gorm:"not null;default:true" json:"active"`
	CreatedAt    time.Time `gorm:"column:created_at;not null" json:"createdAt"`
}

func (User) TableName() string { return "users" }

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	return nil
}

// ---- request/response DTOs ----

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"fullName" binding:"required"`
	Role     Role   `json:"role" binding:"required"`
}

// UpdateUserRequest is a partial update — only non-nil fields are applied.
type UpdateUserRequest struct {
	FullName *string `json:"fullName"`
	Role     *Role   `json:"role"`
	Active   *bool   `json:"active"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	Role      Role      `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      u.Role,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
	}
}
