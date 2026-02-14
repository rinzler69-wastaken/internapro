package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"dsi_interna_sys/internal/config"
	"dsi_interna_sys/internal/holiday"
	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/gorilla/mux"
)

type AttendanceHandler struct {
	db *sql.DB
}

// local wrapper to keep existing calls

var holidayCache = struct {
	mu   sync.RWMutex
	data map[int]map[string]bool
}{
	data: map[int]map[string]bool{},
}

type workCalendarConfig struct {
	workdays      map[time.Weekday]bool
	manualOffDate *time.Time
}

func defaultWorkdays() map[time.Weekday]bool {
	return map[time.Weekday]bool{
		time.Monday:    true,
		time.Tuesday:   true,
		time.Wednesday: true,
		time.Thursday:  true,
		time.Friday:    true,
		time.Saturday:  true,
	}
}

func parseWorkdays(val string) map[time.Weekday]bool {
	parsed := map[time.Weekday]bool{}
	if strings.TrimSpace(val) == "" {
		return parsed
	}

	for _, part := range strings.Split(val, ",") {
		p := strings.TrimSpace(strings.ToLower(part))
		switch p {
		case "0", "7", "sun", "sunday", "minggu":
			parsed[time.Sunday] = true
		case "1", "mon", "monday", "senin":
			parsed[time.Monday] = true
		case "2", "tue", "tuesday", "selasa":
			parsed[time.Tuesday] = true
		case "3", "wed", "wednesday", "rabu":
			parsed[time.Wednesday] = true
		case "4", "thu", "thursday", "kamis":
			parsed[time.Thursday] = true
		case "5", "fri", "friday", "jumat", "jum'at", "jum\u2019at":
			parsed[time.Friday] = true
		case "6", "sat", "saturday", "sabtu":
			parsed[time.Saturday] = true
		}
	}
	return parsed
}

func NewAttendanceHandler(db *sql.DB) *AttendanceHandler {
	return &AttendanceHandler{db: db}
}

// getRuntimeOfficeConfig loads office settings from DB (settings table) and falls back to config.Loaded.
// It lets admins change time/radius without restarting the service.
func (h *AttendanceHandler) getRuntimeOfficeConfig() config.OfficeConfig {
	cfg := config.Loaded.Office

	query := "SELECT `key`, `value` FROM settings WHERE `key` IN (" +
		"'office_latitude'," +
		"'office_longitude'," +
		"'max_checkin_distance'," +
		"'office_radius'," +
		"'attendance_open_time'," +
		"'check_in_time'," +
		"'check_out_time'," +
		"'late_tolerance_minutes'," +
		"'office_start_time'," +
		"'office_end_time'," +
		"'late_tolerance_time'" +
		")"

	rows, err := h.db.Query(query)
	if err != nil {
		return cfg
	}
	defer rows.Close()

	m := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err == nil {
			m[k] = v
		}
	}

	if val, ok := m["office_latitude"]; ok {
		if f, e := strconv.ParseFloat(val, 64); e == nil {
			cfg.Latitude = f
		}
	}
	if val, ok := m["office_longitude"]; ok {
		if f, e := strconv.ParseFloat(val, 64); e == nil {
			cfg.Longitude = f
		}
	}
	if val, ok := m["max_checkin_distance"]; ok {
		if f, e := strconv.ParseFloat(val, 64); e == nil {
			cfg.Radius = f
		}
	}
	if val, ok := m["office_radius"]; ok {
		if f, e := strconv.ParseFloat(val, 64); e == nil {
			cfg.Radius = f
		}
	}
	if val, ok := m["attendance_open_time"]; ok && val != "" {
		cfg.AttendanceOpenTime = val
	}
	if val, ok := m["check_in_time"]; ok && val != "" {
		cfg.CheckInTime = val
	} else if val, ok := m["office_start_time"]; ok && val != "" {
		cfg.CheckInTime = val
	}
	if val, ok := m["check_out_time"]; ok && val != "" {
		cfg.CheckOutTime = val
	} else if val, ok := m["office_end_time"]; ok && val != "" {
		cfg.CheckOutTime = val
	}
	if val, ok := m["late_tolerance_minutes"]; ok {
		if i, e := strconv.Atoi(val); e == nil {
			cfg.LateToleranceMinutes = i
		}
	}
	// Fallback: derive minutes from late_tolerance_time - check_in_time
	if cfg.LateToleranceMinutes == 0 {
		if ltt, ok := m["late_tolerance_time"]; ok && ltt != "" {
			if cfg.CheckInTime != "" {
				if start, err := time.Parse("15:04", cfg.CheckInTime[:5]); err == nil {
					if tol, err := time.Parse("15:04", ltt[:5]); err == nil {
						diff := int(tol.Sub(start).Minutes())
						if diff > 0 {
							cfg.LateToleranceMinutes = diff
						}
					}
				}
			}
		}
	}

	return cfg
}

