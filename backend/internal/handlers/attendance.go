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

var holidayCache = struct {
	mu   sync.RWMutex
	data map[int]map[string]bool
}{
	data: map[int]map[string]bool{},
}

func NewAttendanceHandler(db *sql.DB) *AttendanceHandler {
	return &AttendanceHandler{db: db}
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

func isOffDay(t time.Time) bool {
	if t.Weekday() == time.Sunday {
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

	cfg := config.Loaded
	distance := utils.HaversineDistance(cfg.Office.Latitude, cfg.Office.Longitude, req.Latitude, req.Longitude)
	if distance > cfg.Office.Radius {
		utils.RespondBadRequest(w, "You are not within the office radius.")
		return
	}
	distanceMeters := int(distance + 0.5)

	today := time.Now().Format("2006-01-02")
	var existingID int64
	err = h.db.QueryRow("SELECT id FROM attendances WHERE intern_id = ? AND date = ?", internID, today).Scan(&existingID)
	if err == nil {
		utils.RespondBadRequest(w, "Already checked in today")
		return
	}

	// Time Logic
	now := time.Now()
	openTime, _ := time.Parse("15:04:05", cfg.Office.AttendanceOpenTime)
	targetTime, _ := time.Parse("15:04:05", cfg.Office.CheckInTime)

	openLimit := time.Date(now.Year(), now.Month(), now.Day(), openTime.Hour(), openTime.Minute(), 0, 0, now.Location())
	lateStart := time.Date(now.Year(), now.Month(), now.Day(), targetTime.Hour(), targetTime.Minute(), 0, 0, now.Location())
	hardLimit := lateStart.Add(time.Duration(cfg.Office.LateToleranceMinutes) * time.Minute)

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

	now := time.Now()
	today := now.Format("2006-01-02")

	// If today is an off-day (Sunday/holiday), short-circuit with friendly message
	if isOffDay(now) {
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
		if !isOffDay(yesterday) {
			yesterdayStr := yesterday.Format("2006-01-02")
			var yID int64
			if yErr := h.db.QueryRow("SELECT id FROM attendances WHERE intern_id = ? AND date = ?", internID, yesterdayStr).Scan(&yID); yErr == sql.ErrNoRows {
				_, _ = h.db.Exec(`INSERT INTO attendances (intern_id, date, status, created_at) VALUES (?, ?, 'absent', NOW())`,
					internID, yesterdayStr)
			}
		}

		// If today is weekend/holiday, show off-day message
		if isOffDay(now) {
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
	} else if internFilter != "" {
		if id, err := strconv.ParseInt(internFilter, 10, 64); err == nil {
			where = append(where, "a.intern_id = ?")
			args = append(args, id)
		}
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
		       u.name
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
		var internName sql.NullString

		if err := rows.Scan(
			&a.ID, &a.InternID, &a.Date,
			&cit, &cila, &cilo,
			&cot, &cola, &colo,
			&a.Status, &lr,
			&notes, &dist, &proof, &a.CreatedAt, &a.UpdatedAt,
			&internName,
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
		       u.name
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
	var internName sql.NullString

	err := h.db.QueryRow(query, id).Scan(
		&a.ID, &a.InternID, &a.Date,
		&cit, &cila, &cilo,
		&cot, &cola, &colo,
		&a.Status, &lr,
		&notes, &dist, &proof, &a.CreatedAt, &a.UpdatedAt,
		&internName,
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
