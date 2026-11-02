package main

import (
	"log"

	"github.com/krnskrll/event-mpos/internal/app"
	"github.com/krnskrll/event-mpos/internal/platform/config"
	"github.com/krnskrll/event-mpos/internal/platform/database"
	"github.com/krnskrll/event-mpos/internal/seed"
)

func main() {
	cfg := config.Load()

	// 1. Database + schema
	db, err := database.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if err := database.RunMigrations(cfg.DBUrl, cfg.MigrationsPath); err != nil {
		log.Fatalf("golang-migrate failed: %v", err)
	}
	log.Println("database connected + migrations applied")

	// 2. Services + router
	a, err := app.New(cfg, db)
	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}

	// 3. Seed dev data
	if cfg.SeedEnabled {
		if err := seed.Run(db, a.ERPSync, a.MemberSync, a.Stock, a.Promotions); err != nil {
			log.Fatalf("seed failed: %v", err)
		}
	}

	// 4. Start the dynamic sync scheduler + serve
	a.Scheduler.Start()

	log.Printf("listening on :%s", cfg.Port)
	if err := a.Router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
