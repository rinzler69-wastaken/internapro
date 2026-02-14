package models

import (
	"database/sql"
	"time"
)

type Assessment struct {
	ID                 int64          `json:"id"`
	InternID           int64          `json:"intern_id"`
	TaskID             sql.NullInt64  `json:"task_id,omitempty"`
	AssessedBy         int64          `json:"assessed_by"` // user_id (admin/pembimbing)
	Score              int            `json:"score"`       // 0-100 (average of criteria)
	Category           string         `json:"category"`    // auto-calculated: very_good, good, not_good, very_bad
	Aspect             string         `json:"aspect"`      // legacy: discipline, work_quality, attitude
	QualityScore       sql.NullInt64  `json:"quality_score,omitempty"`
	SpeedScore         sql.NullInt64  `json:"speed_score,omitempty"`
	InitiativeScore    sql.NullInt64  `json:"initiative_score,omitempty"`
	TeamworkScore      sql.NullInt64  `json:"teamwork_score,omitempty"`
	CommunicationScore sql.NullInt64  `json:"communication_score,omitempty"`
	Strengths          sql.NullString `json:"strengths,omitempty"`
	Improvements       sql.NullString `json:"improvements,omitempty"`
	Comments           sql.NullString `json:"comments,omitempty"`
	Notes              sql.NullString `json:"notes,omitempty"`
	AssessmentDate     time.Time      `json:"assessment_date"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`

	// Related data
	InternName     string `json:"intern_name,omitempty"`
	InternAvatar   string `json:"intern_avatar,omitempty"`
	AssessorName   string `json:"assessor_name,omitempty"`
	AssessorAvatar string `json:"assessor_avatar,omitempty"`
	TaskTitle      string `json:"task_title,omitempty"`
}

type CreateAssessmentRequest struct {
	InternID           int64  `json:"intern_id" validate:"required"`
	TaskID             *int64 `json:"task_id,omitempty"`
	QualityScore       int    `json:"quality_score" validate:"required,min=0,max=100"`
	SpeedScore         int    `json:"speed_score" validate:"required,min=0,max=100"`
	InitiativeScore    int    `json:"initiative_score" validate:"required,min=0,max=100"`
	TeamworkScore      int    `json:"teamwork_score" validate:"required,min=0,max=100"`
	CommunicationScore int    `json:"communication_score" validate:"required,min=0,max=100"`
	Strengths          string `json:"strengths,omitempty"`
	Improvements       string `json:"improvements,omitempty"`
	Comments           string `json:"comments,omitempty"`
	Aspect             string `json:"aspect,omitempty"`
	Notes              string `json:"notes,omitempty"`
	AssessmentDate     string `json:"assessment_date"` // accepts YYYY-MM-DD; parsed server-side
}

type UpdateAssessmentRequest struct {
	QualityScore       int       `json:"quality_score,omitempty" validate:"omitempty,min=0,max=100"`
	SpeedScore         int       `json:"speed_score,omitempty" validate:"omitempty,min=0,max=100"`
	InitiativeScore    int       `json:"initiative_score,omitempty" validate:"omitempty,min=0,max=100"`
	TeamworkScore      int       `json:"teamwork_score,omitempty" validate:"omitempty,min=0,max=100"`
	CommunicationScore int       `json:"communication_score,omitempty" validate:"omitempty,min=0,max=100"`
	Strengths          string    `json:"strengths,omitempty"`
	Improvements       string    `json:"improvements,omitempty"`
	Comments           string    `json:"comments,omitempty"`
	Aspect             string    `json:"aspect,omitempty"`
	Notes              string    `json:"notes,omitempty"`
	AssessmentDate     time.Time `json:"assessment_date,omitempty"`
}

// GetCategory returns the category based on score
func (a *Assessment) GetCategory() string {
	if a.Score >= 85 {
		return "very_good"
	} else if a.Score >= 70 {
		return "good"
	} else if a.Score >= 50 {
		return "not_good"
	}
	return "very_bad"
}

// GetCategoryIndo returns Indonesian translation
func (a *Assessment) GetCategoryIndo() string {
	switch a.Category {
	case "very_good":
		return "Sangat Baik"
	case "good":
		return "Baik"
	case "not_good":
		return "Tidak Baik"
	case "very_bad":
		return "Sangat Tidak Baik"
	default:
		return ""
	}
}
