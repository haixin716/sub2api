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

	// Check schema_migrations structure
	rows, err := db.Query(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = 'schema_migrations'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatalf("Error querying columns: %v", err)
	}
	defer rows.Close()

	log.Println("schema_migrations table structure:")
	for rows.Next() {
		var colName, dataType string
		if err := rows.Scan(&colName, &dataType); err != nil {
			log.Printf("Error scanning: %v", err)
			continue
		}
		log.Printf("  %s (%s)", colName, dataType)
	}

	// Show all migrations
	rows2, err := db.Query("SELECT * FROM schema_migrations ORDER BY 1 DESC LIMIT 15")
	if err != nil {
		log.Fatalf("Error querying migrations: %v", err)
	}
	defer rows2.Close()

	cols, _ := rows2.Columns()
	log.Printf("\nRecent migrations (columns: %v):", cols)

	for rows2.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows2.Scan(valuePtrs...); err != nil {
			log.Printf("Error scanning: %v", err)
			continue
		}

		log.Printf("  %v", values)
	}
}
