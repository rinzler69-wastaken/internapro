package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"dsi_interna_sys/internal/config"
	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/utils"
)

type DashboardHandler struct {
	db *sql.DB
}

func NewDashboardHandler(db *sql.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// GetInternDashboard returns dashboard data for intern view
func (h *DashboardHandler) GetInternDashboard(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	if claims.Role != "intern" {
		utils.RespondForbidden(w, "Only interns can access this dashboard")
		return
	}

	var internID int64
	err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&internID)
	if err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	cfg := config.Loaded
	now := time.Now()
	weekStart := now.AddDate(0, 0, -6)

	// 1. Today's Attendance
	today := now.Format("2006-01-02")
	var todayAttendance struct {
		ID             sql.NullInt64
		CheckInTime    sql.NullTime
		CheckOutTime   sql.NullTime
		Status         sql.NullString
		LateReason     sql.NullString
		DistanceMeters sql.NullInt64
	}

	err = h.db.QueryRow(`
		SELECT id, check_in_time, check_out_time, status, late_reason, distance_meters
		FROM attendances WHERE intern_id = ? AND date = ?`, internID, today).Scan(
		&todayAttendance.ID, &todayAttendance.CheckInTime, &todayAttendance.CheckOutTime,
		&todayAttendance.Status, &todayAttendance.LateReason, &todayAttendance.DistanceMeters,
	)

	todayAtt := map[string]interface{}{
		"checked_in":    false,
		"checked_out":   false,
		"status":        nil,
		"check_in_time": nil,
		"distance":      nil,
	}

	if err == nil && todayAttendance.ID.Valid {
		todayAtt["checked_in"] = true
		if todayAttendance.CheckInTime.Valid {
			todayAtt["check_in_time"] = todayAttendance.CheckInTime.Time.Format("15:04")
		}
		if todayAttendance.Status.Valid {
			todayAtt["status"] = todayAttendance.Status.String
		}
		if todayAttendance.DistanceMeters.Valid {
			todayAtt["distance"] = int(todayAttendance.DistanceMeters.Int64)
		}
		if todayAttendance.CheckOutTime.Valid {
			todayAtt["checked_out"] = true
		}
	}

	// 2. Task Statistics
	var totalTasks, pendingTasks, inProgressTasks, completedTasks int
	err = h.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed
		FROM task_assignments WHERE intern_id = ?`, internID).Scan(
		&totalTasks, &pendingTasks, &inProgressTasks, &completedTasks,
	)

	taskStats := map[string]interface{}{
		"total":       totalTasks,
		"pending":     pendingTasks,
		"in_progress": inProgressTasks,
		"completed":   completedTasks,
		"percentage":  0,
	}
	if totalTasks > 0 {
		taskStats["percentage"] = int(float64(completedTasks) / float64(totalTasks) * 100)
	}

	// 2.1 Task Breakdown for Pie Chart (including submitted and revision)
	var submittedTasks, revisionTasks int
	h.db.QueryRow(`
		SELECT 
			SUM(CASE WHEN status = 'submitted' THEN 1 ELSE 0 END) as submitted,
			SUM(CASE WHEN status = 'revision' THEN 1 ELSE 0 END) as revision
		FROM task_assignments WHERE intern_id = ?`, internID).Scan(&submittedTasks, &revisionTasks)

	taskBreakdown := map[string]interface{}{
		"pending":     pendingTasks,
		"in_progress": inProgressTasks,
		"submitted":   submittedTasks,
		"completed":   completedTasks,
		"revision":    revisionTasks,
	}

	// 2.2 Weekly Attendance Counts for Bar Chart (formatted for CSS charts)
	weeklyLabels := []string{}
	weeklyData := []int{}
	weeklyColors := []string{}

	for i := 0; i < 7; i++ {
		dayName := weekStart.AddDate(0, 0, i).Format("Mon")
		dayDate := weekStart.AddDate(0, 0, i).Format("2006-01-02")
		var attStatus sql.NullString

		h.db.QueryRow(`
			SELECT status FROM attendances WHERE intern_id = ? AND date = ?`, internID, dayDate).Scan(&attStatus)

		status := "absent"
		if attStatus.Valid {
			status = attStatus.String
		}

		weeklyLabels = append(weeklyLabels, dayName)

		// Color based on status for chart (Vercel-inspired colors)
		color := "#f43f5e" // absent - rose/red
		switch status {
		case "present":
			color = "#10b981" // emerald/green
		case "late":
			color = "#f59e0b" // amber/yellow
		case "sick", "permission":
			color = "#6366f1" // indigo/blue
		}
		weeklyColors = append(weeklyColors, color)

		// Data value: 1 for present/late/sick/permission, 0 for absent
		presentVal := 0
		if status != "absent" {
			presentVal = 1
		}
		weeklyData = append(weeklyData, presentVal)
	}

	weeklyAttendanceCounts := map[string]interface{}{
		"labels": weeklyLabels,
		"data":   weeklyData,
		"colors": weeklyColors,
	}

	// 3. Recent Tasks (5 latest)
	tasksRows, err := h.db.Query(`
		SELECT id, title, status, priority, deadline, deadline_time, submitted_at, grade, is_late
		FROM task_assignments 
		WHERE intern_id = ? 
		ORDER BY created_at DESC LIMIT 5`, internID)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch tasks")
		return
	}
	defer tasksRows.Close()

	var recentTasks []map[string]interface{}
	for tasksRows.Next() {
		var t struct {
			ID           sql.NullInt64
			Title        sql.NullString
			Status       sql.NullString
			Priority     sql.NullString
			Deadline     sql.NullTime
			DeadlineTime sql.NullString
			SubmittedAt  sql.NullTime
			Grade        sql.NullString
			IsLate       sql.NullBool
		}
		tasksRows.Scan(&t.ID, &t.Title, &t.Status, &t.Priority, &t.Deadline, &t.DeadlineTime, &t.SubmittedAt, &t.Grade, &t.IsLate)

		task := map[string]interface{}{
			"id":       int(t.ID.Int64),
			"title":    t.Title.String,
			"status":   t.Status.String,
			"priority": t.Priority.String,
		}

		if t.Deadline.Valid {
			task["deadline"] = t.Deadline.Time.Format("2006-01-02")
		}
		if t.DeadlineTime.Valid {
			task["deadline_time"] = t.DeadlineTime.String
		}
		if t.SubmittedAt.Valid {
			task["submitted_at"] = t.SubmittedAt.Time.Format("2006-01-02T15:04:05")
		}
		if t.Grade.Valid {
			task["grade"] = t.Grade.String
		}
		task["is_late"] = t.IsLate.Bool

		recentTasks = append(recentTasks, task)
	}

	// 4. Weekly Attendance (last 7 days) - detailed version for list
	weeklyAttendance := []map[string]interface{}{}

	for i := 0; i < 7; i++ {
		date := weekStart.AddDate(0, 0, i).Format("2006-01-02")
		var attStatus sql.NullString

		h.db.QueryRow(`
			SELECT status FROM attendances WHERE intern_id = ? AND date = ?`, internID, date).Scan(&attStatus)

		status := "absent"
		if attStatus.Valid {
			status = attStatus.String
		}

		dayName := weekStart.AddDate(0, 0, i).Format("Mon")
		weeklyAttendance = append(weeklyAttendance, map[string]interface{}{
			"date":   date,
			"day":    dayName,
			"status": status,
		})
	}

	// 5. Attendance Percentage (last 30 days)
	var totalDays, presentDays int
	h.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status IN ('present', 'late') THEN 1 ELSE 0 END) as present
		FROM attendances 
		WHERE intern_id = ? AND date >= ?`,
		internID, now.AddDate(0, 0, -30).Format("2006-01-02")).Scan(&totalDays, &presentDays)

	attendancePercentage := 0
	if totalDays > 0 {
		attendancePercentage = int(float64(presentDays) / float64(totalDays) * 100)
	}

	// 6. Recent Attendance History (last 5)
	attendanceRows, err := h.db.Query(`
		SELECT date, status, check_in_time 
		FROM attendances 
		WHERE intern_id = ? 
		ORDER BY date DESC LIMIT 5`, internID)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch attendance history")
		return
	}
	defer attendanceRows.Close()

	var attendanceHistory []map[string]interface{}
	for attendanceRows.Next() {
		var date sql.NullTime
		var status, checkInTime sql.NullString

		attendanceRows.Scan(&date, &status, &checkInTime)

		attendanceHistory = append(attendanceHistory, map[string]interface{}{
			"date":     date.Time.Format("2006-01-02"),
			"status":   status.String,
			"check_in": checkInTime.String,
		})
	}

	// 7. Office config for map
	officeConfig := map[string]interface{}{
		"latitude":  cfg.Office.Latitude,
		"longitude": cfg.Office.Longitude,
		"radius":    cfg.Office.Radius,
		"name":      "PT. DUTA SOLUSI INFORMATIKA",
	}

	// Get user name from database
	var userName string
	h.db.QueryRow("SELECT name FROM users WHERE id = ?", claims.UserID).Scan(&userName)

	utils.RespondSuccess(w, "Dashboard data retrieved", map[string]interface{}{
		"today_attendance":         todayAtt,
		"task_stats":               taskStats,
		"task_breakdown":           taskBreakdown,
		"recent_tasks":             recentTasks,
		"weekly_attendance":        weeklyAttendance,
		"weekly_attendance_counts": weeklyAttendanceCounts,
		"attendance_percentage":    attendancePercentage,
		"attendance_history":       attendanceHistory,
		"office":                   officeConfig,
		"user": map[string]interface{}{
			"name": userName,
			"role": claims.Role,
		},
	})
}

