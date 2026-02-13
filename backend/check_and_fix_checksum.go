package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=192.168.52.111 port=5432 user=postgres password=215o10 dbname=sub2api sslmode=disable"

	// Read migration file exactly as the system does (with TrimSpace)
	contentBytes, err := os.ReadFile("migrations/046_add_client_request_id.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Trim space exactly as the system does in migrations_runner.go line 134
	content := strings.TrimSpace(string(contentBytes))

	// Calculate checksum
	sum := sha256.Sum256([]byte(content))
	correctChecksum := hex.EncodeToString(sum[:])

	log.Printf("File checksum (with TrimSpace): %s", correctChecksum)
	log.Printf("File size: %d bytes (raw), %d bytes (trimmed)", len(contentBytes), len(content))

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Get existing checksum
	var existingChecksum string
	err = db.QueryRow("SELECT checksum FROM schema_migrations WHERE filename = $1", "046_add_client_request_id.sql").Scan(&existingChecksum)
	if err != nil {
		log.Fatalf("Failed to query existing checksum: %v", err)
	}

	log.Printf("Database checksum:               %s", existingChecksum)

	if existingChecksum == correctChecksum {
		log.Println("✓ Checksums match, no update needed")
		return
	}

	// Update checksum
	log.Println("✗ Checksums don't match, updating database...")
	_, err = db.Exec("UPDATE schema_migrations SET checksum = $1 WHERE filename = $2", correctChecksum, "046_add_client_request_id.sql")
	if err != nil {
		log.Fatalf("Failed to update checksum: %v", err)
	}

	log.Println("✓ Checksum updated successfully!")
	log.Printf("  Updated from: %s", existingChecksum)
	log.Printf("  Updated to:   %s", correctChecksum)
}
