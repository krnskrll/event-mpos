package app_test

import (
	"net/http"
	"testing"

	"github.com/krnskrll/event-mpos/internal/auth"
	"github.com/krnskrll/event-mpos/internal/testutil"
	"github.com/krnskrll/event-mpos/internal/user"
)

// Ported from the Kotlin AuthFlowTest: login → protected route → refresh rotation → role boundaries.

func login(t *testing.T, a *testutil.TestApp, email, password string) auth.TokenResponse {
	t.Helper()
	rec := a.Do(http.MethodPost, "/api/auth/login", map[string]string{"email": email, "password": password}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s: %d %s", email, rec.Code, rec.Body.String())
	}
	return testutil.Decode[auth.TokenResponse](t, rec)
}

func TestLoginReturnsTokensAndProtectsRoutes(t *testing.T) {
	a := testutil.NewApp(t)
	admin := a.CreateUser(user.RoleAdmin, "Admin123!")
	tokens := login(t, a, admin.Email, "Admin123!")

	rec := a.Do(http.MethodGet, "/api/me", nil, tokens.AccessToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("/api/me with token: %d", rec.Code)
	}
	if me := testutil.Decode[map[string]any](t, rec); me["role"] != "ADMIN" {
		t.Errorf("role = %v, want ADMIN", me["role"])
	}
	if rec := a.Do(http.MethodGet, "/api/me", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("/api/me without token: %d, want 401", rec.Code)
	}
}

func TestWrongPasswordIsRejected(t *testing.T) {
	a := testutil.NewApp(t)
	admin := a.CreateUser(user.RoleAdmin, "Admin123!")
	rec := a.Do(http.MethodPost, "/api/auth/login", map[string]string{"email": admin.Email, "password": "nope-nope"}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: %d, want 401", rec.Code)
	}
}

func TestRefreshRotatesAndRevokedTokenIsRejected(t *testing.T) {
	a := testutil.NewApp(t)
	admin := a.CreateUser(user.RoleAdmin, "Admin123!")
	tokens := login(t, a, admin.Email, "Admin123!")

	rec := a.Do(http.MethodPost, "/api/auth/refresh", map[string]string{"refreshToken": tokens.RefreshToken}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh: %d %s", rec.Code, rec.Body.String())
	}
	refreshed := testutil.Decode[auth.TokenResponse](t, rec)
	if refreshed.RefreshToken == tokens.RefreshToken {
		t.Error("refresh token was not rotated")
	}

	// The rotated-out token is revoked.
	rec = a.Do(http.MethodPost, "/api/auth/refresh", map[string]string{"refreshToken": tokens.RefreshToken}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("reused refresh token: %d, want 401", rec.Code)
	}

	// Logout revokes every refresh token the user holds.
	if rec := a.Do(http.MethodPost, "/api/auth/logout", nil, refreshed.AccessToken); rec.Code != http.StatusOK {
		t.Fatalf("logout: %d", rec.Code)
	}
	rec = a.Do(http.MethodPost, "/api/auth/refresh", map[string]string{"refreshToken": refreshed.RefreshToken}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("refresh after logout: %d, want 401", rec.Code)
	}
}

func TestBACannotReachAdminRoutes(t *testing.T) {
	a := testutil.NewApp(t)
	ba := a.CreateUser(user.RoleBA, "Ba123456!")
	tokens := login(t, a, ba.Email, "Ba123456!")
	if rec := a.Do(http.MethodGet, "/api/admin/users", nil, tokens.AccessToken); rec.Code != http.StatusForbidden {
		t.Errorf("BA on admin route: %d, want 403", rec.Code)
	}
}
