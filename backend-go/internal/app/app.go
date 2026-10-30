// Package app wires every service, handler and route into one Gin engine. main.go runs it; the
// integration tests build the same App against a test database and drive it with httptest, so the
// tests exercise the exact production router.
package app

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/audit"
	"github.com/krnskrll/event-mpos/internal/auth"
	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/config"
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
	"github.com/krnskrll/event-mpos/internal/stock"
	"github.com/krnskrll/event-mpos/internal/syncconfig"
	"github.com/krnskrll/event-mpos/internal/user"
	"github.com/krnskrll/event-mpos/internal/ws"
)

// App holds the wired services and the HTTP router.
type App struct {
	Router *gin.Engine

	Users      *user.Service
	Auth       *auth.Service
	Events     *event.Service
	Products   *product.Service
	ERPSync    *erp.SyncService
	Members    *member.Service
	MemberSync *membersync.Service
	Stock      *stock.Service
	Promotions *promotion.Service
	Audit      *audit.Service
	Receipts   *receipt.Service
	Orders     *order.Service
	Reports    *report.Service
	SyncConfig *syncconfig.Service
	Scheduler  *syncconfig.Scheduler
	Hub        *ws.Hub
}

// New builds services, handlers and routes on an already-migrated database. It starts the WebSocket
// hub goroutine but not the sync scheduler (the caller decides; tests leave it off).
func New(cfg *config.Config, db *gorm.DB) (*App, error) {
	a := &App{}

	jwtService, err := auth.NewJwtService(cfg.JWTSecret, cfg.AccessTokenTTLMin)
	if err != nil {
		return nil, err
	}
	refreshService := auth.NewRefreshService(db, cfg.RefreshTokenTTLDays)
	a.Users = user.NewService(db)
	a.Auth = auth.NewService(a.Users, jwtService, refreshService)
	a.Events = event.NewService(db, a.Users)
	a.Products = product.NewService(db)
	erpSource := erp.NewSource(cfg.ERPSource, cfg.ERPFilePath, cfg.ERPRestURL)
	a.ERPSync = erp.NewSyncService(erpSource, db)
	a.Members = member.NewService(db)
	memberSource := membersync.NewSource(cfg.MemberSource, cfg.MemberFilePath, cfg.MemberRestURL)
	a.MemberSync = membersync.NewService(memberSource, db, cfg.MemberRegionFilter)
	a.Stock = stock.NewService(db)
	a.Promotions = promotion.NewService(db)
	a.Audit = audit.NewService(db)
	a.Receipts = receipt.NewService(db, a.Audit)
	a.Orders = order.NewService(db, a.Stock, a.Promotions, a.Audit, a.Users)

	// Real-time warehouse board: run the WebSocket hub and push order changes to it after commit.
	a.Hub = ws.NewHub()
	go a.Hub.Run()
	a.Orders.SetBroadcaster(func(eventID uuid.UUID, view order.WarehouseOrderView) {
		a.Hub.Broadcast(eventID.String(), view)
	})

	// Reporting + runtime-editable sync schedules with a dynamic cron scheduler.
	a.Reports = report.NewService(db)
	a.SyncConfig = syncconfig.NewService(db, cfg.ERPSyncCron, cfg.MemberSyncCron)
	if err := a.SyncConfig.SeedDefaults(); err != nil {
		return nil, err
	}
	a.Scheduler = syncconfig.NewScheduler(a.SyncConfig,
		func() { a.ERPSync.Sync() },
		func() { a.MemberSync.Sync() },
	)

	// Snapshot contributors: all four domains live now. Products carry remaining event stock.
	a.Products.SetStockRemaining(a.Stock.Remaining)
	a.Events.SetProviders(event.Providers{
		Products: a.Products, Members: a.Members, Stock: a.Stock, Promotions: a.Promotions,
	})

	// Append-only audit trigger (Postgres; skipped with a warning elsewhere).
	audit.InstallAppendOnlyGuard(db)

	a.Router = a.routes(cfg, jwtService)
	return a, nil
}

func (a *App) routes(cfg *config.Config, jwtService *auth.JwtService) *gin.Engine {
	authHandler := auth.NewHandler(a.Auth)
	userHandler := user.NewHandler(a.Users)
	eventHandler := event.NewHandler(a.Events)
	productHandler := product.NewHandler(a.Products)
	erpHandler := erp.NewHandler(a.ERPSync)
	memberHandler := member.NewHandler(a.Members)
	memberSyncHandler := membersync.NewHandler(a.MemberSync)
	stockAdminHandler := stock.NewAdminHandler(a.Stock)
	warehouseStockHandler := stock.NewWarehouseHandler(a.Stock, a.Events)
	promotionHandler := promotion.NewHandler(a.Promotions)
	warehouseOrderHandler := order.NewWarehouseHandler(a.Orders)
	reportHandler := report.NewHandler(a.Reports, a.Audit)
	syncConfigHandler := syncconfig.NewHandler(a.SyncConfig, a.Scheduler, a.ERPSync, a.MemberSync)
	meHandler := me.NewHandler(a.Events, a.Products, a.Members, a.Stock, a.Promotions, a.Orders, a.Receipts)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(common.AccessLogger(), gin.Recovery(), common.CORS(cfg.CORSOrigins))

	common.RegisterHealth(r)

	// Real-time warehouse board WebSocket (JWT validated inside the handler via ?token=).
	wsHandler := ws.NewHandler(a.Hub, jwtService)
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
		meGroup.POST("/orders/:orderId/cancel", auth.RequireRole(user.RoleBA), warehouseOrderHandler.Cancel)

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
		warehouse.POST("/orders/:orderId/cancel", warehouseOrderHandler.Cancel)
	}
	return r
}
