package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/krnskrll/event-mpos/internal/auth"
	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/config"
	"github.com/krnskrll/event-mpos/internal/database"
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

	// 2. Seed dev data
	if cfg.SeedEnabled {
		if err := seed.Run(db); err != nil {
			log.Fatalf("seed failed: %v", err)
		}
	}

	// 3. Services
	jwtService, err := auth.NewJwtService(cfg.JWTSecret, cfg.AccessTokenTTLMin)
	if err != nil {
		log.Fatalf("jwt init failed: %v", err)
	}
	refreshService := auth.NewRefreshService(db, cfg.RefreshTokenTTLDays)
	userService := user.NewService(db)
	authService := auth.NewService(userService, jwtService, refreshService)

	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)

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

		// Admin-only
		admin := authed.Group("/admin")
		admin.Use(auth.RequireRole(user.RoleAdmin))
		admin.GET("/users", userHandler.List)
		admin.POST("/users", userHandler.Create)
		admin.PATCH("/users/:id", userHandler.Update)
	}

	// 5. Serve
	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
