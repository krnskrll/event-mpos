// Package app wires every service, handler and route into one Gin engine. main.go runs it; the
// integration tests build the same App against a test database and drive it with httptest, so the
// tests exercise the exact production router.
package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	auditsvc "github.com/krnskrll/event-mpos/internal/audit/service"
	authhandler "github.com/krnskrll/event-mpos/internal/auth/handler"
	cataloghandler "github.com/krnskrll/event-mpos/internal/catalog/handler"
	catalogsvc "github.com/krnskrll/event-mpos/internal/catalog/service"
	devicehandler "github.com/krnskrll/event-mpos/internal/device/handler"
	devicesvc "github.com/krnskrll/event-mpos/internal/device/service"
	eventhandler "github.com/krnskrll/event-mpos/internal/event/handler"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	eventsvc "github.com/krnskrll/event-mpos/internal/event/service"
	"github.com/krnskrll/event-mpos/internal/integration/bhoomy"
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
	receipthandler "github.com/krnskrll/event-mpos/internal/receipt/handler"
	receiptsvc "github.com/krnskrll/event-mpos/internal/receipt/service"
	reporthandler "github.com/krnskrll/event-mpos/internal/report/handler"
	reportsvc "github.com/krnskrll/event-mpos/internal/report/service"
	staffhandler "github.com/krnskrll/event-mpos/internal/staff/handler"
	staffsvc "github.com/krnskrll/event-mpos/internal/staff/service"
	synchandler "github.com/krnskrll/event-mpos/internal/syncconfig/handler"
	syncsvc "github.com/krnskrll/event-mpos/internal/syncconfig/service"
)

// App holds the wired services and the HTTP router.
type App struct {
	Router *gin.Engine

	Events          *eventsvc.Service
	Products        *productsvc.Service
	Members         *membersvc.Service
	MemberSync      *membersrc.Service
	Promotions      *promotionsvc.Service
	Audit           *auditsvc.Service
	Receipts        *receiptsvc.Service
	ReceiptSettings *receiptsvc.Settings
	Orders          *ordersvc.Service
	Reports         *reportsvc.Service
	SyncConfig      *syncsvc.Service
	Scheduler       *syncsvc.Scheduler
	Hub             *realtime.Hub
	Bhoomy          bhoomy.Client
	Media           *media.Store
	Catalog         *catalogsvc.Service
	CatalogSync     *bhoomy.CatalogSync
	SalesSync       *bhoomy.SalesSync
	StockReader     *bhoomy.StockReader
	Staff           *staffsvc.Service
	Devices         *devicesvc.Service
}

// New builds services, handlers and routes on an already-migrated database. It starts the WebSocket
// hub goroutine but not the sync scheduler (the caller decides; tests leave it off).
func New(cfg *config.Config, db *gorm.DB) (*App, error) {
	a := &App{}

	var err error
	if a.Bhoomy, err = bhoomy.New(cfg.BhoomyMode, cfg.BhoomyURL, cfg.BhoomyServiceToken); err != nil {
		return nil, err
	}
	a.Events = eventsvc.NewService(db)
	a.Media = media.NewStore(cfg.MediaDir, cfg.MediaBaseURL)
	a.Products = productsvc.NewService(db, a.Media)
	a.Catalog = catalogsvc.NewService(db)
	a.Products.SetSource(a.Catalog)
	a.CatalogSync = bhoomy.NewCatalogSync(a.Bhoomy, a.Catalog, a.Products)
	a.Staff = staffsvc.NewService(db, a.Events)
	a.ReceiptSettings = receiptsvc.NewSettings(db)
	a.Devices = devicesvc.NewService(db, a.Events, a.Catalog.StoreName, a.ReceiptSettings.ForDevice)
	// A closed location sells nothing: its devices keep the late-sync window, then stop working.
	a.Events.SetOnClose(func(tx database.Tx, locationID uuid.UUID, closedAt time.Time) error {
		return a.Devices.ScheduleRemoval(tx, locationID, closedAt.Add(24*time.Hour))
	})
	a.Members = membersvc.NewService(db)
	memberSource := membersrc.NewSource(cfg.MemberSource, cfg.MemberFilePath, cfg.MemberRestURL)
	a.MemberSync = membersrc.NewService(memberSource, db, a.Members, cfg.MemberRegionFilter)
	a.Promotions = promotionsvc.NewService(db, a.Products, a.Events)
	a.Audit = auditsvc.NewService(db, a.Staff)
	a.Orders = ordersvc.NewService(db, a.Events, a.Products, a.Promotions, a.Audit, a.Staff)
	a.Receipts = receiptsvc.NewService(db, a.Orders, a.Audit)
	// 6c: sales go to Bhoomy's stock; the app reads Bhoomy's numbers (30 s in memory).
	a.StockReader = bhoomy.NewStockReader(a.Bhoomy, a.Catalog, a.Events, a.Products)
	a.SalesSync = bhoomy.NewSalesSync(a.Bhoomy, a.Orders, a.Events, a.Catalog)
	a.Orders.SetPlaceLookup(a.StockReader.PlaceOf)

	// Real-time warehouse board: run the WebSocket hub and push order changes to it after commit.
	a.Hub = realtime.NewHub()
	go a.Hub.Run()
	a.Orders.SetBroadcaster(func(eventID uuid.UUID, view orderiface.WarehouseOrderView) {
		a.Hub.Broadcast(eventID.String(), view)
	})

	// Reporting + the runtime-editable member sync schedule.
	a.Reports = reportsvc.NewService(a.Events, a.Orders)
	a.SyncConfig = syncsvc.NewService(db, cfg.MemberSyncCron)
	if err := a.SyncConfig.SeedDefaults(); err != nil {
		return nil, err
	}
	a.Scheduler = syncsvc.NewScheduler(a.SyncConfig, func() { a.MemberSync.Sync() })

	// Snapshot contributors. Stock is not in the snapshot: the app reads it from stock-levels (6c).
	a.Events.SetProviders(eventiface.Providers{
		Products: a.Products, Members: a.Members, Promotions: a.Promotions,
	})

	// Append-only audit trigger (Postgres; skipped with a warning elsewhere).
	auditsvc.InstallAppendOnlyGuard(db)

	a.Router = a.routes(cfg)
	return a, nil
}

