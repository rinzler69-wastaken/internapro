package models

import (
	"time"
)

// AttendanceTrends represents weekly attendance analysis
type AttendanceTrends struct {
	InternID        int64                  `json:"intern_id"`
	InternName      string                 `json:"intern_name"`
	WeekStart       time.Time              `json:"week_start"`
	WeekEnd         time.Time              `json:"week_end"`
	DailyRecords    []DailyAttendanceStats `json:"daily_records"`
	Summary         WeeklySummary          `json:"summary"`
}

// DailyAttendanceStats represents a single day's attendance stats
type DailyAttendanceStats struct {
	Date          time.Time `json:"date"`
	DayOfWeek     string    `json:"day_of_week"`
	CheckInTime   *time.Time `json:"check_in_time,omitempty"`
	CheckInHour   *int      `json:"check_in_hour,omitempty"`
	CheckInMinute *int      `json:"check_in_minute,omitempty"`
	Status        string    `json:"status"` // present, late, absent, on_leave
	MinutesLate   *int      `json:"minutes_late,omitempty"`
	LateReason    *string   `json:"late_reason,omitempty"`
}

// WeeklySummary contains aggregated weekly statistics
type WeeklySummary struct {
	TotalDays         int     `json:"total_days"`
	PresentDays       int     `json:"present_days"`
	LateDays          int     `json:"late_days"`
	AbsentDays        int     `json:"absent_days"`
	OnLeaveDays       int     `json:"on_leave_days"`
	AttendanceRate    float64 `json:"attendance_rate"`    // Percentage
	PunctualityRate   float64 `json:"punctuality_rate"`   // Percentage
	AverageCheckInTime string `json:"average_check_in_time"` // Format: "HH:MM"
	EarliestCheckIn   *time.Time `json:"earliest_check_in,omitempty"`
	LatestCheckIn     *time.Time `json:"latest_check_in,omitempty"`
	Tendency          string  `json:"tendency"` // "early_bird", "on_time", "frequently_late", "inconsistent"
}

// MonthlyTrends represents monthly attendance analysis
type MonthlyTrends struct {
	InternID      int64                  `json:"intern_id"`
	InternName    string                 `json:"intern_name"`
	Month         time.Time              `json:"month"` // First day of month
	WeeklyData    []WeeklySummary        `json:"weekly_data"`
	MonthlySummary MonthlySummary        `json:"monthly_summary"`
}

// MonthlySummary contains aggregated monthly statistics
type MonthlySummary struct {
	TotalWorkingDays   int     `json:"total_working_days"`
	PresentDays        int     `json:"present_days"`
	LateDays           int     `json:"late_days"`
	AbsentDays         int     `json:"absent_days"`
	OnLeaveDays        int     `json:"on_leave_days"`
	AttendanceRate     float64 `json:"attendance_rate"`
	PunctualityRate    float64 `json:"punctuality_rate"`
	OverallTendency    string  `json:"overall_tendency"`
	ImprovementTrend   string  `json:"improvement_trend"` // "improving", "declining", "stable"
}

// CheckInPattern represents hourly distribution of check-ins
type CheckInPattern struct {
	Hour      int     `json:"hour"`       // 0-23
	Count     int     `json:"count"`      // Number of check-ins at this hour
	Percentage float64 `json:"percentage"` // Percentage of total
}

// PerformanceInsights provides AI-like insights
type PerformanceInsights struct {
	InternID     int64    `json:"intern_id"`
	InternName   string   `json:"intern_name"`
	Period       string   `json:"period"` // e.g., "Last 4 weeks"
	Strengths    []string `json:"strengths"`
	Concerns     []string `json:"concerns"`
	Suggestions  []string `json:"suggestions"`
	OverallScore int      `json:"overall_score"` // 0-100
}
