package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/seriestracker/backend/config"
	"github.com/seriestracker/backend/db"
	"github.com/seriestracker/backend/routes"
)

func main() {
	cfg := config.Load()

	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	db.Migrate(database)

	handler := routes.Register(
		database,
		cfg.JWTSecret,
		cfg.CloudinaryCloud,
		cfg.CloudinaryKey,
		cfg.CloudinarySecret,
	)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Server running on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}