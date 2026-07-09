package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/audit"
	"github.com/krnskrll/event-mpos/internal/auth"
	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/config"
	"github.com/krnskrll/event-mpos/internal/database"
	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/me"
	"github.com/krnskrll/event-mpos/internal/member"
	membersync "github.com/krnskrll/event-mpos/internal/member/sync"
	"github.com/krnskrll/event-mpos/internal/order"
	"github.com/krnskrll/event-mpos/internal/product"
	"github.com/krnskrll/event-mpos/internal/product/erp"
	"github.com/krnskrll/event-mpos/internal/promotion"
	"github.com/krnskrll/event-mpos/internal/receipt"
	"github.com/krnskrll/event-mpos/internal/report"
	"github.com/krnskrll/event-mpos/internal/seed"
	"github.com/krnskrll/event-mpos/internal/stock"
	"github.com/krnskrll/event-mpos/internal/syncconfig"
	"github.com/krnskrll/event-mpos/internal/user"
	"github.com/krnskrll/event-mpos/internal/ws"
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
	stockService := stock.NewService(db)
	promotionService := promotion.NewService(db)
	auditService := audit.NewService(db, userService)
	receiptService := receipt.NewService(db)
	orderService := order.NewService(db, stockService, promotionService, auditService, userService)

	// Real-time warehouse board: run the WebSocket hub and push order changes to it after commit.
	hub := ws.NewHub()
	go hub.Run()
	orderService.SetBroadcaster(func(eventID uuid.UUID, view order.WarehouseOrderView) {
		hub.Broadcast(eventID.String(), view)
	})

	// Reporting + runtime-editable sync schedules with a dynamic cron scheduler.
	reportService := report.NewService(db)
	syncConfigService := syncconfig.NewService(db, cfg.ERPSyncCron, cfg.MemberSyncCron)
	if err := syncConfigService.SeedDefaults(); err != nil {
		log.Fatalf("sync config seed failed: %v", err)
	}
	scheduler := syncconfig.NewScheduler(syncConfigService,
		func() { erpSyncService.Sync() },
		func() { memberSyncService.Sync() },
	)

	// Snapshot contributors: all four domains live now. Products carry remaining event stock.
	productService.SetStockRemaining(stockService.Remaining)
	eventService.SetProviders(event.Providers{
		Products: productService, Members: memberService, Stock: stockService, Promotions: promotionService,
	})

	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)
	eventHandler := event.NewHandler(eventService)
	productHandler := product.NewHandler(productService)
	erpHandler := erp.NewHandler(erpSyncService)
	memberHandler := member.NewHandler(memberService)
	memberSyncHandler := membersync.NewHandler(memberSyncService)
	stockAdminHandler := stock.NewAdminHandler(stockService)
	warehouseStockHandler := stock.NewWarehouseHandler(stockService, eventService)
	promotionHandler := promotion.NewHandler(promotionService)
	warehouseOrderHandler := order.NewWarehouseHandler(orderService)
	reportHandler := report.NewHandler(reportService, auditService)
	syncConfigHandler := syncconfig.NewHandler(syncConfigService, scheduler, erpSyncService, memberSyncService)
	meHandler := me.NewHandler(eventService, productService, memberService, stockService, promotionService, orderService, receiptService)

	// Append-only audit trigger (Postgres; skipped with a warning elsewhere).
	audit.InstallAppendOnlyGuard(db)

	// 3. Seed dev data
	if cfg.SeedEnabled {
		if err := seed.Run(db, erpSyncService, memberSyncService, stockService, promotionService); err != nil {
			log.Fatalf("seed failed: %v", err)
		}
	}

	// 4. Router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(gin.Logger(), gin.Recovery(), common.CORS(cfg.CORSOrigins))

	common.RegisterHealth(r)

	// Real-time warehouse board WebSocket (JWT validated inside the handler via ?token=).
	wsHandler := ws.NewHandler(hub, jwtService)
	r.GET("/ws/events/:eventId/orders", wsHandler.ServeWS)

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
		meGroup.GET("/events/:eventId/stock", meHandler.Stock)
		meGroup.POST("/events/:eventId/promotions/quote", meHandler.Quote)
		meGroup.POST("/events/:eventId/orders", meHandler.SubmitOrder)
		meGroup.GET("/events/:eventId/orders", meHandler.OrderHistory)
		meGroup.GET("/orders/:orderId", meHandler.Order)
		meGroup.POST("/orders/:orderId/receipts", meHandler.LogReceipt)

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
		admin.GET("/sync", syncConfigHandler.Overview)
		admin.PUT("/sync/:type", syncConfigHandler.Update)
		admin.GET("/events/:id/report", reportHandler.Report)
		admin.GET("/events/:id/report.csv", reportHandler.ReportCSV)
		admin.GET("/events/:id/audit", reportHandler.Audit)
		admin.GET("/events/:id/stock", stockAdminHandler.Allocations)
		admin.PUT("/events/:id/stock/:productId", stockAdminHandler.SetAllocation)
		admin.GET("/events/:id/promotions", promotionHandler.List)
		admin.POST("/events/:id/promotions", promotionHandler.Create)
		admin.PUT("/promotions/:id", promotionHandler.Update)
		admin.DELETE("/promotions/:id", promotionHandler.Delete)

		// Warehouse-only
		warehouse := authed.Group("/warehouse")
		warehouse.Use(auth.RequireRole(user.RoleWarehouse, user.RoleAdmin))
		warehouse.GET("/events", warehouseStockHandler.Events)
		warehouse.GET("/events/:eventId/stock", warehouseStockHandler.Levels)
		warehouse.GET("/events/:eventId/orders", warehouseOrderHandler.Orders)
		warehouse.PATCH("/orders/:orderId/status", warehouseOrderHandler.Advance)
	}

	// 5. Start the dynamic sync scheduler + serve
	scheduler.Start()

	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
