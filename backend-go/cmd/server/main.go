package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/krnskrll/event-mpos/internal/auth"
	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/config"
	"github.com/krnskrll/event-mpos/internal/database"
	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/me"
	"github.com/krnskrll/event-mpos/internal/member"
	membersync "github.com/krnskrll/event-mpos/internal/member/sync"
	"github.com/krnskrll/event-mpos/internal/product"
	"github.com/krnskrll/event-mpos/internal/product/erp"
	"github.com/krnskrll/event-mpos/internal/seed"
	"github.com/krnskrll/event-mpos/internal/user"
)

func main() {
	cfg := config.Load()

	// 1. Database + schema
	db, err := database.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("schema migration failed: %v", err)
	}
	log.Println("database connected + schema ready")

	// 2. Services
	jwtService, err := auth.NewJwtService(cfg.JWTSecret, cfg.AccessTokenTTLMin)
	if err != nil {
		log.Fatalf("jwt init failed: %v", err)
	}
	refreshService := auth.NewRefreshService(db, cfg.RefreshTokenTTLDays)
	userService := user.NewService(db)
	authService := auth.NewService(userService, jwtService, refreshService)
	eventService := event.NewService(db, userService)
	productService := product.NewService(db)
	erpSource := erp.NewSource(cfg.ERPSource, cfg.ERPFilePath, cfg.ERPRestURL)
	erpSyncService := erp.NewSyncService(erpSource, db)
	memberService := member.NewService(db)
	memberSource := membersync.NewSource(cfg.MemberSource, cfg.MemberFilePath, cfg.MemberRestURL)
	memberSyncService := membersync.NewService(memberSource, db, cfg.MemberRegionFilter)

	// Snapshot contributors: products + members live now; stock/promotions wired in later phases.
	eventService.SetProviders(event.Providers{Products: productService, Members: memberService})

	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)
	eventHandler := event.NewHandler(eventService)
	productHandler := product.NewHandler(productService)
	erpHandler := erp.NewHandler(erpSyncService)
	memberHandler := member.NewHandler(memberService)
	memberSyncHandler := membersync.NewHandler(memberSyncService)
	meHandler := me.NewHandler(eventService, productService, memberService)

	// 3. Seed dev data
	if cfg.SeedEnabled {
		if err := seed.Run(db, erpSyncService, memberSyncService); err != nil {
			log.Fatalf("seed failed: %v", err)
		}
	}

	// 4. Router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(gin.Logger(), gin.Recovery(), common.CORS(cfg.CORSOrigins))

	common.RegisterHealth(r)

	api := r.Group("/api")
	{
		// Public
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.Refresh)

		// Authenticated (any valid token)
		authed := api.Group("")
		authed.Use(auth.JWTMiddleware(jwtService))
		authed.POST("/auth/logout", authHandler.Logout)

		// BA-facing "me" surface (role enforced per-handler)
		meGroup := authed.Group("/me")
		meGroup.GET("", meHandler.Me)
		meGroup.GET("/event", meHandler.MyEvent)
		meGroup.GET("/events", meHandler.MyEvents)
		meGroup.GET("/events/:eventId/snapshot", meHandler.Snapshot)
		meGroup.GET("/events/:eventId/snapshot/hash", meHandler.SnapshotHash)
		meGroup.GET("/events/:eventId/products/by-barcode/:barcode", meHandler.ProductByBarcode)
		meGroup.GET("/members", meHandler.Members)
		meGroup.GET("/members/by-qr/:qrCode", meHandler.MemberByQr)

		// Admin-only
		admin := authed.Group("/admin")
		admin.Use(auth.RequireRole(user.RoleAdmin))
		admin.GET("/users", userHandler.List)
		admin.POST("/users", userHandler.Create)
		admin.PATCH("/users/:id", userHandler.Update)
		admin.GET("/events", eventHandler.List)
		admin.POST("/events", eventHandler.Create)
		admin.PUT("/events/:id", eventHandler.Update)
		admin.PATCH("/events/:id/status", eventHandler.ChangeStatus)
		admin.POST("/assignments", eventHandler.AssignBa)
		admin.GET("/products", productHandler.List)
		admin.GET("/products/:id", productHandler.Get)
		admin.POST("/products", productHandler.Create)
		admin.PUT("/products/:id", productHandler.Update)
		admin.POST("/erp/sync", erpHandler.SyncNow)
		admin.GET("/erp/status", erpHandler.SyncStatus)
		admin.GET("/events/:id/products", productHandler.EventProducts)
		admin.PUT("/events/:id/products/:productId", productHandler.SetAvailability)
		admin.GET("/members", memberHandler.List)
		admin.POST("/members/sync", memberSyncHandler.SyncNow)
		admin.GET("/members/sync/status", memberSyncHandler.Status)
	}

	// 5. Serve
	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
