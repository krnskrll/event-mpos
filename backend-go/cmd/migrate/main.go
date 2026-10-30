// Command migrate applies the golang-migrate files to DB_URL and exits. The server does the same on
// startup; this exists for CI and for migrating a database before a deploy.
package main

import (
	"log"

	"github.com/krnskrll/event-mpos/internal/config"
	"github.com/krnskrll/event-mpos/internal/database"
)

func main() {
	cfg := config.Load()
	if err := database.RunMigrations(cfg.DBUrl, cfg.MigrationsPath); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}
	log.Printf("migrations applied from %s", cfg.MigrationsPath)
}