// GetAdminDashboard returns dashboard data for admin/supervisor view
func (h *DashboardHandler) GetAdminDashboard(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	if claims.Role == "intern" {
		utils.RespondForbidden(w, "Access denied")
		return
	}

	now := time.Now()

	// 1. Total Interns
	var totalInterns int
	h.db.QueryRow("SELECT COUNT(*) FROM interns WHERE status = 'active'").Scan(&totalInterns)

	// 2. Total Tasks
	var totalTasks int
	h.db.QueryRow("SELECT COUNT(*) FROM task_assignments").Scan(&totalTasks)

	// 3. Task Completion Stats
	var completedOnTime, completedLate, pendingTasks int
	h.db.QueryRow(`
		SELECT 
			SUM(CASE WHEN status = 'completed' AND is_late = 0 THEN 1 ELSE 0 END) as on_time,
			SUM(CASE WHEN status = 'completed' AND is_late = 1 THEN 1 ELSE 0 END) as late,
			SUM(CASE WHEN status IN ('pending', 'in_progress', 'submitted') THEN 1 ELSE 0 END) as pending
		FROM task_assignments`).Scan(&completedOnTime, &completedLate, &pendingTasks)

	// 4. Today's Attendance
	today := now.Format("2006-01-02")
	var presentToday, totalToday int
	h.db.QueryRow(`
		SELECT 
			COUNT(CASE WHEN status IN ('present', 'late') THEN 1 END) as present,
			COUNT(*) as total
		FROM attendances WHERE date = ?`, today).Scan(&presentToday, &totalToday)

	// 5. Recent Tasks
	tasksRows, err := h.db.Query(`
		SELECT ta.id, ta.title, ta.status, ta.deadline, ta.is_late, u.name as intern_name
		FROM task_assignments ta
		JOIN interns i ON ta.intern_id = i.id
		JOIN users u ON i.user_id = u.id
		ORDER BY ta.created_at DESC LIMIT 5`)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch tasks")
		return
	}
	defer tasksRows.Close()

	var recentTasks []map[string]interface{}
	for tasksRows.Next() {
		var t struct {
			ID         sql.NullInt64
			Title      sql.NullString
			Status     sql.NullString
			Deadline   sql.NullTime
			IsLate     sql.NullBool
			InternName sql.NullString
		}
		tasksRows.Scan(&t.ID, &t.Title, &t.Status, &t.Deadline, &t.IsLate, &t.InternName)

		recentTasks = append(recentTasks, map[string]interface{}{
			"id":          int(t.ID.Int64),
			"title":       t.Title.String,
			"status":      t.Status.String,
			"intern_name": t.InternName.String,
			"is_late":     t.IsLate.Bool,
		})
	}

	// 6. Today's Attendance Records
	attendanceRows, err := h.db.Query(`
		SELECT a.id, a.status, a.check_in_time, u.name, a.distance_meters
		FROM attendances a
		JOIN interns i ON a.intern_id = i.id
		JOIN users u ON i.user_id = u.id
		WHERE a.date = ?
		ORDER BY a.created_at DESC`, today)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch attendance")
		return
	}
	defer attendanceRows.Close()

	var todayAttendance []map[string]interface{}
	for attendanceRows.Next() {
		var a struct {
			ID             sql.NullInt64
			Status         sql.NullString
			CheckInTime    sql.NullTime
			InternName     sql.NullString
			DistanceMeters sql.NullInt64
		}
		attendanceRows.Scan(&a.ID, &a.Status, &a.CheckInTime, &a.InternName, &a.DistanceMeters)

		att := map[string]interface{}{
			"intern_name": a.InternName.String,
			"status":      a.Status.String,
		}
		if a.CheckInTime.Valid {
			att["check_in_time"] = a.CheckInTime.Time.Format("15:04")
		}
		if a.DistanceMeters.Valid {
			att["distance"] = int(a.DistanceMeters.Int64)
		}
		todayAttendance = append(todayAttendance, att)
	}

	// 7. Weekly Attendance Trend
	weeklyTrend := []map[string]interface{}{}
	weekStart := now.AddDate(0, 0, -6)

	for i := 0; i < 7; i++ {
		date := weekStart.AddDate(0, 0, i).Format("2006-01-02")
		var present, absent int
		h.db.QueryRow(`
			SELECT 
				COUNT(CASE WHEN status IN ('present', 'late') THEN 1 END) as present,
				COUNT(CASE WHEN status NOT IN ('present', 'late', 'sick', 'permission') THEN 1 END) as absent
			FROM attendances WHERE date = ?`, date).Scan(&present, &absent)

		dayName := weekStart.AddDate(0, 0, i).Format("Mon")
		weeklyTrend = append(weeklyTrend, map[string]interface{}{
			"date":    date,
			"day":     dayName,
			"present": present,
			"absent":  absent,
		})
	}

	utils.RespondSuccess(w, "Admin dashboard data retrieved", map[string]interface{}{
		"stats": map[string]interface{}{
			"total_interns":     totalInterns,
			"total_tasks":       totalTasks,
			"completed_on_time": completedOnTime,
			"completed_late":    completedLate,
			"pending_tasks":     pendingTasks,
			"present_today":     presentToday,
			"total_today":       totalToday,
		},
		"recent_tasks":     recentTasks,
		"today_attendance": todayAttendance,
		"weekly_trend":     weeklyTrend,
	})
}
