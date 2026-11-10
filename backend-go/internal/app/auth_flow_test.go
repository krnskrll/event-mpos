package app_test

import (
	"net/http"
	"testing"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

// Ported from the Kotlin AuthFlowTest: login → protected route → refresh rotation → role boundaries.

func login(t *testing.T, a *testutil.TestApp, email, password string) authiface.TokenResponse {
	t.Helper()
	rec := a.Do(http.MethodPost, "/api/auth/login", map[string]string{"email": email, "password": password}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s: %d %s", email, rec.Code, rec.Body.String())
	}
	return testutil.Decode[authiface.TokenResponse](t, rec)
}

func TestLoginReturnsTokensAndProtectsRoutes(t *testing.T) {
	a := testutil.NewApp(t)
	admin := a.CreateUser(useriface.RoleAdmin, "Admin123!")
	tokens := login(t, a, admin.Email, "Admin123!")

	if tokens.User.Role != useriface.RoleAdmin {
		t.Errorf("role = %v, want ADMIN", tokens.User.Role)
	}
	if rec := a.Do(http.MethodGet, "/api/admin/events", nil, tokens.AccessToken); rec.Code != http.StatusOK {
		t.Fatalf("admin route with token: %d", rec.Code)
	}
	if rec := a.Do(http.MethodGet, "/api/admin/events", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("admin route without token: %d, want 401", rec.Code)
	}
}

func TestWrongPasswordIsRejected(t *testing.T) {
	a := testutil.NewApp(t)
	admin := a.CreateUser(useriface.RoleAdmin, "Admin123!")
	rec := a.Do(http.MethodPost, "/api/auth/login", map[string]string{"email": admin.Email, "password": "nope-nope"}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: %d, want 401", rec.Code)
	}
}

func TestRefreshRotatesAndRevokedTokenIsRejected(t *testing.T) {
	a := testutil.NewApp(t)
	admin := a.CreateUser(useriface.RoleAdmin, "Admin123!")
	tokens := login(t, a, admin.Email, "Admin123!")

	rec := a.Do(http.MethodPost, "/api/auth/refresh", map[string]string{"refreshToken": tokens.RefreshToken}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh: %d %s", rec.Code, rec.Body.String())
	}
	refreshed := testutil.Decode[authiface.TokenResponse](t, rec)
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

// BA and WAREHOUSE accounts are retired: cashiers work on paired devices, so only ADMIN signs in.
func TestOnlyAdminsLogIn(t *testing.T) {
	a := testutil.NewApp(t)
	for _, role := range []string{"BA", "WAREHOUSE"} {
		u := a.CreateUser(useriface.RoleAdmin, "Passw0rd!")
		a.DB.Exec("UPDATE users SET role = ? WHERE id = ?", role, u.ID)
		if rec := a.Do(http.MethodPost, "/api/auth/login", map[string]string{"email": u.Email, "password": "Passw0rd!"}, ""); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s login: %d, want 401", role, rec.Code)
		}
	}
	admin := a.CreateUser(useriface.RoleAdmin, "Passw0rd!")
	tokens := login(t, a, admin.Email, "Passw0rd!")
	a.DB.Exec("UPDATE users SET role = 'BA' WHERE id = ?", admin.ID)
	if rec := a.Do(http.MethodPost, "/api/auth/refresh", map[string]string{"refreshToken": tokens.RefreshToken}, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("refresh after the role became BA: %d, want 401", rec.Code)
	}
}
