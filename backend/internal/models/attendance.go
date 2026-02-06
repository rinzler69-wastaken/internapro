package models

import (
	"time"
)

type Attendance struct {
	ID       int64     `json:"id"`
	InternID int64     `json:"intern_id"`
	Date     time.Time `json:"date"`

	// FIX: Use Pointers (*) instead of sql.Null types for clean JSON
	CheckInTime       *time.Time `json:"check_in_time"`
	CheckInLatitude   *float64   `json:"check_in_latitude"`
	CheckInLongitude  *float64   `json:"check_in_longitude"`
	CheckOutTime      *time.Time `json:"check_out_time"`
	CheckOutLatitude  *float64   `json:"check_out_latitude"`
	CheckOutLongitude *float64   `json:"check_out_longitude"`

	Status        string   `json:"status"`
	LateReason    *string  `json:"late_reason"`
	Notes         *string  `json:"notes,omitempty"`
	DistanceMeters *int    `json:"distance_meters,omitempty"`
	ProofFile     *string  `json:"proof_file,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	InternName string `json:"intern_name,omitempty"`
}

type CheckInRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
	Reason    string  `json:"reason,omitempty"`
}

// CheckOutRequest: We removed Latitude/Longitude requirements!
type CheckOutRequest struct {
	Latitude  float64 `json:"latitude"`  // Optional now
	Longitude float64 `json:"longitude"` // Optional now
}

type OfficeSettings struct {
	ID                   int64     `json:"id"`
	Latitude             float64   `json:"latitude"`
	Longitude            float64   `json:"longitude"`
	RadiusMeters         int       `json:"radius_meters"`
	CheckInTime          string    `json:"check_in_time"`
	CheckOutTime         string    `json:"check_out_time"`
	LateToleranceMinutes int       `json:"late_tolerance_minutes"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
