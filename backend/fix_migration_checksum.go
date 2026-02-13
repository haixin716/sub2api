package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=192.168.52.111 port=5432 user=postgres password=215o10 dbname=sub2api sslmode=disable"

	// Read migration file and calculate checksum
	content, err := os.ReadFile("migrations/046_add_client_request_id.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	sum := sha256.Sum256(content)
	newChecksum := hex.EncodeToString(sum[:])

	log.Printf("Current file checksum: %s", newChecksum)

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

	log.Printf("Database checksum: %s", existingChecksum)

	if existingChecksum == newChecksum {
		log.Println("Checksums match, no update needed")
		return
	}

	// Update checksum
	log.Println("Updating checksum in database...")
	_, err = db.Exec("UPDATE schema_migrations SET checksum = $1 WHERE filename = $2", newChecksum, "046_add_client_request_id.sql")
	if err != nil {
		log.Fatalf("Failed to update checksum: %v", err)
	}

	log.Println("Checksum updated successfully!")
}
