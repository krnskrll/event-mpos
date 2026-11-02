package iface

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// TokenResponse: accessToken is a JWT; refreshToken is opaque and rotated on every refresh.
type TokenResponse struct {
	AccessToken      string      `json:"accessToken"`
	RefreshToken     string      `json:"refreshToken"`
	ExpiresInSeconds int64       `json:"expiresInSeconds"`
	User             UserSummary `json:"user"`
}

type UserSummary struct {
	ID       uuid.UUID      `json:"id"`
	Email    string         `json:"email"`
	FullName string         `json:"fullName"`
	Role     useriface.Role `json:"role"`
}

func SummaryOf(u *useriface.Account) UserSummary {
	return UserSummary{ID: u.ID, Email: u.Email, FullName: u.FullName, Role: u.Role}
}

// Principal is the authenticated user attached to the request context.
type Principal struct {
	UserID   uuid.UUID
	Email    string
	FullName string
	Role     useriface.Role
}

const principalKey = "authPrincipal"

func SetPrincipal(c *gin.Context, p Principal) { c.Set(principalKey, p) }

func PrincipalFrom(c *gin.Context) (Principal, bool) {
	v, exists := c.Get(principalKey)
	if !exists {
		return Principal{}, false
	}
	p, ok := v.(Principal)
	return p, ok
}

// TokenVerifier checks an access token; ok=false when invalid, expired or tampered.
type TokenVerifier interface {
	Parse(token string) (jwt.MapClaims, bool)
}
