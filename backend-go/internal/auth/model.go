package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/user"
)

// ---- request/response DTOs (identical JSON shape to the Kotlin backend) ----

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// TokenResponse is returned by /login and /refresh. accessToken is a JWT; refreshToken is opaque
// and rotated on every refresh; expiresInSeconds is the access-token lifetime.
type TokenResponse struct {
	AccessToken      string      `json:"accessToken"`
	RefreshToken     string      `json:"refreshToken"`
	ExpiresInSeconds int64       `json:"expiresInSeconds"`
	User             UserSummary `json:"user"`
}

type UserSummary struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	FullName string    `json:"fullName"`
	Role     user.Role `json:"role"`
}

func summaryOf(u *user.User) UserSummary {
	return UserSummary{ID: u.ID, Email: u.Email, FullName: u.FullName, Role: u.Role}
}

// RefreshToken persists only a SHA-256 hash of the opaque token — never the raw value — so a DB
// leak cannot be replayed. Tokens are rotated on use and can be revoked.
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;not null;index"`
	TokenHash string    `gorm:"column:token_hash;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
	Revoked   bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

func (t *RefreshToken) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	return nil
}
