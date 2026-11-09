package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
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

	if len(os.Args) > 1 && os.Args[1] == "create-admin" {
		createAdmin(a, os.Args[2:])
		return
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
	syncCtx, stopSync := context.WithCancel(context.Background())
	if cfg.BhoomyMode == "http" {
		go syncCatalog(syncCtx, a, time.Duration(cfg.BhoomySyncMinutes)*time.Minute)
	}

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
	stopSync()
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

// syncCatalog pulls Produk Toko changes for every linked store now and then every interval.
func syncCatalog(ctx context.Context, a *app.App, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		a.CatalogSync.RunAll(ctx)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// createAdmin: server create-admin -bhoomy-store <id> -email <email> -name <name>, password on stdin.
func createAdmin(a *app.App, args []string) {
	fs := flag.NewFlagSet("create-admin", flag.ExitOnError)
	store := fs.String("bhoomy-store", "", "Bhoomy store id")
	email := fs.String("email", "", "login email")
	name := fs.String("name", "", "full name")
	_ = fs.Parse(args)
	id, err := uuid.Parse(*store)
	if err != nil {
		log.Fatalf("create-admin: -bhoomy-store: %v", err)
	}
	password, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	u, err := a.CreateStoreAdmin(id, *email, *name, strings.TrimRight(password, "\r\n"))
	if err != nil {
		log.Fatalf("create-admin: %v", err)
	}
	log.Printf("admin %s created in POS store %s", u.Email, u.StoreID)
}
