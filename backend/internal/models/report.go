package models

import (
	"time"
)

type Report struct {
	ID          int64     `json:"id"`
	InternID    int64     `json:"intern_id"`
	CreatedBy   int64     `json:"created_by"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Type        string    `json:"type"`   // weekly, monthly, final
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Status      string    `json:"status"` // draft, submitted, reviewed
	Feedback    string    `json:"feedback,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Related data
	InternName   string `json:"intern_name,omitempty"`
	CreatedByName string `json:"created_by_name,omitempty"`
}
