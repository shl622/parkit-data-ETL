package main

import (
	"log"
	"parkit-data-ETL/internal/config"
	"parkit-data-ETL/internal/database"
	"parkit-data-ETL/internal/nyc"
	"parkit-data-ETL/internal/service"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize NYC client
	client := nyc.NewClient(*cfg.NYCAPI)

	// Initialize database
	db, err := database.Connect(&database.Config{
		URI:      cfg.MongoDB.URI,
		Database: cfg.MongoDB.Database,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Disconnect()

	// Initialize and run sync service
	syncService := service.NewSyncService(client, db)
	if err := syncService.Run(); err != nil {
		log.Fatal(err)
	}
}
