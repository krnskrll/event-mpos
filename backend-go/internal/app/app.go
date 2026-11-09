// Package app wires every service, handler and route into one Gin engine. main.go runs it; the
// integration tests build the same App against a test database and drive it with httptest, so the
// tests exercise the exact production router.
package app

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	auditsvc "github.com/krnskrll/event-mpos/internal/audit/service"
	authhandler "github.com/krnskrll/event-mpos/internal/auth/handler"
	authsvc "github.com/krnskrll/event-mpos/internal/auth/service"
	cataloghandler "github.com/krnskrll/event-mpos/internal/catalog/handler"
	catalogsvc "github.com/krnskrll/event-mpos/internal/catalog/service"
	devicehandler "github.com/krnskrll/event-mpos/internal/device/handler"
	devicesvc "github.com/krnskrll/event-mpos/internal/device/service"
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
	"github.com/krnskrll/event-mpos/internal/platform/database"
	"github.com/krnskrll/event-mpos/internal/platform/media"
	"github.com/krnskrll/event-mpos/internal/platform/realtime"
	producthandler "github.com/krnskrll/event-mpos/internal/product/handler"
	productsvc "github.com/krnskrll/event-mpos/internal/product/service"
	promotionhandler "github.com/krnskrll/event-mpos/internal/promotion/handler"
	promotionsvc "github.com/krnskrll/event-mpos/internal/promotion/service"
	receiptsvc "github.com/krnskrll/event-mpos/internal/receipt/service"
	reporthandler "github.com/krnskrll/event-mpos/internal/report/handler"
	reportsvc "github.com/krnskrll/event-mpos/internal/report/service"
	staffhandler "github.com/krnskrll/event-mpos/internal/staff/handler"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
	staffsvc "github.com/krnskrll/event-mpos/internal/staff/service"
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

	Users       *usersvc.Service
	Auth        *authsvc.Service
	Events      *eventsvc.Service
	Products    *productsvc.Service
	ERPSync     *erp.SyncService
	Members     *membersvc.Service
	MemberSync  *membersrc.Service
	Stock       *stocksvc.Service
	Promotions  *promotionsvc.Service
	Audit       *auditsvc.Service
	Receipts    *receiptsvc.Service
	Orders      *ordersvc.Service
	Reports     *reportsvc.Service
	SyncConfig  *syncsvc.Service
	Scheduler   *syncsvc.Scheduler
	Hub         *realtime.Hub
	Bhoomy      bhoomy.Client
	Media       *media.Store
	Catalog     *catalogsvc.Service
	CatalogSync *bhoomy.CatalogSync
	Staff       *staffsvc.Service
	Devices     *devicesvc.Service
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
	a.Media = media.NewStore(cfg.MediaDir, cfg.MediaBaseURL)
	a.Products = productsvc.NewService(db, a.Media)
	erpSource := erp.NewSource(cfg.ERPSource, cfg.ERPFilePath, cfg.ERPRestURL)
	a.ERPSync = erp.NewSyncService(erpSource, db, a.Products)
	a.Catalog = catalogsvc.NewService(db)
	a.Products.SetSource(a.Catalog)
	a.CatalogSync = bhoomy.NewCatalogSync(a.Bhoomy, a.Catalog, a.Products)
	a.Staff = staffsvc.NewService(db, a.Events)
	a.Devices = devicesvc.NewService(db, a.Events, a.Catalog.StoreName)
	// A closed location sells nothing: its devices keep the late-sync window, then stop working.
	a.Events.SetOnClose(func(tx database.Tx, locationID uuid.UUID, closedAt time.Time) error {
		return a.Devices.ScheduleRemoval(tx, locationID, closedAt.Add(24*time.Hour))
	})
	a.Members = membersvc.NewService(db)
	memberSource := membersrc.NewSource(cfg.MemberSource, cfg.MemberFilePath, cfg.MemberRestURL)
	a.MemberSync = membersrc.NewService(memberSource, db, a.Members, cfg.MemberRegionFilter)
	a.Stock = stocksvc.NewService(db, a.Events, a.Products)
	a.Promotions = promotionsvc.NewService(db, a.Products, a.Events)
	names := actorNames{a.Staff, a.Users}
	a.Audit = auditsvc.NewService(db, names)
	a.Orders = ordersvc.NewService(db, a.Events, a.Products, a.Stock, a.Promotions, a.Audit, names)
	a.Receipts = receiptsvc.NewService(db, a.Orders, a.Audit)

	// Real-time warehouse board: run the WebSocket hub and push order changes to it after commit.
	a.Hub = realtime.NewHub()
	go a.Hub.Run()
	a.Orders.SetBroadcaster(func(eventID uuid.UUID, view orderiface.WarehouseOrderView) {
		a.Hub.Broadcast(eventID.String(), view)
	})

	// Reporting + the runtime-editable member sync schedule. The ERP product sync only runs from the dev seed.
	a.Reports = reportsvc.NewService(a.Events, a.Orders)
	a.SyncConfig = syncsvc.NewService(db, cfg.MemberSyncCron)
	if err := a.SyncConfig.SeedDefaults(); err != nil {
		return nil, err
	}
	a.Scheduler = syncsvc.NewScheduler(a.SyncConfig, func() { a.MemberSync.Sync() })

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
	productHandler := producthandler.NewHandler(a.Products, a.Events)
	catalogHandler := cataloghandler.NewHandler(a.Catalog, func(storeID uuid.UUID) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		return a.CatalogSync.RunStore(ctx, storeID)
	})
	memberHandler := memberhandler.NewHandler(a.Members)
	memberSyncHandler := membersrc.NewHandler(a.MemberSync)
	stockAdminHandler := stockhandler.NewAdminHandler(a.Stock, a.Events)
	warehouseStockHandler := stockhandler.NewWarehouseHandler(a.Stock, a.Events)
	promotionHandler := promotionhandler.NewHandler(a.Promotions)
	warehouseOrderHandler := orderhandler.NewWarehouseHandler(a.Orders, a.Staff)
	reportHandler := reporthandler.NewHandler(a.Reports, a.Audit, a.Events)
	syncConfigHandler := synchandler.NewHandler(a.SyncConfig, a.Scheduler, a.MemberSync)
	staffHandler := staffhandler.NewHandler(a.Staff)
	deviceHandler := devicehandler.NewHandler(a.Devices, a.Staff)
	meHandler := mehandler.NewHandler(a.Events, a.Products, a.Members, a.Stock, a.Promotions, a.Orders, a.Receipts)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(common.AccessLogger(), gin.Recovery(), common.CORS(cfg.CORSOrigins))

	common.RegisterHealth(r)

	// Real-time pickup board WebSocket (device token validated inside the handler via ?token=).
	wsHandler := realtime.NewHandler(a.Hub, a.boardAccess)
	r.GET("/ws/events/:eventId/orders", wsHandler.ServeWS)
	r.GET("/media/:file", a.Media.Serve)

	api := r.Group("/api")
	{
		// Public
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.Refresh)
		api.POST("/device/pair", common.RateLimit(10, time.Minute), deviceHandler.Pair)

		// Cashier app on a paired device (Authorization: Device <token>; X-Staff-Id when acting for a person).
		device := api.Group("")
		device.Use(devicehandler.Auth(a.Devices, a.Staff))
		device.GET("/device", deviceHandler.Me)
		device.GET("/device/staff", deviceHandler.Staff)
		scoped := device.Group("", devicehandler.Scope(a.Orders.LocationOf))
		me := scoped.Group("/me")
		me.GET("/event", meHandler.MyEvent)
		me.GET("/events", meHandler.MyEvents)
		me.GET("/events/:eventId/snapshot", meHandler.Snapshot)
		me.GET("/events/:eventId/snapshot/hash", meHandler.SnapshotHash)
		me.GET("/events/:eventId/products/by-barcode/:barcode", meHandler.ProductByBarcode)
		me.GET("/members", meHandler.Members)
		me.GET("/members/by-qr/:qrCode", meHandler.MemberByQr)
		me.GET("/events/:eventId/stock", meHandler.Stock)
		me.POST("/events/:eventId/promotions/quote", meHandler.Quote)
		me.POST("/events/:eventId/orders", meHandler.SubmitOrder)
		me.POST("/events/:eventId/number-blocks", meHandler.IssueNumberBlock)
		me.GET("/events/:eventId/orders", meHandler.OrderHistory)
		me.GET("/orders/:orderId", meHandler.Order)
		me.POST("/orders/:orderId/receipts", devicehandler.RequireStaff, meHandler.LogReceipt)
		me.POST("/orders/:orderId/cancel", devicehandler.RequireStaff, warehouseOrderHandler.CashierCancel)
		// Pickup board: any staff on the location's devices.
		board := scoped.Group("/warehouse")
		board.GET("/events", warehouseStockHandler.Events)
		board.GET("/events/:eventId/stock", warehouseStockHandler.Levels)
		board.GET("/events/:eventId/orders", warehouseOrderHandler.Orders)
		board.PATCH("/orders/:orderId/status", devicehandler.RequireStaff, warehouseOrderHandler.Advance)
		board.POST("/orders/:orderId/cancel", devicehandler.RequireStaff, warehouseOrderHandler.BoardCancel)

		// Authenticated (any valid token)
		authed := api.Group("")
		authed.Use(authhandler.JWTMiddleware(jwtService))
		authed.POST("/auth/logout", authHandler.Logout)

		// Admin-only: what Bhoomy does not own yet (accounts, cashier assignment, stock, member sync).
		admin := authed.Group("/admin")
		admin.Use(authhandler.RequireRole(useriface.RoleAdmin))
		admin.GET("/users", userHandler.List)
		admin.POST("/users", userHandler.Create)
		admin.PATCH("/users/:id", userHandler.Update)
		admin.GET("/events", eventHandler.List)
		admin.POST("/assignments", eventHandler.AssignBa)
		admin.GET("/catalog/sync", catalogHandler.SyncStatus)
		admin.POST("/catalog/sync", catalogHandler.Sync)
		admin.POST("/members/sync", memberSyncHandler.SyncNow)
		admin.GET("/members/sync/status", memberSyncHandler.Status)
		admin.GET("/sync", syncConfigHandler.Overview)
		admin.PUT("/sync/:type", syncConfigHandler.Update)
		admin.GET("/events/:id/stock", stockAdminHandler.Allocations)
		admin.PUT("/events/:id/stock/:productId", stockAdminHandler.SetAllocation)

		// Bhoomy back office (Katalog POS, Outlet & Event, Promosi, Member POS, Laporan POS), called only by the Bhoomy backend.
		back := api.Group("/backoffice")
		back.Use(authhandler.ServiceMiddleware(cfg.BhoomyServiceToken, a.storeFor(cfg.BhoomyMode == "http")))
		back.GET("/catalog/status", catalogHandler.SyncStatus)
		back.POST("/catalog/sync", catalogHandler.Sync)
		back.GET("/catalog/source", productHandler.CatalogueSource)
		back.GET("/items", productHandler.CatalogueList)
		back.POST("/items", productHandler.CatalogueAdd)
		back.GET("/items/:id", productHandler.CatalogueGet)
		back.PUT("/items/:id", productHandler.CatalogueUpdate)
		back.POST("/items/:id/photo", productHandler.UploadPhoto)
		back.GET("/locations", productHandler.Locations)
		back.POST("/locations", eventHandler.CreateLocation)
		back.GET("/locations/:id", eventHandler.GetLocation)
		back.PUT("/locations/:id", eventHandler.Update)
		back.PATCH("/locations/:id/status", eventHandler.ChangeStatus)
		back.GET("/locations/:id/items", productHandler.LocationRows)
		back.PUT("/locations/:id/items", productHandler.SaveLocationRows)
		back.POST("/locations/:id/items", productHandler.AddLocationItems)
		back.GET("/locations/:id/report", reportHandler.Sales)
		back.GET("/locations/:id/audit", reportHandler.Audit)
		back.GET("/categories", productHandler.Categories)
		back.GET("/promotions", promotionHandler.StoreList)
		back.POST("/promotions", promotionHandler.StoreCreate)
		back.GET("/promotions/:id", promotionHandler.StoreGet)
		back.PUT("/promotions/:id", promotionHandler.StoreUpdate)
		back.DELETE("/promotions/:id", promotionHandler.Delete)
		back.GET("/members", memberHandler.StoreList)
		back.GET("/staff", staffHandler.List)
		back.POST("/staff", staffHandler.Create)
		back.GET("/staff/:id", staffHandler.Get)
		back.PUT("/staff/:id", staffHandler.Update)
		back.GET("/locations/:id/devices", deviceHandler.List)
		back.POST("/locations/:id/pairing-code", deviceHandler.NewCode)
		back.PATCH("/devices/:id", deviceHandler.Rename)
		back.DELETE("/devices/:id", deviceHandler.Remove)

	}
	return r
}