func (h *AttendanceHandler) getWorkCalendarConfig() workCalendarConfig {
	cfg := workCalendarConfig{
		workdays: defaultWorkdays(),
	}

	rows, err := h.db.Query("SELECT `key`, `value` FROM settings WHERE `key` IN ('workdays', 'manual_off_date')")
	if err != nil {
		return cfg
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			continue
		}
		switch k {
		case "workdays":
			if parsed := parseWorkdays(v); len(parsed) > 0 {
				cfg.workdays = parsed
			}
		case "manual_off_date":
			if strings.TrimSpace(v) != "" {
				if t, e := time.Parse("2006-01-02", v); e == nil {
					cfg.manualOffDate = &t
				}
			}
		}
	}

	return cfg
}

func isHolidayDate(t time.Time) bool {
	year := t.Year()

	holidayCache.mu.RLock()
	if m, ok := holidayCache.data[year]; ok {
		defer holidayCache.mu.RUnlock()
		_, exists := m[t.Format("2006-01-02")]
		return exists
	}
	holidayCache.mu.RUnlock()

	hols, _ := holiday.GetHolidays(year)
	m := make(map[string]bool, len(hols))
	for _, h := range hols {
		m[h.Date] = true
	}
	holidayCache.mu.Lock()
	holidayCache.data[year] = m
	holidayCache.mu.Unlock()
	return m[t.Format("2006-01-02")]
}

func isOffDay(t time.Time, cal workCalendarConfig) bool {
	dateStr := t.Format("2006-01-02")

	if cal.manualOffDate != nil && dateStr == cal.manualOffDate.Format("2006-01-02") {
		return true
	}

	if cal.workdays != nil {
		if !cal.workdays[t.Weekday()] {
			return true
		}
	} else if t.Weekday() == time.Sunday {
		return true
	}

	return isHolidayDate(t)
}

