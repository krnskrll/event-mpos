package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all runtime settings, sourced from environment variables. It mirrors the Kotlin
// backend's application.yml so the two are behaviourally interchangeable. The single value that
// changes across environments (local Docker → Supabase → AWS RDS) is DBUrl.
type Config struct {
	// Database — a full Postgres URL, e.g. postgres://user:pass@host:5432/db?sslmode=disable.
	// Serves both GORM (app runtime) and golang-migrate (schema), so upscaling is a URL swap.
	DBUrl string

	// Schema migrations directory (golang-migrate). Falls back to GORM AutoMigrate if absent.
	MigrationsPath string

	// HTTP
	Port           string
	CORSOrigins    []string
	SeedEnabled    bool

	// JWT
	JWTSecret            string
	AccessTokenTTLMin    int
	RefreshTokenTTLDays  int

	// ERP adapter: mock | file | rest
	ERPSource          string
	ERPFilePath        string
	ERPRestURL         string
	ERPScheduledSync   bool
	ERPSyncCron        string

	// Member adapter: mock | file | rest
	MemberSource        string
	MemberFilePath      string
	MemberRestURL       string
	MemberRegionFilter  string
	MemberScheduledSync bool
	MemberSyncCron      string
}

// Load reads configuration from the environment, applying the same defaults as the Kotlin backend.
// A .env file is loaded if present (ignored in production where real env vars are set).
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DBUrl: getenv("DB_URL", "postgres://eventmpos:eventmpos@localhost:5432/eventmpos?sslmode=disable"),

		MigrationsPath: getenv("MIGRATIONS_PATH", "migrations"),

		Port:        getenv("PORT", getenv("SERVER_PORT", "8080")),
		CORSOrigins: splitCSV(getenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
		SeedEnabled: getbool("SEED_ENABLED", true),

		JWTSecret:           getenv("JWT_SECRET", "ZHYtbG9jYWwtZGV2LXNlY3JldC1jaGFuZ2UtbWUtaW4tcHJvZHVjdGlvbi0xMjM0"),
		AccessTokenTTLMin:   getint("JWT_ACCESS_TTL", 15),
		RefreshTokenTTLDays: getint("JWT_REFRESH_TTL", 7),

		ERPSource:        getenv("ERP_SOURCE", "mock"),
		ERPFilePath:      getenv("ERP_FILE_PATH", ""),
		ERPRestURL:       getenv("ERP_REST_URL", ""),
		ERPScheduledSync: getbool("ERP_SCHEDULED", false),
		ERPSyncCron:      getenv("ERP_SYNC_CRON", "0 0 * * * *"),

		MemberSource:        getenv("MEMBER_SOURCE", "mock"),
		MemberFilePath:      getenv("MEMBER_FILE_PATH", ""),
		MemberRestURL:       getenv("MEMBER_REST_URL", ""),
		MemberRegionFilter:  getenv("MEMBER_REGION_FILTER", ""),
		MemberScheduledSync: getbool("MEMBER_SCHEDULED", false),
		MemberSyncCron:      getenv("MEMBER_SYNC_CRON", "0 0 2 * * *"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getbool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getint(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