// boardAccess lets a device watch only its own location's pickup board.
func (a *App) boardAccess(token, eventID string) int {
	d, err := a.Devices.Authenticate(token)
	switch {
	case err != nil:
		return http.StatusInternalServerError
	case d == nil:
		return http.StatusUnauthorized
	case d.LocationID.String() != eventID:
		return http.StatusForbidden
	}
	return 0
}

// actorNames names an audit actor: a staff member, or an admin user for older entries.
type actorNames struct {
	staff staffiface.Staff
	users useriface.Users
}

func (n actorNames) NameOf(tx database.Tx, id uuid.UUID) (string, error) {
	if name, err := n.staff.NameOf(tx, id); err == nil {
		return name, nil
	}
	return n.users.NameOf(tx, id)
}

// storeFor finds the POS store of a Bhoomy store; a new one pulls its Produk Toko copy right away.
func (a *App) storeFor(pull bool) func(uuid.UUID, string) (uuid.UUID, error) {
	return func(bhoomyStore uuid.UUID, name string) (uuid.UUID, error) {
		id, created, err := a.Catalog.StoreFor(bhoomyStore, name)
		if created && pull {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				_ = a.CatalogSync.RunStore(ctx, id)
			}()
		}
		return id, err
	}
}

// CreateStoreAdmin adds a POS admin to a Bhoomy store, so the old dashboard can set up its cashiers
// until Phase 5 moves accounts to Bhoomy.
func (a *App) CreateStoreAdmin(bhoomyStore uuid.UUID, email, name, password string) (*useriface.UserResponse, error) {
	if len(password) < 8 || strings.TrimSpace(name) == "" {
		return nil, errors.New("name required, password at least 8 characters")
	}
	id, err := a.storeFor(false)(bhoomyStore, "")
	if err != nil {
		return nil, err
	}
	return a.Users.Create(id, useriface.CreateUserRequest{Email: email, Password: password, FullName: name, Role: useriface.RoleAdmin})
}
