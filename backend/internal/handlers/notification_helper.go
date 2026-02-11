package handlers

import (
	"database/sql"
	"log"
	"time"
)

// Shared helper to create notification
// Accessible by all handlers in this package
func createNotification(db *sql.DB, userID int64, typeStr, title, message, link string, data interface{}) error {
	_, err := db.Exec(
		`INSERT INTO notifications (user_id, type, title, message, link, is_read, created_at)
		 VALUES (?, ?, ?, ?, ?, 0, ?)`,
		userID, typeStr, title, message, nullIfEmpty(link), time.Now(),
	)
	if err != nil {
		log.Printf("[ERR] Failed to create notification for user %d: %v", userID, err)
	}
	return err
}
