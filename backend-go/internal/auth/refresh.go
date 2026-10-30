package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshService manages opaque refresh tokens with rotation + revocation. The raw token is
// returned to the client exactly once; only its hash is persisted.
type RefreshService struct {
	db  *gorm.DB
	ttl time.Duration
}

func NewRefreshService(db *gorm.DB, ttlDays int) *RefreshService {
	return &RefreshService{db: db, ttl: time.Duration(ttlDays) * 24 * time.Hour}
}

// Rotation is the outcome of rotating a refresh token.
type Rotation struct {
	UserID          uuid.UUID
	NewRefreshToken string
}

// Issue mints a new refresh token for the user, stores its hash, and returns the raw value.
func (s *RefreshService) Issue(userID uuid.UUID) (string, error) {
	return s.issue(s.db, userID)
}

func (s *RefreshService) issue(tx *gorm.DB, userID uuid.UUID) (string, error) {
	raw, err := newRawToken()
	if err != nil {
		return "", err
	}
	token := RefreshToken{
		UserID:    userID,
		TokenHash: hashToken(raw),
		ExpiresAt: time.Now().UTC().Add(s.ttl),
	}
	if err := tx.Create(&token).Error; err != nil {
		return "", err
	}
	return raw, nil
}

// Rotate validates a presented token and rotates it: the old one is revoked and a new one issued.
// Returns nil if the token is unknown, revoked, or expired.
func (s *RefreshService) Rotate(raw string) (*Rotation, error) {
	var result *Rotation
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var existing RefreshToken
		err := tx.Where("token_hash = ?", hashToken(raw)).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if existing.Revoked || existing.ExpiresAt.Before(time.Now()) {
			return nil
		}
		if err := tx.Model(&existing).Update("revoked", true).Error; err != nil {
			return err
		}
		newRaw, err := s.issue(tx, existing.UserID)
		if err != nil {
			return err
		}
		result = &Rotation{UserID: existing.UserID, NewRefreshToken: newRaw}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RevokeAllForUser revokes every active refresh token owned by the user (used on logout).
func (s *RefreshService) RevokeAllForUser(userID uuid.UUID) error {
	return s.db.Model(&RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}

func newRawToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.StdEncoding.EncodeToString(sum[:])
}
