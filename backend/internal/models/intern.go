package models

import (
	"database/sql"
	"time"
)

type Intern struct {
	ID                  int64         `json:"id"`
	UserID              int64         `json:"user_id"`
	InstitutionID       sql.NullInt64 `json:"institution_id"`
	SupervisorID        sql.NullInt64 `json:"supervisor_id"`
	FullName            string         `json:"full_name"`
	NIS                 sql.NullString `json:"nis"`
	StudentID           sql.NullString `json:"student_id"`
	School              sql.NullString `json:"school"`
	Department          sql.NullString `json:"department"`
	DateOfBirth         sql.NullTime  `json:"date_of_birth"`
	Gender              string        `json:"gender"` // male, female
	Phone               sql.NullString `json:"phone"`
	Address             sql.NullString `json:"address"`
	StartDate           time.Time     `json:"start_date"`
	EndDate             time.Time     `json:"end_date"`
	Status              string        `json:"status"` // pending, active, completed, cancelled, terminated
	CertificateNumber   sql.NullString `json:"certificate_number,omitempty"`
	CertificateIssuedAt sql.NullTime   `json:"certificate_issued_at,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	
	// Related data (for joins)
	User        *User        `json:"user,omitempty"`
	Institution *Institution `json:"institution,omitempty"`
	Supervisor  *Supervisor  `json:"supervisor,omitempty"`
}

type InternWithDetails struct {
	Intern
	SupervisorName   string `json:"supervisor_name"`
	InstitutionName  string `json:"institution_name"`
	Email            string `json:"email"`
}
