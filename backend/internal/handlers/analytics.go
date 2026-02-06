package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"dsi_interna_sys/internal/config"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/gorilla/mux"
)

type AnalyticsHandler struct {
	db *sql.DB
}

func NewAnalyticsHandler(db *sql.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

// GetWeeklyTrends returns weekly attendance trends for an intern
func (h *AnalyticsHandler) GetWeeklyTrends(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid intern ID")
		return
	}

	// Get week parameter (default: current week)
	weekOffset := 0
	if offset := r.URL.Query().Get("week_offset"); offset != "" {
		weekOffset, _ = strconv.Atoi(offset)
	}

	// Calculate week start (Monday) and end (Sunday)
	now := time.Now()
	weekStart := startOfWeek(now.AddDate(0, 0, weekOffset*7))
	weekEnd := weekStart.AddDate(0, 0, 6)

	// Get intern name
	var internName string
	err = h.db.QueryRow("SELECT full_name FROM interns WHERE id = ?", internID).Scan(&internName)
	if err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	// Get daily attendance records for the week
	query := `
		SELECT date, check_in_time, status, late_reason
		FROM attendances
		WHERE intern_id = ? AND date BETWEEN ? AND ?
		ORDER BY date ASC
	`

	rows, err := h.db.Query(query, internID, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}
	defer rows.Close()

	// Map date to attendance record
	attendanceMap := make(map[string]models.DailyAttendanceStats)
	for rows.Next() {
		var date time.Time
		var checkInTime sql.NullTime
		var status string
		var lateReason sql.NullString

		if err := rows.Scan(&date, &checkInTime, &status, &lateReason); err != nil {
			continue
		}

		stats := models.DailyAttendanceStats{
			Date:      date,
			DayOfWeek: date.Weekday().String(),
			Status:    status,
		}

		if checkInTime.Valid {
			stats.CheckInTime = &checkInTime.Time
			hour := checkInTime.Time.Hour()
			minute := checkInTime.Time.Minute()
			stats.CheckInHour = &hour
			stats.CheckInMinute = &minute

			// Calculate minutes late
			if status == "late" {
				minutesLate := h.calculateMinutesLate(checkInTime.Time)
				stats.MinutesLate = &minutesLate

				if lateReason.Valid {
					stats.LateReason = &lateReason.String
				}
			}
		}

		attendanceMap[date.Format("2006-01-02")] = stats
	}

	// Build daily records for entire week (including days without records)
	var dailyRecords []models.DailyAttendanceStats
	for i := 0; i < 7; i++ {
		date := weekStart.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")

		if record, exists := attendanceMap[dateStr]; exists {
			dailyRecords = append(dailyRecords, record)
		} else {
			// No record for this day
			dailyRecords = append(dailyRecords, models.DailyAttendanceStats{
				Date:      date,
				DayOfWeek: date.Weekday().String(),
				Status:    "absent",
			})
		}
	}

	// Calculate weekly summary
	summary := h.calculateWeeklySummary(dailyRecords)

	trends := models.AttendanceTrends{
		InternID:     internID,
		InternName:   internName,
		WeekStart:    weekStart,
		WeekEnd:      weekEnd,
		DailyRecords: dailyRecords,
		Summary:      summary,
	}

	utils.RespondSuccess(w, "Weekly trends retrieved successfully", trends)
}

// GetCheckInPatterns returns hourly distribution of check-ins
func (h *AnalyticsHandler) GetCheckInPatterns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid intern ID")
		return
	}

	// Get date range (default: last 30 days)
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		days, _ = strconv.Atoi(d)
	}

	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	// Query to get hourly distribution
	query := `
		SELECT HOUR(check_in_time) as hour, COUNT(*) as count
		FROM attendances
		WHERE intern_id = ? AND date >= ? AND check_in_time IS NOT NULL
		GROUP BY HOUR(check_in_time)
		ORDER BY hour ASC
	`

	rows, err := h.db.Query(query, internID, startDate)
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}
	defer rows.Close()

	var patterns []models.CheckInPattern
	var total int

	// Get counts per hour
	hourCounts := make(map[int]int)
	for rows.Next() {
		var hour, count int
		if err := rows.Scan(&hour, &count); err != nil {
			continue
		}
		hourCounts[hour] = count
		total += count
	}

	// Build pattern array with percentages
	for hour := 0; hour < 24; hour++ {
		count := hourCounts[hour]
		percentage := 0.0
		if total > 0 {
			percentage = float64(count) / float64(total) * 100
		}

		patterns = append(patterns, models.CheckInPattern{
			Hour:       hour,
			Count:      count,
			Percentage: percentage,
		})
	}

	utils.RespondSuccess(w, "Check-in patterns retrieved successfully", map[string]interface{}{
		"intern_id":       internID,
		"period":          strconv.Itoa(days) + " days",
		"total_check_ins": total,
		"patterns":        patterns,
	})
}

