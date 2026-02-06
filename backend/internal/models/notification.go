package models

import (
	"database/sql"
	"time"
)

// Notification types
const (
	NotificationTaskAssigned       = "task_assigned"
	NotificationTaskDeadline       = "task_deadline"
	NotificationTaskApproved       = "task_approved"
	NotificationTaskRevision       = "task_revision"
	NotificationTaskSubmitted      = "task_submitted"
	NotificationNewIntern          = "new_intern_registration"
	NotificationNewSupervisor      = "new_supervisor_registration"
)

type Notification struct {
	ID        int64          `json:"id"`
	UserID    int64          `json:"user_id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Message   string         `json:"message"`
	Icon      sql.NullString `json:"icon,omitempty"`
	Link      sql.NullString `json:"link,omitempty"`
	Data      sql.NullString `json:"data,omitempty"` // JSON string
	ReadAt    sql.NullTime   `json:"read_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
