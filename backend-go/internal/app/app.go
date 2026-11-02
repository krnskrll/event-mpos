// Package app wires every service, handler and route into one Gin engine. main.go runs it; the
// integration tests build the same App against a test database and drive it with httptest, so the
// tests exercise the exact production router.
package app

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	auditsvc "github.com/krnskrll/event-mpos/internal/audit/service"
	authhandler "github.com/krnskrll/event-mpos/internal/auth/handler"
	authsvc "github.com/krnskrll/event-mpos/internal/auth/service"
	eventhandler "github.com/krnskrll/event-mpos/internal/event/handler"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	eventsvc "github.com/krnskrll/event-mpos/internal/event/service"
	"github.com/krnskrll/event-mpos/internal/integration/bhoomy"
	"github.com/krnskrll/event-mpos/internal/integration/erp"
	"github.com/krnskrll/event-mpos/internal/integration/membersrc"
	mehandler "github.com/krnskrll/event-mpos/internal/me/handler"
	memberhandler "github.com/krnskrll/event-mpos/internal/member/handler"
	membersvc "github.com/krnskrll/event-mpos/internal/member/service"
	orderhandler "github.com/krnskrll/event-mpos/internal/order/handler"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	ordersvc "github.com/krnskrll/event-mpos/internal/order/service"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/config"
	"github.com/krnskrll/event-mpos/internal/platform/realtime"
	producthandler "github.com/krnskrll/event-mpos/internal/product/handler"
	productsvc "github.com/krnskrll/event-mpos/internal/product/service"
	promotionhandler "github.com/krnskrll/event-mpos/internal/promotion/handler"
	promotionsvc "github.com/krnskrll/event-mpos/internal/promotion/service"
	receiptsvc "github.com/krnskrll/event-mpos/internal/receipt/service"
	reporthandler "github.com/krnskrll/event-mpos/internal/report/handler"
	reportsvc "github.com/krnskrll/event-mpos/internal/report/service"
	stockhandler "github.com/krnskrll/event-mpos/internal/stock/handler"
	stocksvc "github.com/krnskrll/event-mpos/internal/stock/service"
	synchandler "github.com/krnskrll/event-mpos/internal/syncconfig/handler"
	syncsvc "github.com/krnskrll/event-mpos/internal/syncconfig/service"
	userhandler "github.com/krnskrll/event-mpos/internal/user/handler"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
	usersvc "github.com/krnskrll/event-mpos/internal/user/service"
)

// App holds the wired services and the HTTP router.
type App struct {
	Router *gin.Engine

	Users      *usersvc.Service
	Auth       *authsvc.Service
	Events     *eventsvc.Service
	Products   *productsvc.Service
	ERPSync    *erp.SyncService
	Members    *membersvc.Service
	MemberSync *membersrc.Service
	Stock      *stocksvc.Service
	Promotions *promotionsvc.Service
	Audit      *auditsvc.Service
	Receipts   *receiptsvc.Service
	Orders     *ordersvc.Service
	Reports    *reportsvc.Service
	SyncConfig *syncsvc.Service
	Scheduler  *syncsvc.Scheduler
	Hub        *realtime.Hub
	Bhoomy     bhoomy.Client
}

