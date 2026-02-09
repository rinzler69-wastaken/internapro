package models

import "time"

type SubmissionLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type Task struct {
	ID               int64            `json:"id"`
	TaskAssignmentID *int64           `json:"task_assignment_id,omitempty"`
	InternID         *int64           `json:"intern_id,omitempty"`
	AssignedBy       int64            `json:"assigned_by"` // user_id (admin/pembimbing)
	Title            string           `json:"title"`
	Description      *string          `json:"description,omitempty"`
	Priority         string           `json:"priority"` // low, medium, high
	Status           string           `json:"status"`   // pending, scheduled, in_progress, submitted, revision, completed
	StartDate        *time.Time       `json:"start_date,omitempty"`
	Deadline         *time.Time       `json:"deadline,omitempty"`
	DeadlineTime     *string          `json:"deadline_time,omitempty"` // HH:MM:SS
	StartedAt        *time.Time       `json:"started_at,omitempty"`
	SubmittedAt      *time.Time       `json:"submitted_at,omitempty"`
	CompletedAt      *time.Time       `json:"completed_at,omitempty"`
	ApprovedAt       *time.Time       `json:"approved_at,omitempty"`
	IsLate           bool             `json:"is_late"`
	SubmissionNotes  *string          `json:"submission_notes,omitempty"`
	SubmissionLinks  []SubmissionLink `json:"submission_links,omitempty"`
	Score            *int             `json:"score,omitempty"`
	AdminFeedback    *string          `json:"admin_feedback,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`

	// Related data (optional joins)
	InternName     string `json:"intern_name,omitempty"`
	AssignedByName string `json:"assigned_by_name,omitempty"`
}

type TaskAssignment struct {
	ID           int64       `json:"id"`
	Title        string      `json:"title"`
	Description  *string     `json:"description,omitempty"`
	AssignedBy   int64       `json:"assigned_by"`
	Priority     string      `json:"priority"`
	StartDate    *time.Time  `json:"start_date,omitempty"`
	Deadline     *time.Time  `json:"deadline,omitempty"`
	DeadlineTime *string     `json:"deadline_time,omitempty"`
	AssignToAll  bool        `json:"assign_to_all"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`

	AssignedByName string `json:"assigned_by_name,omitempty"`
	TasksCount     int    `json:"tasks_count,omitempty"`
}

type TaskAttachment struct {
	ID         int64     `json:"id"`
	TaskID     int64     `json:"task_id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileType   string    `json:"file_type"` // jpg, jpeg, png, pdf
	FileSize   int64     `json:"file_size"`
	UploadedAt time.Time `json:"uploaded_at"`
}
