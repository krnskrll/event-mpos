package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/config"
	"github.com/krnskrll/event-mpos/internal/database"
)

func main() {
	cfg := config.Load()

	// 1. Database
	db, err := database.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	log.Println("database connected")

	// TODO(phase 1+): run migrations, seed data, wire domain handlers, cron, websocket hub.
	_ = db

	// 2. HTTP router
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	common.RegisterHealth(r)

	// 3. Serve
	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