// New builds services, handlers and routes on an already-migrated database. It starts the WebSocket
// hub goroutine but not the sync scheduler (the caller decides; tests leave it off).
func New(cfg *config.Config, db *gorm.DB) (*App, error) {
	a := &App{}

	jwtService, err := authsvc.NewJwtService(cfg.JWTSecret, cfg.AccessTokenTTLMin)
	if err != nil {
		return nil, err
	}
	if a.Bhoomy, err = bhoomy.New(cfg.BhoomyMode, cfg.BhoomyURL, cfg.BhoomyServiceToken); err != nil {
		return nil, err
	}
	refreshService := authsvc.NewRefreshService(db, cfg.RefreshTokenTTLDays)
	a.Users = usersvc.NewService(db)
	a.Auth = authsvc.NewService(a.Users, jwtService, refreshService)
	a.Events = eventsvc.NewService(db, a.Users)
	a.Products = productsvc.NewService(db)
	erpSource := erp.NewSource(cfg.ERPSource, cfg.ERPFilePath, cfg.ERPRestURL)
	a.ERPSync = erp.NewSyncService(erpSource, db, a.Products)
	a.Members = membersvc.NewService(db)
	memberSource := membersrc.NewSource(cfg.MemberSource, cfg.MemberFilePath, cfg.MemberRestURL)
	a.MemberSync = membersrc.NewService(memberSource, db, a.Members, cfg.MemberRegionFilter)
	a.Stock = stocksvc.NewService(db, a.Events, a.Products)
	a.Promotions = promotionsvc.NewService(db, a.Products)
	a.Audit = auditsvc.NewService(db, a.Users)
	a.Orders = ordersvc.NewService(db, a.Events, a.Products, a.Stock, a.Promotions, a.Audit, a.Users)
	a.Receipts = receiptsvc.NewService(db, a.Orders, a.Audit)

	// Real-time warehouse board: run the WebSocket hub and push order changes to it after commit.
	a.Hub = realtime.NewHub()
	go a.Hub.Run()
	a.Orders.SetBroadcaster(func(eventID uuid.UUID, view orderiface.WarehouseOrderView) {
		a.Hub.Broadcast(eventID.String(), view)
	})

	// Reporting + runtime-editable sync schedules with a dynamic cron scheduler.
	a.Reports = reportsvc.NewService(a.Events, a.Products, a.Stock, a.Orders)
	a.SyncConfig = syncsvc.NewService(db, cfg.ERPSyncCron, cfg.MemberSyncCron)
	if err := a.SyncConfig.SeedDefaults(); err != nil {
		return nil, err
	}
	a.Scheduler = syncsvc.NewScheduler(a.SyncConfig,
		func() { a.ERPSync.Sync() },
		func() { a.MemberSync.Sync() },
	)

	// Snapshot contributors: all four domains live now. Products carry remaining event stock.
	a.Products.SetStockRemaining(a.Stock.Remaining)
	a.Events.SetProviders(eventiface.Providers{
		Products: a.Products, Members: a.Members, Stock: a.Stock, Promotions: a.Promotions,
	})

	// Append-only audit trigger (Postgres; skipped with a warning elsewhere).
	auditsvc.InstallAppendOnlyGuard(db)

	a.Router = a.routes(cfg, jwtService)
	return a, nil
}

func (a *App) routes(cfg *config.Config, jwtService *authsvc.JwtService) *gin.Engine {
	authHandler := authhandler.NewHandler(a.Auth)
	userHandler := userhandler.NewHandler(a.Users)
	eventHandler := eventhandler.NewHandler(a.Events)
	productHandler := producthandler.NewHandler(a.Products)
	erpHandler := erp.NewHandler(a.ERPSync)
	memberHandler := memberhandler.NewHandler(a.Members)
	memberSyncHandler := membersrc.NewHandler(a.MemberSync)
	stockAdminHandler := stockhandler.NewAdminHandler(a.Stock)
	warehouseStockHandler := stockhandler.NewWarehouseHandler(a.Stock, a.Events)
	promotionHandler := promotionhandler.NewHandler(a.Promotions)
	warehouseOrderHandler := orderhandler.NewWarehouseHandler(a.Orders)
	reportHandler := reporthandler.NewHandler(a.Reports, a.Audit)
	syncConfigHandler := synchandler.NewHandler(a.SyncConfig, a.Scheduler, a.ERPSync, a.MemberSync)
	meHandler := mehandler.NewHandler(a.Events, a.Products, a.Members, a.Stock, a.Promotions, a.Orders, a.Receipts)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(common.AccessLogger(), gin.Recovery(), common.CORS(cfg.CORSOrigins))

	common.RegisterHealth(r)

	// Real-time warehouse board WebSocket (JWT validated inside the handler via ?token=).
	wsHandler := realtime.NewHandler(a.Hub, jwtService)
	r.GET("/ws/events/:eventId/orders", wsHandler.ServeWS)

	api := r.Group("/api")
	{
		// Public
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.Refresh)

		// Authenticated (any valid token)
		authed := api.Group("")
		authed.Use(authhandler.JWTMiddleware(jwtService))
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
		meGroup.POST("/events/:eventId/number-blocks", meHandler.IssueNumberBlock)
		meGroup.GET("/events/:eventId/orders", meHandler.OrderHistory)
		meGroup.GET("/orders/:orderId", meHandler.Order)
		meGroup.POST("/orders/:orderId/receipts", meHandler.LogReceipt)
		meGroup.POST("/orders/:orderId/cancel", authhandler.RequireRole(useriface.RoleBA), warehouseOrderHandler.Cancel)

		// Admin-only
		admin := authed.Group("/admin")
		admin.Use(authhandler.RequireRole(useriface.RoleAdmin))
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
		warehouse.Use(authhandler.RequireRole(useriface.RoleWarehouse, useriface.RoleAdmin))
		warehouse.GET("/events", warehouseStockHandler.Events)
		warehouse.GET("/events/:eventId/stock", warehouseStockHandler.Levels)
		warehouse.GET("/events/:eventId/orders", warehouseOrderHandler.Orders)
		warehouse.PATCH("/orders/:orderId/status", warehouseOrderHandler.Advance)
		warehouse.POST("/orders/:orderId/cancel", warehouseOrderHandler.Cancel)
	}
	return r
}