// GetPerformanceInsights generates AI-like insights for an intern
func (h *AnalyticsHandler) GetPerformanceInsights(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid intern ID")
		return
	}

	// Get intern name
	var internName string
	err = h.db.QueryRow("SELECT full_name FROM interns WHERE id = ?", internID).Scan(&internName)
	if err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	// Analyze last 4 weeks
	weeks := 4
	startDate := time.Now().AddDate(0, 0, -weeks*7).Format("2006-01-02")

	// Get attendance statistics
	var totalDays, presentDays, lateDays, absentDays int
	var avgCheckInMinutes sql.NullFloat64

	statsQuery := `
		SELECT 
			COUNT(*) as total_days,
			SUM(CASE WHEN status IN ('present', 'late') THEN 1 ELSE 0 END) as present_days,
			SUM(CASE WHEN status = 'late' THEN 1 ELSE 0 END) as late_days,
			SUM(CASE WHEN status = 'absent' THEN 1 ELSE 0 END) as absent_days,
			AVG(CASE WHEN check_in_time IS NOT NULL 
				THEN HOUR(check_in_time) * 60 + MINUTE(check_in_time) 
				ELSE NULL END) as avg_check_in_minutes
		FROM attendances
		WHERE intern_id = ? AND date >= ?
	`

	err = h.db.QueryRow(statsQuery, internID, startDate).Scan(
		&totalDays, &presentDays, &lateDays, &absentDays, &avgCheckInMinutes,
	)
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	// Calculate rates
	attendanceRate := 0.0
	punctualityRate := 0.0
	if totalDays > 0 {
		attendanceRate = float64(presentDays) / float64(totalDays) * 100
		if presentDays > 0 {
			punctualityRate = float64(presentDays-lateDays) / float64(presentDays) * 100
		}
	}

	// Generate insights
	var strengths []string
	var concerns []string
	var suggestions []string

	// Analyze attendance
	if attendanceRate >= 95 {
		strengths = append(strengths, "Excellent attendance record ("+strconv.FormatFloat(attendanceRate, 'f', 1, 64)+"%)")
	} else if attendanceRate >= 85 {
		strengths = append(strengths, "Good attendance record")
	} else if attendanceRate < 75 {
		concerns = append(concerns, "Low attendance rate ("+strconv.FormatFloat(attendanceRate, 'f', 1, 64)+"%)")
		suggestions = append(suggestions, "Focus on improving daily attendance")
	}

	// Analyze punctuality
	if punctualityRate >= 90 {
		strengths = append(strengths, "Consistently punctual")
	} else if punctualityRate >= 70 {
		suggestions = append(suggestions, "Try to arrive on time more consistently")
	} else if punctualityRate < 70 {
		concerns = append(concerns, "Frequently arrives late")
		suggestions = append(suggestions, "Consider adjusting morning routine to arrive earlier")
	}

	// Analyze check-in time
	cfg := config.Loaded
	checkInTime, _ := time.Parse("15:04:05", cfg.Office.CheckInTime)
	expectedMinutes := float64(checkInTime.Hour()*60 + checkInTime.Minute())

	if avgCheckInMinutes.Valid {
		diff := avgCheckInMinutes.Float64 - expectedMinutes
		if diff < -10 {
			strengths = append(strengths, "Usually arrives early")
		} else if diff > 15 {
			concerns = append(concerns, "Average arrival time is after scheduled start")
		}
	}

	// Calculate overall score
	overallScore := int((attendanceRate*0.5 + punctualityRate*0.5))

	insights := models.PerformanceInsights{
		InternID:     internID,
		InternName:   internName,
		Period:       "Last " + strconv.Itoa(weeks) + " weeks",
		Strengths:    strengths,
		Concerns:     concerns,
		Suggestions:  suggestions,
		OverallScore: overallScore,
	}

	utils.RespondSuccess(w, "Performance insights generated successfully", insights)
}

