package services

import (
	"database/sql"
	"log"
	"time"
)

func StartAgendaScheduler(db *sql.DB) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			checkAgendas(db)
		}
	}()
}

func checkAgendas(db *sql.DB) {
	// Look for agendas in the next 15 minutes that haven't been notified
	now := time.Now()
	upcoming := now.Add(15 * time.Minute)

	// We need to construct a timestamp from date and time columns
	// In MySQL: CONCAT(date, ' ', time)
	query := `
		SELECT id, user_id, title, date, time 
		FROM agendas 
		WHERE is_notified = FALSE 
		AND TIMESTAMP(CONCAT(date, ' ', time)) BETWEEN ? AND ?
	`

	rows, err := db.Query(query, now, upcoming)
	if err != nil {
		log.Printf("Error checking agendas: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var userID int64
		var title string
		var dateStr string
		var timeStr string

		if err := rows.Scan(&id, &userID, &title, &dateStr, &timeStr); err != nil {
			continue
		}

		// Create notification
		// We use a raw query because importing 'handlers' or 'utils' might cause circular deps if not careful.
		// Be safe and just insert directly or use a clean 'notification' service if available.
		// For now, direct insert into notifications table.

		_, err := db.Exec(
			`INSERT INTO notifications (user_id, type, title, message, link, is_read, created_at)
			 VALUES (?, 'info', ?, ?, ?, FALSE, NOW())`,
			userID,
			"Agenda Title: "+title,
			"Agenda Anda '"+title+"' akan dimulai dalam kurang dari 15 menit.",
			"/calendar", // Link to calendar
		)

		if err == nil {
			// Mark as notified
			_, _ = db.Exec("UPDATE agendas SET is_notified = TRUE WHERE id = ?", id)
		}
	}
}
