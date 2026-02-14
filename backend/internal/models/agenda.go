package models

import "time"

type Agenda struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Date        string    `json:"date"` // YYYY-MM-DD
	Time        string    `json:"time"` // HH:MM:SS
	IsNotified  bool      `json:"is_notified"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
