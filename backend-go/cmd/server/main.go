package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	go pingBhoomy(a)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: a.Router}
	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// 5. On SIGTERM/SIGINT: stop scheduled syncs, then let in-flight requests finish.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down")
	a.Scheduler.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// pingBhoomy only logs: POS keeps serving whether or not Bhoomy is reachable.
func pingBhoomy(a *app.App) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.Bhoomy.Ping(ctx); err != nil {
		log.Printf("bhoomy: unreachable: %v", err)
		return
	}
	log.Println("bhoomy: reachable")
}
