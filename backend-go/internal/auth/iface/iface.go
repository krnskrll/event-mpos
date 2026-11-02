package iface

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/platform/tenant"
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
	StoreID  uuid.UUID      `json:"storeId"`
	Email    string         `json:"email"`
	FullName string         `json:"fullName"`
	Role     useriface.Role `json:"role"`
}

func SummaryOf(u *useriface.Account) UserSummary {
	return UserSummary{ID: u.ID, StoreID: u.StoreID, Email: u.Email, FullName: u.FullName, Role: u.Role}
}

// Principal is the authenticated user attached to the request context.
type Principal struct {
	UserID   uuid.UUID
	StoreID  uuid.UUID
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

// StoreOf is the caller's store; routes behind JWTMiddleware always have one.
func StoreOf(c *gin.Context) uuid.UUID {
	p, _ := PrincipalFrom(c)
	return p.StoreID
}

// StoreFromClaims reads the store claim; tokens issued before stores existed get the default store.
func StoreFromClaims(claims jwt.MapClaims) uuid.UUID {
	raw, _ := claims["store_id"].(string)
	if id, err := uuid.Parse(raw); err == nil {
		return id
	}
	return tenant.DefaultStoreID
}

// TokenVerifier checks an access token; ok=false when invalid, expired or tampered.
type TokenVerifier interface {
	Parse(token string) (jwt.MapClaims, bool)
}
