package models

import (
	"time"
)

type LeaveRequest struct {
	ID             int64      `json:"id"`
	InternID       int64      `json:"intern_id"`
	LeaveType      string     `json:"leave_type"` // sick, permission, other
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	Reason         string     `json:"reason"`
	AttachmentPath *string    `json:"attachment_path,omitempty"` // Pointer for JSON null handling
	Status         string     `json:"status"`                    // pending, approved, rejected
	ApprovedBy     *int64     `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Joined Fields
	InternName   string `json:"intern_name,omitempty"`
	ApproverName string `json:"approver_name,omitempty"`
}

// CreateLeaveRequest struct is not strictly used for Multipart forms
// but useful for validation reference
type CreateLeaveRequest struct {
	LeaveType string `validate:"required,oneof=sick permission other"`
	StartDate string `validate:"required"` // Receive as string "2006-01-02", parse manually
	EndDate   string `validate:"required"`
	Reason    string `validate:"required,min=5"`
}
