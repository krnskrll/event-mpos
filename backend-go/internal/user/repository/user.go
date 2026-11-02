package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/user/iface"
)

type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"column:password_hash;not null" json:"-"`
	FullName     string     `gorm:"column:full_name;not null" json:"fullName"`
	Role         iface.Role `gorm:"type:varchar(20);not null" json:"role"`
	Active       bool       `gorm:"not null;default:true" json:"active"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null" json:"createdAt"`
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

func (u *User) ToResponse() iface.UserResponse {
	return iface.UserResponse{ID: u.ID, Email: u.Email, FullName: u.FullName, Role: u.Role, Active: u.Active, CreatedAt: u.CreatedAt}
}

func (u *User) ToAccount() *iface.Account {
	return &iface.Account{ID: u.ID, Email: u.Email, PasswordHash: u.PasswordHash, FullName: u.FullName, Role: u.Role, Active: u.Active}
}
