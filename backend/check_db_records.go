package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=192.168.52.111 port=5432 user=postgres password=215o10 dbname=sub2api sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Check recent usage_logs
	log.Println("=== Recent usage_logs (last 5 minutes) ===")
	rows, err := db.Query(`
		SELECT id, user_id, api_key_id, client_request_id, request_id, model, created_at
		FROM usage_logs
		WHERE created_at > NOW() - INTERVAL '5 minutes'
		ORDER BY created_at DESC
		LIMIT 10
	`)
	if err != nil {
		log.Printf("Error querying usage_logs: %v", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			var id, userID, apiKeyID int64
			var clientRequestID string
			var requestID *string
			var model string
			var createdAt time.Time

			if err := rows.Scan(&id, &userID, &apiKeyID, &clientRequestID, &requestID, &model, &createdAt); err != nil {
				log.Printf("Error scanning: %v", err)
				continue
			}

			reqID := "NULL"
			if requestID != nil {
				reqID = *requestID
			}
			log.Printf("  ID=%d, UserID=%d, APIKeyID=%d, ClientReqID=%s, ReqID=%s, Model=%s, Created=%v",
				id, userID, apiKeyID, clientRequestID, reqID, model, createdAt)
			count++
		}
		if count == 0 {
			log.Println("  No records found in last 5 minutes")
		}
	}

	// Check recent request_logs
	log.Println("\n=== Recent request_logs (last 5 minutes) ===")
	rows2, err := db.Query(`
		SELECT id, user_id, api_key_id, client_request_id, request_id, model, created_at
		FROM request_logs
		WHERE created_at > NOW() - INTERVAL '5 minutes'
		ORDER BY created_at DESC
		LIMIT 10
	`)
	if err != nil {
		log.Printf("Error querying request_logs: %v", err)
	} else {
		defer rows2.Close()
		count := 0
		for rows2.Next() {
			var id, userID, apiKeyID int64
			var clientRequestID string
			var requestID *string
			var model string
			var createdAt time.Time

			if err := rows2.Scan(&id, &userID, &apiKeyID, &clientRequestID, &requestID, &model, &createdAt); err != nil {
				log.Printf("Error scanning: %v", err)
				continue
			}

			reqID := "NULL"
			if requestID != nil {
				reqID = *requestID
			}
			log.Printf("  ID=%d, UserID=%d, APIKeyID=%d, ClientReqID=%s, ReqID=%s, Model=%s, Created=%v",
				id, userID, apiKeyID, clientRequestID, reqID, model, createdAt)
			count++
		}
		if count == 0 {
			log.Println("  No records found in last 5 minutes")
		}
	}

	// Check total counts
	var totalUsageLogs, totalRequestLogs int
	db.QueryRow("SELECT COUNT(*) FROM usage_logs").Scan(&totalUsageLogs)
	db.QueryRow("SELECT COUNT(*) FROM request_logs").Scan(&totalRequestLogs)

	log.Printf("\n=== Total Records ===")
	log.Printf("  usage_logs: %d", totalUsageLogs)
	log.Printf("  request_logs: %d", totalRequestLogs)
}
