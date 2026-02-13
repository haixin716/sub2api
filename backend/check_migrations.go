package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=192.168.52.111 port=5432 user=postgres password=215o10 dbname=sub2api sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Check for common migration tracking tables
	tables := []string{"schema_migrations", "migrations", "goose_db_version", "flyway_schema_history"}

	for _, table := range tables {
		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", table).Scan(&exists)
		if err != nil {
			log.Printf("Error checking table %s: %v", table, err)
			continue
		}
		if exists {
			log.Printf("Found migration table: %s", table)

			// Show recent migrations
			rows, err := db.Query("SELECT * FROM " + table + " ORDER BY version DESC LIMIT 10")
			if err != nil {
				log.Printf("Error querying %s: %v", table, err)
				continue
			}
			defer rows.Close()

			cols, _ := rows.Columns()
			log.Printf("Columns: %v", cols)

			for rows.Next() {
				values := make([]interface{}, len(cols))
				valuePtrs := make([]interface{}, len(cols))
				for i := range values {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					log.Printf("Error scanning: %v", err)
					continue
				}

				log.Printf("Row: %v", values)
			}
		}
	}

	// Check if client_request_id column exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'usage_logs' AND column_name = 'client_request_id'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking column: %v", err)
	}

	log.Printf("client_request_id column exists in usage_logs: %v", exists)
}
