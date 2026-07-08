package auth

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/user"
)

const testSecretB64 = "ZHYtbG9jYWwtZGV2LXNlY3JldC1jaGFuZ2UtbWUtaW4tcHJvZHVjdGlvbi0xMjM0"

func TestJwtRoundTrip(t *testing.T) {
	j, err := NewJwtService(testSecretB64, 15)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	u := &user.User{ID: uuid.New(), Email: "a@b.com", FullName: "Ann Bee", Role: user.RoleAdmin}
	tok, err := j.GenerateAccessToken(u)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, ok := j.Parse(tok)
	if !ok {
		t.Fatal("parse failed on a freshly-issued token")
	}
	if claims["sub"] != u.ID.String() {
		t.Errorf("sub = %v, want %v", claims["sub"], u.ID)
	}
	if claims["role"] != "ADMIN" {
		t.Errorf("role = %v, want ADMIN", claims["role"])
	}
	if claims["email"] != "a@b.com" {
		t.Errorf("email = %v", claims["email"])
	}
	if j.AccessTTLSeconds() != 900 {
		t.Errorf("ttl seconds = %d, want 900", j.AccessTTLSeconds())
	}
}

func TestJwtRejectsTampered(t *testing.T) {
	j, _ := NewJwtService(testSecretB64, 15)
	if _, ok := j.Parse("not.a.jwt"); ok {
		t.Error("garbage token accepted")
	}
	other, _ := NewJwtService("YW5vdGhlci1zZWNyZXQtdGhhdC1pcy1sb25nLWVub3VnaC0zMmJ5dGVzIQ==", 15)
	u := &user.User{ID: uuid.New(), Email: "x@y.com", Role: user.RoleBA}
	tok, _ := other.GenerateAccessToken(u)
	if _, ok := j.Parse(tok); ok {
		t.Error("token signed with a different key was accepted")
	}
}

func TestRefreshTokenHashing(t *testing.T) {
	raw, err := newRawToken()
	if err != nil {
		t.Fatalf("newRawToken: %v", err)
	}
	if len(raw) < 40 {
		t.Errorf("raw token too short: %d chars", len(raw))
	}
	// Hash is deterministic for a given input, and the raw value is not recoverable from it.
	if hashToken(raw) != hashToken(raw) {
		t.Error("hashToken not deterministic")
	}
	if hashToken(raw) == raw {
		t.Error("stored hash equals raw token")
	}
	other, _ := newRawToken()
	if raw == other {
		t.Error("two generated tokens collided")
	}
}
