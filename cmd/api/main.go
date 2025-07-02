package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/adampetrovic/nrl-scheduler/internal/api"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Database connection
	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "nrl-scheduler.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// TODO: Run migrations - placeholder for now
	log.Println("Migrations skipped - placeholder implementation")

	// Create and start server
	server := api.NewServer(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting NRL Scheduler API server on port %s", port)
	if err := server.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
