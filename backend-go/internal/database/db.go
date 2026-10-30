package database

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a GORM connection to Postgres and verifies it with a ping. The URL is a standard
// postgres:// DSN, identical for local Docker, Supabase, or AWS RDS.
func Connect(dbURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		// Surface gorm.ErrDuplicatedKey / ErrForeignKeyViolated so services can detect e.g. the
		// idempotency-key race on concurrent order submits.
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// Conservative pool defaults; safe for Supabase's pooler and RDS alike.
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
