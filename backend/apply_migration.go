package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection details
	connStr := "host=192.168.52.111 port=5432 user=postgres password=215o10 dbname=sub2api sslmode=disable"

	// Read migration file
	migrationSQL, err := os.ReadFile("migrations/046_add_client_request_id.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Execute migration
	log.Println("Applying migration 046_add_client_request_id.sql...")
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		log.Fatalf("Failed to apply migration: %v", err)
	}

	log.Println("Migration applied successfully!")
}
