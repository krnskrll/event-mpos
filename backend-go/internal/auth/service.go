package auth

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/security"
	"github.com/krnskrll/event-mpos/internal/user"
)

// Service coordinates login, token refresh, and logout — the Go analogue of Kotlin AuthService.
type Service struct {
	users   *user.Service
	jwt     *JwtService
	refresh *RefreshService
}

func NewService(users *user.Service, jwt *JwtService, refresh *RefreshService) *Service {
	return &Service{users: users, jwt: jwt, refresh: refresh}
}

func (s *Service) Login(email, password string) (*TokenResponse, error) {
	u, err := s.users.FindByEmail(strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, err
	}
	// Same error for "no such user" and "wrong password"; run a bcrypt compare even when the user is
	// absent so response timing doesn't reveal which emails exist.
	if u == nil {
		security.CheckPassword("$2a$12$0000000000000000000000000000000000000000000000000000", password)
		return nil, common.NewAppError(http.StatusUnauthorized, "Invalid email or password")
	}
	if !security.CheckPassword(u.PasswordHash, password) {
		return nil, common.NewAppError(http.StatusUnauthorized, "Invalid email or password")
	}
	if !u.Active {
		return nil, common.NewAppError(http.StatusForbidden, "Account is disabled")
	}
	return s.issueTokens(u)
}

func (s *Service) Refresh(rawRefreshToken string) (*TokenResponse, error) {
	rotation, err := s.refresh.Rotate(rawRefreshToken)
	if err != nil {
		return nil, err
	}
	if rotation == nil {
		return nil, common.NewAppError(http.StatusUnauthorized, "Invalid or expired refresh token")
	}
	u, err := s.users.FindByID(rotation.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, common.NewAppError(http.StatusUnauthorized, "Invalid or expired refresh token")
	}
	if !u.Active {
		return nil, common.NewAppError(http.StatusForbidden, "Account is disabled")
	}
	access, err := s.jwt.GenerateAccessToken(u)
	if err != nil {
		return nil, err
	}
	return &TokenResponse{
		AccessToken:      access,
		RefreshToken:     rotation.NewRefreshToken,
		ExpiresInSeconds: s.jwt.AccessTTLSeconds(),
		User:             summaryOf(u),
	}, nil
}

func (s *Service) Logout(userID uuid.UUID) error {
	return s.refresh.RevokeAllForUser(userID)
}

func (s *Service) issueTokens(u *user.User) (*TokenResponse, error) {
	access, err := s.jwt.GenerateAccessToken(u)
	if err != nil {
		return nil, err
	}
	refreshRaw, err := s.refresh.Issue(u.ID)
	if err != nil {
		return nil, err
	}
	return &TokenResponse{
		AccessToken:      access,
		RefreshToken:     refreshRaw,
		ExpiresInSeconds: s.jwt.AccessTTLSeconds(),
		User:             summaryOf(u),
	}, nil
}
