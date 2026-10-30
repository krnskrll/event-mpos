package auth

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/krnskrll/event-mpos/internal/user"
)

// JwtService issues and verifies short-lived stateless access tokens (HS256). The secret is
// Base64-decoded to raw key bytes — identical to the Kotlin Keys.hmacShaKeyFor(BASE64.decode(secret)),
// so tokens are byte-compatible across the two backends.
type JwtService struct {
	key       []byte
	accessTTL time.Duration
}

func NewJwtService(secretBase64 string, accessTTLMinutes int) (*JwtService, error) {
	key, err := base64.StdEncoding.DecodeString(secretBase64)
	if err != nil {
		return nil, fmt.Errorf("jwt secret must be valid base64: %w", err)
	}
	return &JwtService{key: key, accessTTL: time.Duration(accessTTLMinutes) * time.Minute}, nil
}

func (j *JwtService) GenerateAccessToken(u *user.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   u.ID.String(),
		"email": u.Email,
		"role":  string(u.Role),
		"name":  u.FullName,
		"iat":   now.Unix(),
		"exp":   now.Add(j.accessTTL).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.key)
}

// Parse returns the token claims, or ok=false if invalid/expired/tampered.
func (j *JwtService) Parse(token string) (jwt.MapClaims, bool) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.key, nil
	})
	if err != nil || !parsed.Valid {
		return nil, false
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	return claims, ok
}

func (j *JwtService) AccessTTLSeconds() int64 { return int64(j.accessTTL.Seconds()) }