// Helper: Calculate weekly summary
func (h *AnalyticsHandler) calculateWeeklySummary(dailyRecords []models.DailyAttendanceStats) models.WeeklySummary {
	var summary models.WeeklySummary
	summary.TotalDays = len(dailyRecords)

	var checkInMinutes []int
	var earliestCheckIn, latestCheckIn *time.Time

	for _, record := range dailyRecords {
		switch record.Status {
		case "present":
			summary.PresentDays++
		case "late":
			summary.LateDays++
			summary.PresentDays++ // Late still counts as present
		case "absent":
			summary.AbsentDays++
		case "on_leave", "sick", "permission":
			summary.OnLeaveDays++
		}

		// Collect check-in times for average calculation
		if record.CheckInTime != nil {
			minutes := record.CheckInTime.Hour()*60 + record.CheckInTime.Minute()
			checkInMinutes = append(checkInMinutes, minutes)

			if earliestCheckIn == nil || record.CheckInTime.Before(*earliestCheckIn) {
				earliestCheckIn = record.CheckInTime
			}
			if latestCheckIn == nil || record.CheckInTime.After(*latestCheckIn) {
				latestCheckIn = record.CheckInTime
			}
		}
	}

	// Calculate rates
	if summary.TotalDays > 0 {
		summary.AttendanceRate = float64(summary.PresentDays) / float64(summary.TotalDays) * 100
		if summary.PresentDays > 0 {
			onTimeDays := summary.PresentDays - summary.LateDays
			summary.PunctualityRate = float64(onTimeDays) / float64(summary.PresentDays) * 100
		}
	}

	// Calculate average check-in time
	if len(checkInMinutes) > 0 {
		sum := 0
		for _, m := range checkInMinutes {
			sum += m
		}
		avgMinutes := sum / len(checkInMinutes)
		hours := avgMinutes / 60
		minutes := avgMinutes % 60
		summary.AverageCheckInTime = time.Date(0, 1, 1, hours, minutes, 0, 0, time.UTC).Format("15:04")
	}

	summary.EarliestCheckIn = earliestCheckIn
	summary.LatestCheckIn = latestCheckIn

	// Determine tendency
	summary.Tendency = h.determineTendency(summary.LateDays, summary.PresentDays, checkInMinutes)

	return summary
}

// Helper: Determine attendance tendency
func (h *AnalyticsHandler) determineTendency(lateDays, presentDays int, checkInMinutes []int) string {
	if presentDays == 0 {
		return "no_data"
	}

	lateRate := float64(lateDays) / float64(presentDays)

	if lateRate > 0.5 {
		return "frequently_late"
	}

	if len(checkInMinutes) == 0 {
		return "inconsistent"
	}

	// Calculate variance
	sum := 0
	for _, m := range checkInMinutes {
		sum += m
	}
	avg := sum / len(checkInMinutes)

	// Get standard office start time
	cfg := config.Loaded
	checkInTime, _ := time.Parse("15:04:05", cfg.Office.CheckInTime)
	expectedMinutes := checkInTime.Hour()*60 + checkInTime.Minute()

	if avg < expectedMinutes-10 {
		return "early_bird"
	} else if avg <= expectedMinutes+5 {
		return "on_time"
	} else {
		return "inconsistent"
	}
}

// Helper: Calculate minutes late
func (h *AnalyticsHandler) calculateMinutesLate(checkInTime time.Time) int {
	cfg := config.Loaded
	expectedTime, _ := time.Parse("15:04:05", cfg.Office.CheckInTime)

	expected := time.Date(
		checkInTime.Year(), checkInTime.Month(), checkInTime.Day(),
		expectedTime.Hour(), expectedTime.Minute(), 0, 0, checkInTime.Location(),
	)
	expected = expected.Add(time.Duration(cfg.Office.LateToleranceMinutes) * time.Minute)

	if checkInTime.After(expected) {
		return int(checkInTime.Sub(expected).Minutes())
	}
	return 0
}

// Helper: Get start of week (Monday)
func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday
	}
	return time.Date(t.Year(), t.Month(), t.Day()-weekday+1, 0, 0, 0, 0, t.Location())
}
