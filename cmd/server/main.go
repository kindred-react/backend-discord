package main

import (
	"log"

	"discord-backend/config"
	"discord-backend/internal/database"
	"discord-backend/internal/router"
	"discord-backend/internal/websocket"
)

func main() {
	cfg := config.Load()

	if err := database.Initialize(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	if err := database.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	hub := websocket.NewHub()
	go hub.Run()

	r := router.Setup(cfg, hub)

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
