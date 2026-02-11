package models

import (
	"time"
)

// Notification types
const (
	NotificationTaskAssigned      = "task_assigned"
	NotificationTaskDeadline      = "task_deadline"
	NotificationTaskApproved      = "task_approved"
	NotificationTaskRevision      = "task_revision"
	NotificationTaskSubmitted     = "task_submitted"
	NotificationNewIntern         = "new_intern_registration"
	NotificationNewSupervisor     = "new_supervisor_registration"
	NotificationAssessmentCreated = "assessment_created"
	NotificationAttendanceLate    = "attendance_late"
)

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Icon      string    `json:"icon"`
	Link      string    `json:"link"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
