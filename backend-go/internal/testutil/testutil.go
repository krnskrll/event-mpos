// Package testutil gives integration tests a migrated Postgres database and a fully wired App.
//
// Tests that need the database call NewApp(t). It reads TEST_DB_URL; when that is unset the test is
// skipped, so `go test ./...` still works on a machine without Postgres. CI sets REQUIRE_TEST_DB=1,
// which turns a missing TEST_DB_URL into a failure instead of a silent skip.
//
// Every test shares one database, so tests must create their own uniquely-named fixtures (event
// codes, SKUs, emails, idempotency keys) rather than assume an empty schema.
package testutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/app"
	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/config"
	"github.com/krnskrll/event-mpos/internal/database"
	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/product"
	"github.com/krnskrll/event-mpos/internal/user"
)

var migrateOnce sync.Once
var migrateErr error

// RepoPath resolves a path relative to the repository root (the directory above backend-go/).
func RepoPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "..", "..", "..")
	return filepath.Join(append([]string{root}, parts...)...)
}

// DB returns a connection to the migrated test database, or skips the test when none is configured.
func DB(t *testing.T) *gorm.DB {
	t.Helper()
	url := os.Getenv("TEST_DB_URL")
	if url == "" {
		if os.Getenv("REQUIRE_TEST_DB") != "" {
			t.Fatal("REQUIRE_TEST_DB is set but TEST_DB_URL is empty")
		}
		t.Skip("TEST_DB_URL not set — skipping Postgres integration test")
	}
	migrateOnce.Do(func() {
		migrateErr = database.RunMigrations(url, RepoPath("backend-go", "migrations"))
	})
	if migrateErr != nil {
		t.Fatalf("migrations: %v", migrateErr)
	}
	db, err := database.Connect(url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return db
}

// TestApp is a wired App plus its database handle.
type TestApp struct {
	*app.App
	DB *gorm.DB
	t  *testing.T
}

// NewApp builds the production App (mock ERP/member sources, scheduler not started) on the test DB.
func NewApp(t *testing.T) *TestApp {
	t.Helper()
	db := DB(t)
	cfg := &config.Config{
		JWTSecret:           "ZHYtbG9jYWwtZGV2LXNlY3JldC1jaGFuZ2UtbWUtaW4tcHJvZHVjdGlvbi0xMjM0",
		AccessTokenTTLMin:   15,
		RefreshTokenTTLDays: 7,
		ERPSource:           "mock",
		MemberSource:        "mock",
		ERPSyncCron:         "0 0 * * * *",
		MemberSyncCron:      "0 0 2 * * *",
	}
	a, err := app.New(cfg, db)
	if err != nil {
		t.Fatalf("app: %v", err)
	}
	return &TestApp{App: a, DB: db, t: t}
}

// Unique returns prefix plus a short random suffix, for fixture names that must not collide.
func Unique(prefix string) string {
	return prefix + strings.ToUpper(uuid.NewString()[:8])
}

// CreateUser creates a user with the given role and returns it (email is made unique).
func (a *TestApp) CreateUser(role user.Role, password string) *user.UserResponse {
	a.t.Helper()
	u, err := a.Users.Create(user.CreateUserRequest{
		Email: strings.ToLower(Unique("u-")) + "@test.local", Password: password, FullName: "Test " + string(role), Role: role,
	})
	if err != nil {
		a.t.Fatalf("create user: %v", err)
	}
	return u
}

// ActiveEvent inserts an ACTIVE event with a unique code.
func (a *TestApp) ActiveEvent() *event.Event {
	a.t.Helper()
	ev := &event.Event{Code: Unique("T"), Name: "Test event", Status: event.StatusActive}
	if err := a.DB.Create(ev).Error; err != nil {
		a.t.Fatalf("create event: %v", err)
	}
	return ev
}

// Product inserts a product with the given price and a unique SKU.
func (a *TestApp) Product(price string) *product.Product {
	a.t.Helper()
	p := &product.Product{
		SKU: Unique("SKU-"), Name: "Widget", Price: decimal.RequireFromString(price),
		WarehouseStock: 100000, Source: product.SourceManual, Active: true,
	}
	if err := a.DB.Create(p).Error; err != nil {
		a.t.Fatalf("create product: %v", err)
	}
	return p
}

// Allocate sets an event allocation directly (bypasses the DRAFT-only admin rule).
func (a *TestApp) Allocate(eventID, productID uuid.UUID, qty int) {
	a.t.Helper()
	if err := a.Stock.ForceAllocate(eventID, productID, qty); err != nil {
		a.t.Fatalf("allocate: %v", err)
	}
}

// Remaining returns the remaining event stock for a product.
func (a *TestApp) Remaining(eventID, productID uuid.UUID) int {
	a.t.Helper()
	n, err := a.Stock.Remaining(eventID, productID)
	if err != nil {
		a.t.Fatalf("remaining: %v", err)
	}
	return n
}

// Do sends a JSON request through the real router and returns the recorder.
func (a *TestApp) Do(method, path string, body any, token string) *httptest.ResponseRecorder {
	a.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			a.t.Fatalf("encode: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	a.Router.ServeHTTP(rec, req)
	return rec
}

// Decode unmarshals a recorder body into v, failing the test on error.
func Decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return v
}

// StatusOf returns the HTTP status carried by a service error (common.AppError), or 0.
func StatusOf(err error) int {
	var ae *common.AppError
	if errors.As(err, &ae) {
		return ae.Status
	}
	return 0
}