// Helpers to convert SQL Nulls to JSON Pointers
func sqlNullTimeToPointer(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func sqlNullFloatToPointer(f sql.NullFloat64) *float64 {
	if f.Valid {
		return &f.Float64
	}
	return nil
}

func sqlNullStringToPointer(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func sqlNullIntToPointer(i sql.NullInt64) *int {
	if i.Valid {
		val := int(i.Int64)
		return &val
	}
	return nil
}

func prependUpload(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http") || strings.HasPrefix(path, "/uploads/") {
		return path
	}
	clean := strings.TrimLeft(path, "/")
	if strings.HasPrefix(clean, "uploads/") {
		return "/" + clean
	}
	return "/uploads/" + clean
}

// CheckIn (Strict: Location Required)
func (h *AttendanceHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	if claims.Role != "intern" {
		utils.RespondForbidden(w, "Only interns can check in")
		return
	}

	var req models.CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// 1. Strict Location Check
	if !utils.ValidateCoordinates(req.Latitude, req.Longitude) {
		utils.RespondBadRequest(w, "Invalid coordinates")
		return
	}

	var internID int64
	err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&internID)
	if err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	workCal := h.getWorkCalendarConfig()
	now := time.Now()
	if isOffDay(now, workCal) {
		utils.RespondBadRequest(w, "Office closed today")
		return
	}

	cfg := h.getRuntimeOfficeConfig()
	distance := utils.HaversineDistance(cfg.Latitude, cfg.Longitude, req.Latitude, req.Longitude)
	if distance > cfg.Radius {
		utils.RespondBadRequest(w, "You are not within the office radius.")
		return
	}
	distanceMeters := int(distance + 0.5)

	today := now.Format("2006-01-02")
	var existingID int64
	err = h.db.QueryRow("SELECT id FROM attendances WHERE intern_id = ? AND date = ?", internID, today).Scan(&existingID)
	if err == nil {
		utils.RespondBadRequest(w, "Already checked in today")
		return
	}

	// Time Logic
	openTime, _ := time.Parse("15:04:05", cfg.AttendanceOpenTime)
	targetTime, _ := time.Parse("15:04:05", cfg.CheckInTime)

	openLimit := time.Date(now.Year(), now.Month(), now.Day(), openTime.Hour(), openTime.Minute(), 0, 0, now.Location())
	lateStart := time.Date(now.Year(), now.Month(), now.Day(), targetTime.Hour(), targetTime.Minute(), 0, 0, now.Location())
	hardLimit := lateStart.Add(time.Duration(cfg.LateToleranceMinutes) * time.Minute)

	if now.Before(openLimit) {
		utils.RespondBadRequest(w, "Attendance is not open yet.")
		return
	}

	if now.After(hardLimit) {
		utils.RespondBadRequest(w, "Attendance check-in is closed.")
		return
	}

	status := "present"
	var lateReason sql.NullString

	if now.After(lateStart) {
		status = "late"
		if strings.TrimSpace(req.Reason) == "" {
			utils.RespondBadRequest(w, "You are late. A reason is required.")
			return
		}
		lateReason = sql.NullString{String: req.Reason, Valid: true}
	}

	result, err := h.db.Exec(
		`INSERT INTO attendances (intern_id, date, check_in_time, check_in_latitude, check_in_longitude, status, late_reason, distance_meters)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		internID, today, now, req.Latitude, req.Longitude, status, lateReason, distanceMeters,
	)

	if err != nil {
		utils.RespondInternalError(w, "Failed to create attendance record")
		return
	}

	attendanceID, _ := result.LastInsertId()

	if status == "late" {
		// Notify Supervisor
		var supervisorID sql.NullInt64
		_ = h.db.QueryRow("SELECT supervisor_id FROM interns WHERE id = ?", internID).Scan(&supervisorID)
		if supervisorID.Valid {
			_ = createNotification(h.db, supervisorID.Int64, models.NotificationAttendanceLate, "Keterlambatan Intern",
				"Intern terlambat check-in. Alasan: "+req.Reason, "/attendance/"+strconv.FormatInt(attendanceID, 10), nil)
		}
	}

	utils.RespondSuccess(w, "Check-in successful", map[string]interface{}{
		"attendance_id": attendanceID,
		"status":        status,
		"check_in_time": now,
	})
}

// CheckOut (Loose: Location Optional)
func (h *AttendanceHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	var req models.CheckOutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Even if body is invalid/empty, we proceed (we don't strictly need coords)
	}

	var internID int64
	err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&internID)
	if err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	today := time.Now().Format("2006-01-02")
	var attendanceID int64
	var checkOutTime sql.NullTime

	err = h.db.QueryRow(
		"SELECT id, check_out_time FROM attendances WHERE intern_id = ? AND date = ?",
		internID, today,
	).Scan(&attendanceID, &checkOutTime)

	if err != nil {
		utils.RespondBadRequest(w, "No check-in record found for today")
		return
	}

	if checkOutTime.Valid {
		utils.RespondBadRequest(w, "Already checked out today")
		return
	}

	// REMOVED: Geo-Validation logic here. We accept check-outs from anywhere.

	now := time.Now()
	// Use 0.0 if coords missing
	_, err = h.db.Exec(
		`UPDATE attendances SET check_out_time = ?, check_out_latitude = ?, check_out_longitude = ?
		 WHERE id = ?`,
		now, req.Latitude, req.Longitude, attendanceID,
	)

	if err != nil {
		utils.RespondInternalError(w, "Failed to update attendance")
		return
	}

	utils.RespondSuccess(w, "Check-out successful", nil)
}

// GetToday (Fixes "Invalid Date" bug)
func (h *AttendanceHandler) GetToday(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	var internID int64
	err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&internID)
	if err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	workCal := h.getWorkCalendarConfig()
	now := time.Now()
	today := now.Format("2006-01-02")

	// If today is an off-day (Sunday/holiday), short-circuit with friendly message
	if isOffDay(now, workCal) {
		utils.RespondSuccess(w, "Office closed today", map[string]interface{}{
			"checked_in": false,
			"off_day":    true,
			"date":       today,
			"status":     "off",
			"message":    "Tidak ada jadwal kantor!",
			"note":       "Selamat beristirahat!",
		})
		return
	}

	// Temporary SQL Null variables for scanning
	var (
		cit, cot               sql.NullTime
		cila, cilo, cola, colo sql.NullFloat64
		lr                     sql.NullString
		notes                  sql.NullString
		dist                   sql.NullInt64
		proof                  sql.NullString
	)

	var a models.Attendance
	err = h.db.QueryRow(
		`SELECT id, intern_id, date, check_in_time, check_in_latitude, check_in_longitude, 
		        check_out_time, check_out_latitude, check_out_longitude, status, late_reason, notes, distance_meters, proof_file
		 FROM attendances
		 WHERE intern_id = ? AND date = ?`,
		internID, today,
	).Scan(
		&a.ID, &a.InternID, &a.Date,
		&cit, &cila, &cilo,
		&cot, &cola, &colo,
		&a.Status, &lr, &notes, &dist, &proof,
	)

	if err == sql.ErrNoRows {
		cfg := config.Loaded
		targetTime, _ := time.Parse("15:04:05", cfg.Office.CheckInTime)
		hardLimit := time.Date(now.Year(), now.Month(), now.Day(),
			targetTime.Hour(), targetTime.Minute(), 0, 0, now.Location()).
			Add(time.Duration(cfg.Office.LateToleranceMinutes) * time.Minute)

		// If yesterday has no record, log it as absent (a full day passed) unless off-day
		yesterday := now.AddDate(0, 0, -1)
		if !isOffDay(yesterday, workCal) {
			yesterdayStr := yesterday.Format("2006-01-02")
			var yID int64
			if yErr := h.db.QueryRow("SELECT id FROM attendances WHERE intern_id = ? AND date = ?", internID, yesterdayStr).Scan(&yID); yErr == sql.ErrNoRows {
				_, _ = h.db.Exec(`INSERT INTO attendances (intern_id, date, status, created_at) VALUES (?, ?, 'absent', NOW())`,
					internID, yesterdayStr)
			}
		}

		// If today is weekend/holiday, show off-day message
		if isOffDay(now, workCal) {
			utils.RespondSuccess(w, "Office closed today", map[string]interface{}{
				"checked_in": false,
				"off_day":    true,
				"date":       today,
				"status":     "off",
				"message":    "Tidak ada jadwal kantor!",
				"note":       "Selamat beristirahat!",
			})
			return
		}

		// For today, if window closed, show closed flag but do NOT insert yet
		if now.After(hardLimit) {
			utils.RespondSuccess(w, "Attendance closed", map[string]interface{}{
				"checked_in":   false,
				"closed_today": true,
				"date":         today,
				"status":       "absent",
			})
			return
		}

		utils.RespondSuccess(w, "No attendance record", map[string]interface{}{"checked_in": false})
		return
	}

	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	// CONVERT SQL NULLS TO POINTERS (Critical Fix for "Invalid Date")
	a.CheckInTime = sqlNullTimeToPointer(cit)
	a.CheckInLatitude = sqlNullFloatToPointer(cila)
	a.CheckInLongitude = sqlNullFloatToPointer(cilo)
	a.CheckOutTime = sqlNullTimeToPointer(cot)
	a.CheckOutLatitude = sqlNullFloatToPointer(cola)
	a.CheckOutLongitude = sqlNullFloatToPointer(colo)
	a.LateReason = sqlNullStringToPointer(lr)
	a.Notes = sqlNullStringToPointer(notes)
	a.DistanceMeters = sqlNullIntToPointer(dist)
	a.ProofFile = sqlNullStringToPointer(proof)
	if a.ProofFile != nil {
		*a.ProofFile = prependUpload(*a.ProofFile)
	}

	utils.RespondSuccess(w, "Today's attendance", map[string]interface{}{
		"checked_in": true,
		"attendance": a,
	})
}

func (h *AttendanceHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}
	offset := (page - 1) * limit

	date := strings.TrimSpace(r.URL.Query().Get("date"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	internFilter := strings.TrimSpace(r.URL.Query().Get("intern_id"))
	month := strings.TrimSpace(r.URL.Query().Get("month"))

	search := strings.TrimSpace(r.URL.Query().Get("search"))

	where := []string{}
	args := []interface{}{}

	role := normalizeRole(claims.Role)
	if role == "intern" {
		var internID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&internID); err != nil {
			utils.RespondNotFound(w, "Intern not found")
			return
		}
		where = append(where, "a.intern_id = ?")
		args = append(args, internID)
	} else if role == "pembimbing" {
		where = append(where, "i.supervisor_id = ?")
		args = append(args, claims.UserID)
	} else if internFilter != "" {
		if id, err := strconv.ParseInt(internFilter, 10, 64); err == nil {
			where = append(where, "a.intern_id = ?")
			args = append(args, id)
		}
	}

	if search != "" {
		where = append(where, "u.name LIKE ?")
		args = append(args, "%"+search+"%")
	}
	if date != "" {
		where = append(where, "DATE(a.date) = ?")
		args = append(args, date)
	}
	if status != "" {
		where = append(where, "a.status = ?")
		args = append(args, status)
	}
	if month != "" {
		if parsed, err := time.Parse("2006-01", month); err == nil {
			where = append(where, "MONTH(a.date) = ? AND YEAR(a.date) = ?")
			args = append(args, parsed.Month(), parsed.Year())
		}
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	baseFrom := `
		FROM attendances a
		LEFT JOIN interns i ON a.intern_id = i.id
		LEFT JOIN users u ON i.user_id = u.id
	`

	var total int64
	countQuery := "SELECT COUNT(*) " + baseFrom + " " + whereClause
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		utils.RespondInternalError(w, "Failed to count attendance")
		return
	}

	query := `
		SELECT a.id, a.intern_id, a.date, a.check_in_time, a.check_in_latitude, a.check_in_longitude,
		       a.check_out_time, a.check_out_latitude, a.check_out_longitude, a.status, a.late_reason,
		       a.notes, a.distance_meters, a.proof_file, a.created_at, a.updated_at,
		       u.name, u.avatar
	` + baseFrom + " " + whereClause + " ORDER BY a.date DESC, a.created_at DESC LIMIT ? OFFSET ?"

	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch attendance")
		return
	}
	defer rows.Close()

	var records []models.Attendance
	for rows.Next() {
		var a models.Attendance
		var cit, cot sql.NullTime
		var cila, cilo, cola, colo sql.NullFloat64
		var lr, notes, proof sql.NullString
		var dist sql.NullInt64
		var internName, internAvatar sql.NullString

		if err := rows.Scan(
			&a.ID, &a.InternID, &a.Date,
			&cit, &cila, &cilo,
			&cot, &cola, &colo,
			&a.Status, &lr,
			&notes, &dist, &proof, &a.CreatedAt, &a.UpdatedAt,
			&internName, &internAvatar,
		); err != nil {
			continue
		}

		a.CheckInTime = sqlNullTimeToPointer(cit)
		a.CheckInLatitude = sqlNullFloatToPointer(cila)
		a.CheckInLongitude = sqlNullFloatToPointer(cilo)
		a.CheckOutTime = sqlNullTimeToPointer(cot)
		a.CheckOutLatitude = sqlNullFloatToPointer(cola)
		a.CheckOutLongitude = sqlNullFloatToPointer(colo)
		a.LateReason = sqlNullStringToPointer(lr)
		a.Notes = sqlNullStringToPointer(notes)
		a.DistanceMeters = sqlNullIntToPointer(dist)
		a.ProofFile = sqlNullStringToPointer(proof)
		if a.ProofFile != nil {
			*a.ProofFile = prependUpload(*a.ProofFile)
		}
		if internName.Valid {
			a.InternName = internName.String
		}
		if internAvatar.Valid {
			a.InternAvatar = internAvatar.String
		}

		records = append(records, a)
	}

	utils.RespondPaginated(w, records, utils.CalculatePagination(page, limit, total))
}

func (h *AttendanceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	query := `
		SELECT a.id, a.intern_id, a.date, a.check_in_time, a.check_in_latitude, a.check_in_longitude,
		       a.check_out_time, a.check_out_latitude, a.check_out_longitude, a.status, a.late_reason,
		       a.notes, a.distance_meters, a.proof_file, a.created_at, a.updated_at,
		       u.name, u.avatar
		FROM attendances a
		LEFT JOIN interns i ON a.intern_id = i.id
		LEFT JOIN users u ON i.user_id = u.id
		WHERE a.id = ?
	`

	var a models.Attendance
	var cit, cot sql.NullTime
	var cila, cilo, cola, colo sql.NullFloat64
	var lr, notes, proof sql.NullString
	var dist sql.NullInt64
	var internName, internAvatar sql.NullString

	err := h.db.QueryRow(query, id).Scan(
		&a.ID, &a.InternID, &a.Date,
		&cit, &cila, &cilo,
		&cot, &cola, &colo,
		&a.Status, &lr,
		&notes, &dist, &proof, &a.CreatedAt, &a.UpdatedAt,
		&internName, &internAvatar,
	)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Attendance not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	if normalizeRole(claims.Role) == "intern" {
		var myInternID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myInternID); err != nil || myInternID != a.InternID {
			utils.RespondForbidden(w, "You do not have access to this record")
			return
		}
	}

	a.CheckInTime = sqlNullTimeToPointer(cit)
	a.CheckInLatitude = sqlNullFloatToPointer(cila)
	a.CheckInLongitude = sqlNullFloatToPointer(cilo)
	a.CheckOutTime = sqlNullTimeToPointer(cot)
	a.CheckOutLatitude = sqlNullFloatToPointer(cola)
	a.CheckOutLongitude = sqlNullFloatToPointer(colo)
	a.LateReason = sqlNullStringToPointer(lr)
	a.Notes = sqlNullStringToPointer(notes)
	a.DistanceMeters = sqlNullIntToPointer(dist)
	a.ProofFile = sqlNullStringToPointer(proof)
	if a.ProofFile != nil {
		*a.ProofFile = prependUpload(*a.ProofFile)
	}
	if internName.Valid {
		a.InternName = internName.String
	}
	if internAvatar.Valid {
		a.InternAvatar = internAvatar.String
	}

	utils.RespondSuccess(w, "Attendance retrieved", a)
}

func (h *AttendanceHandler) GetByInternID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	q := r.URL.Query()
	q.Set("intern_id", strconv.FormatInt(internID, 10))
	r.URL.RawQuery = q.Encode()
	h.GetAll(w, r)
}

func (h *AttendanceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	// Check if attendance exists
	var internID int64
	err := h.db.QueryRow("SELECT intern_id FROM attendances WHERE id = ?", id).Scan(&internID)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Attendance record not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	// Permission check: Admin/Supervisor can delete any. Intern can delete ONLY their own (if logic allows).
	// For strictness, let's say:
	// - Admin/Supervisor: OK
	// - Intern: OK if it belongs to them.
	role := normalizeRole(claims.Role)
	if role == "intern" {
		var myInternID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myInternID); err != nil {
			utils.RespondInternalError(w, "Failed to verify intern identity")
			return
		}
		if myInternID != internID {
			utils.RespondForbidden(w, "You are not authorized to delete this attendance record")
			return
		}
	} else if role != "admin" && role != "supervisor" && role != "pembimbing" {
		// Just in case other roles exist
		utils.RespondForbidden(w, "Unauthorized action")
		return
	}

	// Perform Delete
	_, err = h.db.Exec("DELETE FROM attendances WHERE id = ?", id)
	if err != nil {
		utils.RespondInternalError(w, "Failed to delete attendance record")
		return
	}

	utils.RespondSuccess(w, "Attendance record deleted successfully", nil)
}

// SubmitPermission handles sick/permission attendance submission
func (h *AttendanceHandler) SubmitPermission(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) != "intern" {
		utils.RespondForbidden(w, "Only interns can submit permission")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondBadRequest(w, "Failed to parse form data")
		return
	}

	status := r.FormValue("status")
	notes := r.FormValue("notes")
	latStr := r.FormValue("latitude")
	lonStr := r.FormValue("longitude")

	if status != "sick" && status != "permission" {
		utils.RespondBadRequest(w, "Invalid status")
		return
	}
	if strings.TrimSpace(notes) == "" {
		utils.RespondBadRequest(w, "Notes are required")
		return
	}

	var lat, lon sql.NullFloat64
	if latStr != "" {
		if v, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = sql.NullFloat64{Float64: v, Valid: true}
		}
	}
	if lonStr != "" {
		if v, err := strconv.ParseFloat(lonStr, 64); err == nil {
			lon = sql.NullFloat64{Float64: v, Valid: true}
		}
	}

	var internID int64
	if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&internID); err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	today := time.Now().Format("2006-01-02")
	var existingID int64
	if err := h.db.QueryRow("SELECT id FROM attendances WHERE intern_id = ? AND date = ?", internID, today).Scan(&existingID); err == nil {
		utils.RespondBadRequest(w, "Attendance already exists for today")
		return
	}

	// Optional proof file
	var proofPath sql.NullString
	file, header, err := r.FormFile("proof_file")
	if err == nil {
		defer file.Close()
		path, err := utils.UploadFile(file, header, "attendance_proofs")
		if err != nil {
			utils.RespondInternalError(w, "Failed to upload proof file")
			return
		}
		proofPath = sql.NullString{String: path, Valid: true}
	}

	_, err = h.db.Exec(
		`INSERT INTO attendances (intern_id, date, status, notes, proof_file, check_in_latitude, check_in_longitude)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		internID, today, status, notes, proofPath, lat, lon,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to create permission record")
		return
	}

	utils.RespondCreated(w, "Permission submitted", nil)
}