func (a *App) routes(cfg *config.Config) *gin.Engine {
	eventHandler := eventhandler.NewHandler(a.Events)
	productHandler := producthandler.NewHandler(a.Products, a.Events)
	catalogHandler := cataloghandler.NewHandler(a.Catalog, func(storeID uuid.UUID) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		return a.CatalogSync.RunStore(ctx, storeID)
	})
	memberHandler := memberhandler.NewHandler(a.Members)
	promotionHandler := promotionhandler.NewHandler(a.Promotions)
	warehouseOrderHandler := orderhandler.NewWarehouseHandler(a.Orders, a.Staff)
	reportHandler := reporthandler.NewHandler(a.Reports, a.Audit, a.Events)
	memberSyncHandler := synchandler.NewHandler(a.SyncConfig, a.Scheduler, a.MemberSync, func() { a.MemberSync.Sync() })
	staffHandler := staffhandler.NewHandler(a.Staff)
	receiptHandler := receipthandler.NewHandler(a.ReceiptSettings)
	deviceHandler := devicehandler.NewHandler(a.Devices, a.Staff)
	meHandler := mehandler.NewHandler(a.Events, a.Members, a.Orders, a.Receipts)
	meHandler.SetStockLevels(func(ctx context.Context, id uuid.UUID) (any, error) { return a.StockReader.Levels(ctx, id) })

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	// Caddy reaches POS over the private network; trust its X-Forwarded-For so per-client limits work.
	_ = r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.1/32"})
	r.Use(common.AccessLogger(), gin.Recovery(), common.CORS(cfg.CORSOrigins))

	common.RegisterHealth(r)

	// Real-time pickup board WebSocket (device token validated inside the handler via ?token=).
	wsHandler := realtime.NewHandler(a.Hub, a.boardAccess)
	r.GET("/ws/events/:eventId/orders", wsHandler.ServeWS)
	r.GET("/media/:file", a.Media.Serve)

	api := r.Group("/api")
	{
		// Public
		api.POST("/device/pair", common.RateLimit(10, time.Minute), deviceHandler.Pair)

		// Cashier app on a paired device (Authorization: Device <token>; X-Staff-Id when acting for a person).
		device := api.Group("")
		device.Use(devicehandler.Auth(a.Devices, a.Staff))
		device.GET("/device", deviceHandler.Me)
		device.GET("/device/staff", deviceHandler.Staff)
		scoped := device.Group("", devicehandler.Scope(a.Orders.LocationOf))
		me := scoped.Group("/me")
		me.GET("/event", meHandler.MyEvent)
		me.GET("/events/:eventId/snapshot", meHandler.Snapshot)
		me.GET("/events/:eventId/snapshot/hash", meHandler.SnapshotHash)
		me.GET("/members/by-qr/:qrCode", meHandler.MemberByQr)
		me.GET("/events/:eventId/stock-levels", meHandler.StockLevels)
		me.POST("/events/:eventId/orders", meHandler.SubmitOrder)
		me.POST("/events/:eventId/number-blocks", meHandler.IssueNumberBlock)
		me.POST("/orders/:orderId/receipts", devicehandler.RequireStaff, meHandler.LogReceipt)
		me.POST("/orders/:orderId/cancel", devicehandler.RequireStaff, warehouseOrderHandler.CashierCancel)
		// Pickup board: any staff on the location's devices.
		board := scoped.Group("/warehouse")
		board.GET("/events/:eventId/orders", warehouseOrderHandler.Orders)
		board.PATCH("/orders/:orderId/status", devicehandler.RequireStaff, warehouseOrderHandler.Advance)
		board.POST("/orders/:orderId/cancel", devicehandler.RequireStaff, warehouseOrderHandler.BoardCancel)

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
		back.GET("/outlet", eventHandler.Outlet)
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
		back.GET("/members/sync", memberSyncHandler.Status)
		back.POST("/members/sync", memberSyncHandler.SyncNow)
		back.PUT("/members/sync/schedule", memberSyncHandler.Schedule)
		back.GET("/receipt", receiptHandler.Get)
		back.PUT("/receipt", receiptHandler.Save)
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
	d, err := a.Devices.Authenticate(token, -1)
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

// storeFor finds the POS store of a Bhoomy store; a new one pulls its Produk Toko copy right away.
func (a *App) storeFor(pull bool) func(uuid.UUID, string, string) (uuid.UUID, error) {
	return func(bhoomyStore uuid.UUID, name, address string) (uuid.UUID, error) {
		id, created, err := a.Catalog.StoreFor(bhoomyStore, name)
		if err != nil {
			return id, err
		}
		if err := a.Events.EnsureOutlet(id, a.Catalog.StoreName(id), address); err != nil {
			return id, err
		}
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
