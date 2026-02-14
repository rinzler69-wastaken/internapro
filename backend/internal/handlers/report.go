package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/gorilla/mux"
	"github.com/phpdave11/gofpdf"
)

type ReportHandler struct {
	db *sql.DB
}

func NewReportHandler(db *sql.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

type createReportRequest struct {
	InternID    int64  `json:"intern_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Type        string `json:"type"` // weekly, monthly, final
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	Status      string `json:"status"` // submitted, draft
}

type updateReportRequest struct {
	Title       *string `json:"title,omitempty"`
	Content     *string `json:"content,omitempty"`
	Type        *string `json:"type,omitempty"`
	PeriodStart *string `json:"period_start,omitempty"`
	PeriodEnd   *string `json:"period_end,omitempty"`
	Status      *string `json:"status,omitempty"`
	Feedback    *string `json:"feedback,omitempty"`
}

func (h *ReportHandler) GetAll(w http.ResponseWriter, r *http.Request) {
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
		limit = 10
	}
	offset := (page - 1) * limit

	filterType := strings.TrimSpace(r.URL.Query().Get("type"))
	filterStatus := strings.TrimSpace(r.URL.Query().Get("status"))
	filterIntern := strings.TrimSpace(r.URL.Query().Get("intern_id"))
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
		where = append(where, "r.intern_id = ?")
		args = append(args, internID)
	} else if filterIntern != "" {
		if id, err := strconv.ParseInt(filterIntern, 10, 64); err == nil {
			where = append(where, "r.intern_id = ?")
			args = append(args, id)
		}
	}

	if search != "" {
		where = append(where, "(r.title LIKE ? OR iu.name LIKE ?)")
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	// OLD LOGIC removed:
	/*
		if role == "admin" || role == "supervisor" || role == "pembimbing" {
			where = append(where, "(r.status != 'draft' OR r.created_by = ?)")
			args = append(args, claims.UserID)
		}
	*/

	// NEW STRICT LOGIC:
	// Everyone (Admin, Supervisor, Intern) can only see:
	// 1. Reports that are NOT 'draft' (submitted, approved, etc.)
	// 2. 'draft' reports that THEY created (created_by = me)
	where = append(where, "(r.status != 'draft' OR r.created_by = ?)")
	args = append(args, claims.UserID)

	if filterType != "" {
		where = append(where, "r.type = ?")
		args = append(args, filterType)
	}
	if filterStatus != "" {
		where = append(where, "r.status = ?")
		args = append(args, filterStatus)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	baseFrom := `
		FROM reports r
		LEFT JOIN interns i ON r.intern_id = i.id
		LEFT JOIN users iu ON i.user_id = iu.id
		LEFT JOIN users cu ON r.created_by = cu.id
	`

	var total int64
	if err := h.db.QueryRow("SELECT COUNT(*) "+baseFrom+" "+whereClause, args...).Scan(&total); err != nil {
		utils.RespondInternalError(w, "Failed to count reports")
		return
	}

	query := `
		SELECT r.id, r.intern_id, r.created_by, r.title, r.content, r.type,
		       r.period_start, r.period_end, r.status, r.feedback, r.created_at, r.updated_at,
		       iu.name, cu.name, iu.avatar, cu.avatar
	` + baseFrom + " " + whereClause + " ORDER BY r.created_at DESC LIMIT ? OFFSET ?"

	args = append(args, limit, offset)
	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch reports")
		return
	}
	defer rows.Close()

	var reports []models.Report
	for rows.Next() {
		var rep models.Report
		var feedback sql.NullString
		var internName, createdByName, internAvatar, createdByAvatar sql.NullString
		if err := rows.Scan(
			&rep.ID, &rep.InternID, &rep.CreatedBy, &rep.Title, &rep.Content, &rep.Type,
			&rep.PeriodStart, &rep.PeriodEnd, &rep.Status, &feedback, &rep.CreatedAt, &rep.UpdatedAt,
			&internName, &createdByName, &internAvatar, &createdByAvatar,
		); err == nil {
			if feedback.Valid {
				rep.Feedback = feedback.String
			}
			if internName.Valid {
				rep.InternName = internName.String
			}
			if createdByName.Valid {
				rep.CreatedByName = createdByName.String
			}
			if internAvatar.Valid {
				rep.InternAvatar = internAvatar.String
			}
			if createdByAvatar.Valid {
				rep.CreatedByAvatar = createdByAvatar.String
			}
			reports = append(reports, rep)
		}
	}

	utils.RespondPaginated(w, reports, utils.CalculatePagination(page, limit, total))
}

func (h *ReportHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	query := `
		SELECT r.id, r.intern_id, r.created_by, r.title, r.content, r.type,
		       r.period_start, r.period_end, r.status, r.feedback, r.created_at, r.updated_at,
		       iu.name, cu.name, iu.avatar, cu.avatar
		FROM reports r
		LEFT JOIN interns i ON r.intern_id = i.id
		LEFT JOIN users iu ON i.user_id = iu.id
		LEFT JOIN users cu ON r.created_by = cu.id
		WHERE r.id = ?
	`

	var rep models.Report
	var feedback sql.NullString
	var internName, createdByName, internAvatar, createdByAvatar sql.NullString
	err := h.db.QueryRow(query, id).Scan(
		&rep.ID, &rep.InternID, &rep.CreatedBy, &rep.Title, &rep.Content, &rep.Type,
		&rep.PeriodStart, &rep.PeriodEnd, &rep.Status, &feedback, &rep.CreatedAt, &rep.UpdatedAt,
		&internName, &createdByName, &internAvatar, &createdByAvatar,
	)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Report not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	if normalizeRole(claims.Role) == "intern" {
		var myInternID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myInternID); err != nil || myInternID != rep.InternID {
			utils.RespondForbidden(w, "You do not have access to this report")
			return
		}
	}

	if feedback.Valid {
		rep.Feedback = feedback.String
	}
	if internName.Valid {
		rep.InternName = internName.String
	}
	if createdByName.Valid {
		rep.CreatedByName = createdByName.String
	}
	if internAvatar.Valid {
		rep.InternAvatar = internAvatar.String
	}
	if createdByAvatar.Valid {
		rep.CreatedByAvatar = createdByAvatar.String
	}

	utils.RespondSuccess(w, "Report retrieved", rep)
}

func (h *ReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	var req createReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// Permission check & ID assignment
	role := normalizeRole(claims.Role)
	if role == "intern" {
		// Interns can only create for themselves
		var internID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&internID); err != nil {
			utils.RespondInternalError(w, "Failed to retrieve intern profile")
			return
		}
		req.InternID = internID
	} else if role != "admin" && role != "pembimbing" {
		// Only admin, supervisor, and intern can create reports
		utils.RespondForbidden(w, "Unauthorized to create reports")
		return
	}

	if req.InternID == 0 || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" || strings.TrimSpace(req.Type) == "" {
		utils.RespondBadRequest(w, "intern_id, title, content, and type are required")
		return
	}
	if req.Type != "weekly" && req.Type != "monthly" && req.Type != "final" {
		utils.RespondBadRequest(w, "Invalid report type")
		return
	}

	start, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid period_start")
		return
	}
	end, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid period_end")
		return
	}

	status := req.Status
	if status == "" {
		status = "submitted"
	}

	// Use 'res' from Exec
	res, err := h.db.Exec(
		`INSERT INTO reports (intern_id, created_by, title, content, type, period_start, period_end, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.InternID, claims.UserID, req.Title, req.Content, req.Type, start, end, status,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to create report")
		return
	}

	reportID, _ := res.LastInsertId()

	// Notify Supervisor (Pembimbing) ONLY if status is submitted
	if status == "submitted" {
		var supervisorID sql.NullInt64
		// Get supervisor of the intern
		err = h.db.QueryRow("SELECT supervisor_id FROM interns WHERE id = ?", req.InternID).Scan(&supervisorID)
		if err != nil {
			log.Printf("[NOTIF_DEBUG] Failed to find supervisor for intern_id %d: %v", req.InternID, err)
		} else if supervisorID.Valid {
			log.Printf("[NOTIF_DEBUG] Found supervisor user_id %d for intern_id %d. Creating notification...", supervisorID.Int64, req.InternID)
			notifErr := createNotification(h.db, supervisorID.Int64, "info", "Laporan Baru",
				"Seorang intern telah membuat laporan "+req.Type+".", "/reports/"+strconv.FormatInt(reportID, 10), nil)
			if notifErr != nil {
				log.Printf("[NOTIF_DEBUG] Failed to create notification for user_id %d: %v", supervisorID.Int64, notifErr)
			} else {
				log.Printf("[NOTIF_DEBUG] Successfully created notification for analyst user_id %d", supervisorID.Int64)
			}
		} else {
			log.Printf("[NOTIF_DEBUG] Intern %d has no supervisor assigned (supervisor_id is NULL)", req.InternID)
		}
	}

	utils.RespondCreated(w, "Report created", nil)
}

func (h *ReportHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	// Check permissions
	if normalizeRole(claims.Role) == "intern" {
		// Verify ownership
		var reportInternID int64
		err := h.db.QueryRow("SELECT intern_id FROM reports WHERE id = ?", id).Scan(&reportInternID)
		if err != nil {
			if err == sql.ErrNoRows {
				utils.RespondNotFound(w, "Report not found")
				return
			}
			utils.RespondInternalError(w, "Database error")
			return
		}

		var myInternID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myInternID); err != nil || myInternID != reportInternID {
			utils.RespondForbidden(w, "You can only update your own reports")
			return
		}
	} else if normalizeRole(claims.Role) != "admin" && normalizeRole(claims.Role) != "pembimbing" {
		utils.RespondForbidden(w, "Unauthorized to update reports")
		return
	}

	var req updateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	updates := []string{}
	args := []interface{}{}

	if req.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Content != nil {
		updates = append(updates, "content = ?")
		args = append(args, *req.Content)
	}
	if req.Type != nil {
		updates = append(updates, "type = ?")
		args = append(args, *req.Type)
	}
	if req.PeriodStart != nil {
		if parsed, err := time.Parse("2006-01-02", *req.PeriodStart); err == nil {
			updates = append(updates, "period_start = ?")
			args = append(args, parsed)
		}
	}
	if req.PeriodEnd != nil {
		if parsed, err := time.Parse("2006-01-02", *req.PeriodEnd); err == nil {
			updates = append(updates, "period_end = ?")
			args = append(args, parsed)
		}
	}
	if req.Status != nil {
		updates = append(updates, "status = ?")
		args = append(args, *req.Status)
	}
	if req.Feedback != nil {
		updates = append(updates, "feedback = ?")
		args = append(args, nullIfEmpty(*req.Feedback))
	}

	if len(updates) == 0 {
		utils.RespondBadRequest(w, "No updates provided")
		return
	}

	// Fetch old status and intern_id for notification logic
	var oldStatus string
	var reportInternID int64
	var reportType string
	_ = h.db.QueryRow("SELECT status, intern_id, type FROM reports WHERE id = ?", id).Scan(&oldStatus, &reportInternID, &reportType)

	args = append(args, id)
	if _, err := h.db.Exec("UPDATE reports SET "+strings.Join(updates, ", ")+" WHERE id = ?", args...); err != nil {
		utils.RespondInternalError(w, "Failed to update report")
		return
	}

	// Check if status changed to 'submitted'
	newStatus := ""
	if req.Status != nil {
		newStatus = *req.Status
	}

	if oldStatus == "draft" && newStatus == "submitted" {
		// Notify Supervisor
		var supervisorID sql.NullInt64
		// Use := for new err variable or reused one if available
		err := h.db.QueryRow("SELECT supervisor_id FROM interns WHERE id = ?", reportInternID).Scan(&supervisorID)
		if err == nil && supervisorID.Valid {
			_ = createNotification(h.db, supervisorID.Int64, "info", "Laporan Baru",
				"Seorang intern telah mengirim laporan "+reportType+".", "/reports/"+strconv.FormatInt(id, 10), nil)
		}
	}

	utils.RespondSuccess(w, "Report updated", nil)
}

func (h *ReportHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	// Check permissions
	if normalizeRole(claims.Role) == "intern" {
		// Verify ownership
		var reportInternID int64
		err := h.db.QueryRow("SELECT intern_id FROM reports WHERE id = ?", id).Scan(&reportInternID)
		if err != nil {
			if err == sql.ErrNoRows {
				utils.RespondNotFound(w, "Report not found")
				return
			}
			utils.RespondInternalError(w, "Database error")
			return
		}

		var myInternID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myInternID); err != nil || myInternID != reportInternID {
			utils.RespondForbidden(w, "You can only delete your own reports")
			return
		}
	} else if normalizeRole(claims.Role) != "admin" && normalizeRole(claims.Role) != "pembimbing" {
		utils.RespondForbidden(w, "Unauthorized to delete reports")
		return
	}

	if _, err := h.db.Exec("DELETE FROM reports WHERE id = ?", id); err != nil {
		utils.RespondInternalError(w, "Failed to delete report")
		return
	}

	utils.RespondSuccess(w, "Report deleted", nil)
}

func (h *ReportHandler) AddFeedback(w http.ResponseWriter, r *http.Request) {
	// claims, ok := middleware.GetUserFromContext(r.Context())
	// if !ok {
	// 	utils.RespondUnauthorized(w, "Unauthorized")
	// 	return
	// }
	// if normalizeRole(claims.Role) == "intern" {
	// 	utils.RespondForbidden(w, "Only admin or pembimbing can add feedback")
	// 	return
	// }

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	var payload struct {
		Feedback string `json:"feedback"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}
	if strings.TrimSpace(payload.Feedback) == "" {
		utils.RespondBadRequest(w, "Feedback is required")
		return
	}

	if _, err := h.db.Exec("UPDATE reports SET feedback = ?, status = 'reviewed' WHERE id = ?", payload.Feedback, id); err != nil {
		utils.RespondInternalError(w, "Failed to add feedback")
		return
	}

	// Notify Intern
	var internUserID int64
	err := h.db.QueryRow(
		`SELECT i.user_id FROM reports r
		 JOIN interns i ON r.intern_id = i.id
		 WHERE r.id = ?`, id,
	).Scan(&internUserID)
	if err == nil {
		_ = createNotification(h.db, internUserID, "info", "Feedback Laporan",
			"Pembimbing telah memberikan feedback pada laporan Anda.", "/reports/"+strconv.FormatInt(id, 10), nil)
	}

	utils.RespondSuccess(w, "Feedback added", nil)
}

// --- Aggregated Reports ---

func (h *ReportHandler) GetInternReport(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	if normalizeRole(claims.Role) == "intern" {
		var myID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myID); err != nil || myID != internID {
			utils.RespondForbidden(w, "You do not have access to this report")
			return
		}
	}

	// Intern info
	var intern struct {
		ID        int64     `json:"id"`
		FullName  string    `json:"full_name"`
		Email     string    `json:"email"`
		StartDate time.Time `json:"start_date"`
		EndDate   time.Time `json:"end_date"`
	}
	if err := h.db.QueryRow(
		`SELECT i.id, i.full_name, u.email, i.start_date, i.end_date
		 FROM interns i JOIN users u ON i.user_id = u.id
		 WHERE i.id = ?`, internID,
	).Scan(&intern.ID, &intern.FullName, &intern.Email, &intern.StartDate, &intern.EndDate); err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	// Task stats
	var taskStats struct {
		Total           int64   `json:"total"`
		Completed       int64   `json:"completed"`
		InProgress      int64   `json:"in_progress"`
		Pending         int64   `json:"pending"`
		Revision        int64   `json:"revision"`
		CompletedOnTime int64   `json:"completed_on_time"`
		CompletedLate   int64   `json:"completed_late"`
		AverageScore    float64 `json:"average_score"`
	}
	_ = h.db.QueryRow(
		`SELECT COUNT(*) as total,
		        SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
		        SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
		        SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
		        SUM(CASE WHEN status = 'revision' THEN 1 ELSE 0 END) as revision,
		        SUM(CASE WHEN status = 'completed' AND is_late = 0 THEN 1 ELSE 0 END) as completed_on_time,
		        SUM(CASE WHEN status = 'completed' AND is_late = 1 THEN 1 ELSE 0 END) as completed_late,
		        AVG(CASE WHEN status = 'completed' THEN score ELSE NULL END) as average_score
		 FROM tasks WHERE intern_id = ?`, internID,
	).Scan(&taskStats.Total, &taskStats.Completed, &taskStats.InProgress, &taskStats.Pending, &taskStats.Revision,
		&taskStats.CompletedOnTime, &taskStats.CompletedLate, &taskStats.AverageScore)

	// Attendance stats
	var attendanceStats struct {
		Total      int64   `json:"total"`
		Present    int64   `json:"present"`
		Late       int64   `json:"late"`
		Absent     int64   `json:"absent"`
		Sick       int64   `json:"sick"`
		Permission int64   `json:"permission"`
		Percentage float64 `json:"percentage"`
	}
	_ = h.db.QueryRow(
		`SELECT COUNT(*) as total,
		        SUM(CASE WHEN status = 'present' THEN 1 ELSE 0 END) as present,
		        SUM(CASE WHEN status = 'late' THEN 1 ELSE 0 END) as late,
		        SUM(CASE WHEN status = 'absent' THEN 1 ELSE 0 END) as absent,
		        SUM(CASE WHEN status = 'sick' THEN 1 ELSE 0 END) as sick,
		        SUM(CASE WHEN status = 'permission' THEN 1 ELSE 0 END) as permission
		 FROM attendances WHERE intern_id = ?`, internID,
	).Scan(&attendanceStats.Total, &attendanceStats.Present, &attendanceStats.Late, &attendanceStats.Absent, &attendanceStats.Sick, &attendanceStats.Permission)

	if attendanceStats.Total > 0 {
		attendanceStats.Percentage = float64(attendanceStats.Present+attendanceStats.Late) / float64(attendanceStats.Total) * 100
	}

	// Assessment stats
	var assessmentStats struct {
		Count         int64   `json:"count"`
		Quality       float64 `json:"quality"`
		Speed         float64 `json:"speed"`
		Initiative    float64 `json:"initiative"`
		Teamwork      float64 `json:"teamwork"`
		Communication float64 `json:"communication"`
		Overall       float64 `json:"overall"`
	}
	_ = h.db.QueryRow(
		`SELECT COUNT(*) as cnt,
		        AVG(quality_score), AVG(speed_score), AVG(initiative_score), AVG(teamwork_score), AVG(communication_score)
		 FROM assessments WHERE intern_id = ?`, internID,
	).Scan(&assessmentStats.Count, &assessmentStats.Quality, &assessmentStats.Speed, &assessmentStats.Initiative, &assessmentStats.Teamwork, &assessmentStats.Communication)

	if assessmentStats.Count > 0 {
		assessmentStats.Overall = (assessmentStats.Quality + assessmentStats.Speed + assessmentStats.Initiative + assessmentStats.Teamwork + assessmentStats.Communication) / 5
	}

	// Recent tasks
	type miniTask struct {
		ID     int64  `json:"id"`
		Title  string `json:"title"`
		Status string `json:"status"`
	}
	var recentTasks []miniTask
	rows, _ := h.db.Query(
		`SELECT id, title, status FROM tasks WHERE intern_id = ? ORDER BY created_at DESC LIMIT 10`,
		internID,
	)
	for rows != nil && rows.Next() {
		var t miniTask
		if err := rows.Scan(&t.ID, &t.Title, &t.Status); err == nil {
			recentTasks = append(recentTasks, t)
		}
	}
	if rows != nil {
		rows.Close()
	}

	// Recent attendances
	type miniAttendance struct {
		ID     int64     `json:"id"`
		Date   time.Time `json:"date"`
		Status string    `json:"status"`
	}
	var recentAttendances []miniAttendance
	rows, _ = h.db.Query(
		`SELECT id, date, status FROM attendances WHERE intern_id = ? ORDER BY date DESC LIMIT 10`,
		internID,
	)
	for rows != nil && rows.Next() {
		var a miniAttendance
		if err := rows.Scan(&a.ID, &a.Date, &a.Status); err == nil {
			recentAttendances = append(recentAttendances, a)
		}
	}
	if rows != nil {
		rows.Close()
	}

	// Progress
	durationDays := int(intern.EndDate.Sub(intern.StartDate).Hours() / 24)
	if durationDays < 1 {
		durationDays = 1
	}
	daysCompleted := int(time.Since(intern.StartDate).Hours() / 24)
	if daysCompleted < 0 {
		daysCompleted = 0
	}
	progress := float64(daysCompleted) / float64(durationDays) * 100
	if progress > 100 {
		progress = 100
	}

	utils.RespondSuccess(w, "Intern report retrieved", map[string]interface{}{
		"intern":             intern,
		"task_stats":         taskStats,
		"attendance_stats":   attendanceStats,
		"assessment_stats":   assessmentStats,
		"recent_tasks":       recentTasks,
		"recent_attendances": recentAttendances,
		"duration_days":      durationDays,
		"days_completed":     daysCompleted,
		"progress":           progress,
	})
}

func (h *ReportHandler) GetAttendanceReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var stats struct {
		Total      int64 `json:"total"`
		Present    int64 `json:"present"`
		Late       int64 `json:"late"`
		Absent     int64 `json:"absent"`
		Sick       int64 `json:"sick"`
		Permission int64 `json:"permission"`
	}
	_ = h.db.QueryRow(
		`SELECT COUNT(*) as total,
		        SUM(CASE WHEN status = 'present' THEN 1 ELSE 0 END) as present,
		        SUM(CASE WHEN status = 'late' THEN 1 ELSE 0 END) as late,
		        SUM(CASE WHEN status = 'absent' THEN 1 ELSE 0 END) as absent,
		        SUM(CASE WHEN status = 'sick' THEN 1 ELSE 0 END) as sick,
		        SUM(CASE WHEN status = 'permission' THEN 1 ELSE 0 END) as permission
		 FROM attendances WHERE intern_id = ?`, internID,
	).Scan(&stats.Total, &stats.Present, &stats.Late, &stats.Absent, &stats.Sick, &stats.Permission)

	utils.RespondSuccess(w, "Attendance report retrieved", stats)
}

func (h *ReportHandler) GetAssessmentReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var avgScore sql.NullFloat64
	_ = h.db.QueryRow("SELECT AVG(score) FROM assessments WHERE intern_id = ?", internID).Scan(&avgScore)

	// Category distribution
	rows, err := h.db.Query("SELECT category, COUNT(*) FROM assessments WHERE intern_id = ? GROUP BY category", internID)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch assessment report")
		return
	}
	defer rows.Close()

	dist := map[string]int64{}
	for rows.Next() {
		var category string
		var count int64
		if err := rows.Scan(&category, &count); err == nil {
			dist[category] = count
		}
	}

	utils.RespondSuccess(w, "Assessment report retrieved", map[string]interface{}{
		"average_score": avgScore.Float64,
		"distribution":  dist,
	})
}

func (h *ReportHandler) GetCertificate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var cert struct {
		ID                int64           `json:"id"`
		InternID          int64           `json:"intern_id"`
		CertificateNumber string          `json:"certificate_number"`
		IssueDate         time.Time       `json:"issue_date"`
		FinalScore        sql.NullFloat64 `json:"final_score"`
		Remarks           sql.NullString  `json:"remarks"`
		FilePath          sql.NullString  `json:"file_path"`
	}
	err := h.db.QueryRow(
		`SELECT id, intern_id, certificate_number, issue_date, final_score, remarks, file_path
		 FROM certificates WHERE intern_id = ?`, internID,
	).Scan(&cert.ID, &cert.InternID, &cert.CertificateNumber, &cert.IssueDate, &cert.FinalScore, &cert.Remarks, &cert.FilePath)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Certificate not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	utils.RespondSuccess(w, "Certificate retrieved", cert)
}

func (h *ReportHandler) GenerateCertificate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	// Calculate final score
	var avgScore sql.NullFloat64
	_ = h.db.QueryRow("SELECT AVG(score) FROM assessments WHERE intern_id = ?", internID).Scan(&avgScore)

	certNumber := "CERT-" + strconv.FormatInt(internID, 10) + "-" + time.Now().Format("20060102")

	_, err := h.db.Exec(
		`INSERT INTO certificates (intern_id, certificate_number, issue_date, final_score)
		 VALUES (?, ?, ?, ?)`,
		internID, certNumber, time.Now(), avgScore,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to generate certificate")
		return
	}

	utils.RespondCreated(w, "Certificate generated", map[string]interface{}{
		"certificate_number": certNumber,
		"final_score":        avgScore.Float64,
	})
}

// DownloadInternReport generates a simple PDF report for an intern
func (h *ReportHandler) DownloadInternReport(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)
	log.Printf(">>> GENERATING INTERN REPORT FOR ID: %d", internID)

	if normalizeRole(claims.Role) == "intern" {
		var myID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myID); err != nil || myID != internID {
			utils.RespondForbidden(w, "You do not have access to this report")
			return
		}
	}

	// Intern info
	var intern struct {
		ID             int64
		FullName       string
		Email          string
		NIS            string
		School         string
		Department     string
		SupervisorName string
		StartDate      time.Time
		EndDate        time.Time
		Avatar         sql.NullString
	}
	err := h.db.QueryRow(
		`SELECT i.id, i.full_name, u.email, i.nis, i.school, i.department, su.name as supervisor_name, i.start_date, i.end_date, u.avatar
		 FROM interns i 
         JOIN users u ON i.user_id = u.id
         LEFT JOIN users su ON i.supervisor_id = su.id
		 WHERE i.id = ?`, internID,
	).Scan(&intern.ID, &intern.FullName, &intern.Email, &intern.NIS, &intern.School, &intern.Department, &intern.SupervisorName, &intern.StartDate, &intern.EndDate, &intern.Avatar)
	if err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	// Recent Tasks
	type taskItem struct {
		Title    string
		Status   string
		Score    sql.NullFloat64
		Deadline sql.NullTime
		IsLate   bool
	}
	var recentTasks []taskItem
	tRows, err := h.db.Query(`
		SELECT title, status, score, deadline, is_late 
		FROM tasks 
		WHERE intern_id = ? 
		ORDER BY created_at DESC LIMIT 10`, internID)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var t taskItem
			tRows.Scan(&t.Title, &t.Status, &t.Score, &t.Deadline, &t.IsLate)
			recentTasks = append(recentTasks, t)
		}
	}

	// Full Attendance History
	type attItem struct {
		Date     time.Time
		Status   string
		CheckIn  sql.NullTime
		CheckOut sql.NullTime
		Notes    sql.NullString
	}
	var recentAtts []attItem
	aRows, err := h.db.Query(`
		SELECT date, status, check_in_time, check_out_time, notes 
		FROM attendances 
		WHERE intern_id = ? 
		ORDER BY date DESC`, internID)
	if err == nil {
		defer aRows.Close()
		for aRows.Next() {
			var a attItem
			aRows.Scan(&a.Date, &a.Status, &a.CheckIn, &a.CheckOut, &a.Notes)
			recentAtts = append(recentAtts, a)
		}
	}

	// Task stats
	var taskStats struct {
		Total           int64
		Completed       int64
		InProgress      int64
		Pending         int64
		Revision        int64
		CompletedOnTime int64
		CompletedLate   int64
		AverageScore    float64
	}
	_ = h.db.QueryRow(
		`SELECT COUNT(*) as total,
		        SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
		        SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
		        SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
		        SUM(CASE WHEN status = 'revision' THEN 1 ELSE 0 END) as revision,
		        SUM(CASE WHEN status = 'completed' AND is_late = 0 THEN 1 ELSE 0 END) as completed_on_time,
		        SUM(CASE WHEN status = 'completed' AND is_late = 1 THEN 1 ELSE 0 END) as completed_late,
		        AVG(CASE WHEN status = 'completed' THEN score ELSE NULL END) as average_score
		 FROM tasks WHERE intern_id = ?`, internID,
	).Scan(&taskStats.Total, &taskStats.Completed, &taskStats.InProgress, &taskStats.Pending, &taskStats.Revision,
		&taskStats.CompletedOnTime, &taskStats.CompletedLate, &taskStats.AverageScore)

	// Attendance stats
	var attendanceStats struct {
		Total      int64
		Present    int64
		Late       int64
		Absent     int64
		Sick       int64
		Permission int64
		Percentage float64
	}
	_ = h.db.QueryRow(
		`SELECT COUNT(*) as total,
		        SUM(CASE WHEN status = 'present' THEN 1 ELSE 0 END) as present,
		        SUM(CASE WHEN status = 'late' THEN 1 ELSE 0 END) as late,
		        SUM(CASE WHEN status = 'absent' THEN 1 ELSE 0 END) as absent,
		        SUM(CASE WHEN status = 'sick' THEN 1 ELSE 0 END) as sick,
		        SUM(CASE WHEN status = 'permission' THEN 1 ELSE 0 END) as permission
		 FROM attendances WHERE intern_id = ?`, internID,
	).Scan(&attendanceStats.Total, &attendanceStats.Present, &attendanceStats.Late, &attendanceStats.Absent, &attendanceStats.Sick, &attendanceStats.Permission)

	if attendanceStats.Total > 0 {
		attendanceStats.Percentage = float64(attendanceStats.Present+attendanceStats.Late) / float64(attendanceStats.Total) * 100
	}

	// Assessment stats
	var assessmentStats struct {
		Count         int64
		Quality       float64
		Speed         float64
		Initiative    float64
		Teamwork      float64
		Communication float64
		Overall       float64
	}
	_ = h.db.QueryRow(
		`SELECT COUNT(*) as cnt,
		        AVG(quality_score), AVG(speed_score), AVG(initiative_score), AVG(teamwork_score), AVG(communication_score)
		 FROM assessments WHERE intern_id = ?`, internID,
	).Scan(&assessmentStats.Count, &assessmentStats.Quality, &assessmentStats.Speed, &assessmentStats.Initiative, &assessmentStats.Teamwork, &assessmentStats.Communication)

	if assessmentStats.Count > 0 {
		assessmentStats.Overall = (assessmentStats.Quality + assessmentStats.Speed + assessmentStats.Initiative + assessmentStats.Teamwork + assessmentStats.Communication) / 5
	}

	durationDays := int(intern.EndDate.Sub(intern.StartDate).Hours() / 24)
	if durationDays < 1 {
		durationDays = 1
	}
	daysCompleted := int(time.Since(intern.StartDate).Hours() / 24)
	if daysCompleted < 0 {
		daysCompleted = 0
	}
	progress := float64(daysCompleted) / float64(durationDays) * 100
	if progress > 100 {
		progress = 100
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// --- Colors (Vibrant Indigo/Purple) ---
	primaryIndigo := []int{79, 70, 229} // Indigo-600
	accentPurple := []int{139, 92, 246} // Violet-500
	textDark := []int{31, 41, 55}       // Gray-800
	textMuted := []int{107, 114, 128}   // Gray-500
	lightGray := []int{209, 213, 219}   // Gray-300
	bgLighter := []int{249, 250, 251}   // Gray-50
	borderGray := []int{229, 231, 235}  // Gray-200

	// Status Colors
	successGreen := []int{34, 197, 94}
	warningAmber := []int{245, 158, 11}
	errorRed := []int{239, 68, 68}
	infoBlue := []int{14, 165, 233}

	// --- Header (Branded) ---
	pdf.SetXY(15, 15)
	pdf.SetTextColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
	pdf.SetFont("Helvetica", "B", 18)
	pdf.Cell(0, 10, "Sistem Manajemen Magang V2")
	pdf.Ln(8)

	pdf.SetX(15)
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(0, 8, "Laporan Kinerja Peserta Magang")
	pdf.Ln(6)

	pdf.SetX(15)
	pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
	pdf.SetFont("Helvetica", "", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Dibuat pada: %s", time.Now().Format("02 January 2006 15:04")))

	// Line separator
	pdf.SetDrawColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
	pdf.SetLineWidth(0.8)
	pdf.Line(15, 38, 195, 38)
	pdf.SetLineWidth(0.2)

	// --- Profile Section ---
	pdf.SetY(50)
	pdf.SetDrawColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
	pdf.Circle(35, 65, 20, "D")

	hasAvatar := false
	if intern.Avatar.Valid && intern.Avatar.String != "" {
		avatarPath := strings.TrimPrefix(intern.Avatar.String, "/")
		if _, err := os.Stat(avatarPath); err == nil {
			pdf.ImageOptions(avatarPath, 15.5, 45.5, 39, 39, false, gofpdf.ImageOptions{ImageType: "", ReadDpi: true}, 0, "")
			hasAvatar = true
		}
	}

	if !hasAvatar {
		// Draw Initial if no avatar
		pdf.SetXY(15, 50)
		initial := "M"
		if len(intern.FullName) > 0 {
			initial = string([]rune(intern.FullName)[0])
		}
		pdf.SetFont("Helvetica", "B", 36)
		pdf.SetTextColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
		pdf.CellFormat(40, 30, initial, "", 0, "C", false, 0, "")
	}

	// Intern details to the right
	rightX := 60.0
	pdf.SetXY(rightX, 52)
	pdf.SetTextColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
	pdf.SetFont("Helvetica", "B", 18)
	pdf.Cell(0, 10, intern.FullName)
	pdf.Ln(10)

	fields := []struct{ k, v string }{
		{"NIS", intern.NIS},
		{"Sekolah", intern.School},
		{"Jurusan", intern.Department},
		{"Pembimbing", intern.SupervisorName},
		{"Periode", fmt.Sprintf("%s - %s", intern.StartDate.Format("02 Jan 2006"), intern.EndDate.Format("02 Jan 2006"))},
	}

	for _, f := range fields {
		pdf.SetX(rightX)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(accentPurple[0], accentPurple[1], accentPurple[2])
		pdf.Cell(25, 5, f.k+": ")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
		pdf.Cell(0, 5, f.v)
		pdf.Ln(5)
	}

	// Status Badge
	pdf.SetXY(rightX, pdf.GetY()+2)
	pdf.SetFillColor(accentPurple[0], accentPurple[1], accentPurple[2])
	pdf.Rect(pdf.GetX(), pdf.GetY(), 25, 6, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(25, 6, "AKTIF", "", 0, "C", false, 0, "")
	pdf.Ln(10)

	// --- Progress Bar ---
	pdf.SetY(105)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
	pdf.Cell(0, 5, fmt.Sprintf("Progress Magang: %.1f%% (%d dari %d hari)", progress, daysCompleted, durationDays))
	pdf.Ln(6)
	pdf.SetFillColor(borderGray[0], borderGray[1], borderGray[2])
	pdf.Rect(15, pdf.GetY(), 180, 2, "F")
	pdf.SetFillColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
	pdf.Rect(15, pdf.GetY(), 180*(progress/100.0), 2, "F")
	pdf.Ln(10)

	// --- Ringkasan Kinerja ---
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.Cell(0, 8, " Ringkasan Kinerja")
	pdf.Ln(8)

	boxW := 43.5
	boxH := 18.0

	pStats := []struct {
		val   string
		label string
	}{
		{fmt.Sprintf("%d/%d", taskStats.Completed, taskStats.Total), "TUGAS SELESAI"},
		{fmt.Sprintf("%.1f%%", attendanceStats.Percentage), "KEHADIRAN"},
		{fmt.Sprintf("%.1f", taskStats.AverageScore), "RATA-RATA NILAI"},
		{fmt.Sprintf("%.1f", assessmentStats.Overall), "SKOR PENILAIAN"},
	}

	for i, s := range pStats {
		xP := 15 + float64(i)*(boxW+2)
		pdf.SetFillColor(bgLighter[0], bgLighter[1], bgLighter[2])
		pdf.Rect(xP, pdf.GetY(), boxW, boxH, "F")

		// Colored side border
		pdf.SetFillColor(accentPurple[0], accentPurple[1], accentPurple[2])
		pdf.Rect(xP, pdf.GetY(), 1.5, boxH, "F")

		pdf.SetXY(xP, pdf.GetY()+2)
		pdf.SetFont("Helvetica", "B", 16)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(boxW, 10, s.val, "", 0, "C", false, 0, "")

		pdf.SetXY(xP, pdf.GetY()+8)
		pdf.SetFont("Helvetica", "B", 6)
		pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
		pdf.CellFormat(boxW, 4, s.label, "", 0, "C", false, 0, "")
	}
	pdf.Ln(boxH + 10)

	// --- 2 Column Stats Detail ---
	startY := pdf.GetY()

	// Column 1: Statistik Tugas
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
	pdf.Cell(88, 8, "STATISTIK TUGAS")

	// Column 2: Statistik Kehadiran
	pdf.SetX(110)
	pdf.Cell(0, 8, "STATISTIK KEHADIRAN")
	pdf.Ln(8)

	// Task Detail Boxes
	subBoxW := 27.5
	subBoxH := 11.0

	taskD := []struct {
		val   int64
		label string
	}{
		{taskStats.Total, "TOTAL"},
		{taskStats.CompletedOnTime, "TEPAT WAKTU"},
		{taskStats.CompletedLate, "TERLAMBAT"},
	}

	for i, d := range taskD {
		xP := 15 + float64(i)*(subBoxW+2)
		pdf.SetFillColor(bgLighter[0], bgLighter[1], bgLighter[2])
		pdf.Rect(xP, pdf.GetY(), subBoxW, subBoxH, "F")

		// Colored side border
		colors := [][]int{primaryIndigo, successGreen, warningAmber}
		pdf.SetFillColor(colors[i][0], colors[i][1], colors[i][2])
		pdf.Rect(xP, pdf.GetY(), 1.2, subBoxH, "F")

		pdf.SetXY(xP, pdf.GetY()+1.5)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
		pdf.CellFormat(subBoxW, 5, fmt.Sprintf("%d", d.val), "", 0, "C", false, 0, "")

		pdf.SetXY(xP, pdf.GetY()+4.5)
		pdf.SetFont("Helvetica", "B", 5)
		pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
		pdf.CellFormat(subBoxW, 4, d.label, "", 0, "C", false, 0, "")
	}

	// Attendance Detail Grid
	pdf.SetXY(110, startY+8)
	attW := 16.5
	attL := []string{"HADIR", "TELAT", "ABSEN", "SAKIT", "IZIN"}
	attV := []int64{attendanceStats.Present, attendanceStats.Late, attendanceStats.Absent, attendanceStats.Sick, attendanceStats.Permission}

	for i, v := range attV {
		xP := 110 + float64(i)*(attW+1.2)
		pdf.SetFillColor(bgLighter[0], bgLighter[1], bgLighter[2])
		pdf.Rect(xP, pdf.GetY(), attW, subBoxH, "F")

		pdf.SetXY(xP, pdf.GetY()+1.5)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
		pdf.CellFormat(attW, 5, fmt.Sprintf("%d", v), "", 0, "C", false, 0, "")

		pdf.SetXY(xP, pdf.GetY()+4.5)
		pdf.SetFont("Helvetica", "B", 4.5)
		pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
		pdf.CellFormat(attW, 4, attL[i], "", 0, "C", false, 0, "")
	}
	pdf.Ln(subBoxH + 10)

	// --- Assessment Radar Style Grid ---
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
	pdf.Cell(0, 8, "PENILAIAN KOMPETENSI")
	pdf.Ln(8)

	compW := 35.5
	compL := []string{"Kualitas Kerja", "Kecepatan", "Inisiatif", "Kerjasama", "Komunikasi"}
	compV := []float64{assessmentStats.Quality, assessmentStats.Speed, assessmentStats.Initiative, assessmentStats.Teamwork, assessmentStats.Communication}

	rowY := pdf.GetY()
	for i, v := range compV {
		xP := 15 + float64(i)*(compW+1)
		pdf.SetFillColor(bgLighter[0], bgLighter[1], bgLighter[2])
		pdf.Rect(xP, rowY, compW, 12, "F")

		pdf.SetXY(xP, rowY+1.5)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetTextColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
		pdf.CellFormat(compW, 6, fmt.Sprintf("%.1f", v), "", 0, "C", false, 0, "")

		pdf.SetXY(xP, rowY+6)
		pdf.SetFont("Helvetica", "B", 5.5)
		pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
		pdf.CellFormat(compW, 4, strings.ToUpper(compL[i]), "", 0, "C", false, 0, "")
	}
	pdf.Ln(18)

	// --- Tables ---
	if pdf.GetY() > 220 {
		pdf.AddPage()
	}

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
	pdf.Cell(0, 8, "DETAIL AKTIVITAS TUGAS")
	pdf.Ln(7)

	pdf.SetFillColor(bgLighter[0], bgLighter[1], bgLighter[2])
	pdf.SetFont("Helvetica", "B", 7)
	pdf.CellFormat(90, 6, "JUDUL TUGAS", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 6, "STATUS", "1", 0, "L", true, 0, "")
	pdf.CellFormat(20, 6, "NILAI", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 6, "DEADLINE", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 6, "WAKTU", "1", 0, "C", true, 0, "")
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "", 7)
	for _, t := range recentTasks {
		pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
		pdf.CellFormat(90, 6, t.Title, "B", 0, "L", false, 0, "")

		st := strings.ToUpper(t.Status)
		if t.Status == "completed" {
			pdf.SetTextColor(successGreen[0], successGreen[1], successGreen[2])
		} else {
			pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
		}
		pdf.CellFormat(25, 6, st, "B", 0, "L", false, 0, "")

		sc := "-"
		if t.Score.Valid {
			sc = fmt.Sprintf("%.0f", t.Score.Float64)
		}
		pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
		pdf.CellFormat(20, 6, sc, "B", 0, "C", false, 0, "")

		dl := "-"
		if t.Deadline.Valid {
			dl = t.Deadline.Time.Format("02/01/06")
		}
		pdf.CellFormat(25, 6, dl, "B", 0, "C", false, 0, "")

		lt := "-"
		if t.Status == "completed" {
			if t.IsLate {
				pdf.SetTextColor(errorRed[0], errorRed[1], errorRed[2])
				lt = "LATE"
			} else {
				pdf.SetTextColor(infoBlue[0], infoBlue[1], infoBlue[2])
				lt = "ON TIME"
			}
		}
		pdf.CellFormat(20, 6, lt, "B", 0, "C", false, 0, "")
		pdf.Ln(6)
	}

	pdf.Ln(10)
	if pdf.GetY() > 220 {
		pdf.AddPage()
	}

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(primaryIndigo[0], primaryIndigo[1], primaryIndigo[2])
	pdf.Cell(0, 8, "RIWAYAT KEHADIRAN")
	pdf.Ln(7)

	pdf.SetFillColor(bgLighter[0], bgLighter[1], bgLighter[2])
	pdf.SetFont("Helvetica", "B", 7)
	pdf.CellFormat(35, 6, "TANGGAL", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 6, "STATUS", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 6, "MASUK", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 6, "KELUAR", "1", 0, "C", true, 0, "")
	pdf.CellFormat(55, 6, "CATATAN", "1", 0, "L", true, 0, "")
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "", 7)
	for _, a := range recentAtts {
		pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
		pdf.CellFormat(35, 6, a.Date.Format("02 Jan 2006"), "B", 0, "L", false, 0, "")

		st := strings.ToUpper(a.Status)
		if a.Status == "present" {
			pdf.SetTextColor(34, 197, 94) // success green
		} else if a.Status == "late" {
			pdf.SetTextColor(245, 158, 11) // warning amber
		} else {
			pdf.SetTextColor(239, 68, 68) // error red
		}
		pdf.CellFormat(30, 6, st, "B", 0, "L", false, 0, "")
		pdf.SetTextColor(textDark[0], textDark[1], textDark[2])

		in := "-"
		if a.CheckIn.Valid {
			in = a.CheckIn.Time.Format("15:04")
		}
		pdf.CellFormat(30, 6, in, "B", 0, "C", false, 0, "")

		out := "-"
		if a.CheckOut.Valid {
			out = a.CheckOut.Time.Format("15:04")
		}
		pdf.CellFormat(30, 6, out, "B", 0, "C", false, 0, "")

		nt := "-"
		if a.Notes.Valid {
			nt = a.Notes.String
			if len(nt) > 40 {
				nt = nt[:37] + "..."
			}
		}
		pdf.CellFormat(55, 6, nt, "B", 0, "L", false, 0, "")
		pdf.Ln(6)
	}

	// Signatures
	pdf.Ln(25)
	if pdf.GetY() > 230 {
		pdf.AddPage()
	}

	sy := pdf.GetY()
	pdf.SetDrawColor(lightGray[0], lightGray[1], lightGray[2])

	// Intern
	pdf.SetXY(15, sy+20)
	pdf.CellFormat(55, 1, "", "T", 0, "C", false, 0, "")
	pdf.SetXY(15, sy+21)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.CellFormat(55, 4, strings.ToUpper(intern.FullName), "", 0, "C", false, 0, "")
	pdf.SetXY(15, sy+25)
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
	pdf.CellFormat(55, 4, "PESERTA MAGANG", "", 0, "C", false, 0, "")

	// Supervisor
	pdf.SetXY(75, sy+20)
	pdf.CellFormat(55, 1, "", "T", 0, "C", false, 0, "")
	pdf.SetXY(75, sy+21)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.CellFormat(55, 4, strings.ToUpper(intern.SupervisorName), "", 0, "C", false, 0, "")
	pdf.SetXY(75, sy+25)
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
	pdf.CellFormat(55, 4, "PEMBIMBING LAPANGAN", "", 0, "C", false, 0, "")

	// Dept Head
	pdf.SetXY(135, sy+20)
	pdf.CellFormat(55, 1, "", "T", 0, "C", false, 0, "")
	pdf.SetXY(135, sy+21)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(textDark[0], textDark[1], textDark[2])
	pdf.CellFormat(55, 4, "____________________", "", 0, "C", false, 0, "")
	pdf.SetXY(135, sy+25)
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(textMuted[0], textMuted[1], textMuted[2])
	pdf.CellFormat(55, 4, "KEPALA DIVISI", "", 0, "C", false, 0, "")

	pdf.SetFont("Helvetica", "I", 6)
	pdf.SetXY(0, 285)
	pdf.CellFormat(210, 8, "Laporan Resmi InternaPro - Dicetak Secara Otomatis oleh Sistem", "", 0, "C", false, 0, "")

	filename := fmt.Sprintf("Laporan_%s_%s.pdf", sanitizeFilename(intern.FullName), time.Now().Format("2006-01-02_150405"))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	if err := pdf.Output(w); err != nil {
		utils.RespondInternalError(w, "Failed to generate PDF")
		return
	}
}

// DownloadCertificate generates a simple internship certificate PDF
func (h *ReportHandler) DownloadCertificate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var intern struct {
		ID                  int64
		FullName            string
		Status              string
		CertificateNumber   sql.NullString
		CertificateIssuedAt sql.NullTime
	}
	if err := h.db.QueryRow(
		`SELECT id, full_name, status, certificate_number, certificate_issued_at
		 FROM interns WHERE id = ?`, internID,
	).Scan(&intern.ID, &intern.FullName, &intern.Status, &intern.CertificateNumber, &intern.CertificateIssuedAt); err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	if intern.Status != "completed" {
		utils.RespondBadRequest(w, "Certificate is only available for completed interns")
		return
	}

	// Fetch or generate certificate record
	var certNumber string
	var issueDate time.Time
	var finalScore sql.NullFloat64
	err := h.db.QueryRow(
		`SELECT certificate_number, issue_date, final_score FROM certificates WHERE intern_id = ?`,
		internID,
	).Scan(&certNumber, &issueDate, &finalScore)
	if err == sql.ErrNoRows {
		year := time.Now().Format("2006")
		certNumber = fmt.Sprintf("MG-DSI/%s/%04d", year, internID)
		issueDate = time.Now()
		_ = h.db.QueryRow("SELECT AVG(score) FROM assessments WHERE intern_id = ?", internID).Scan(&finalScore)

		if _, err := h.db.Exec(
			`INSERT INTO certificates (intern_id, certificate_number, issue_date, final_score)
			 VALUES (?, ?, ?, ?)`, internID, certNumber, issueDate, finalScore,
		); err != nil {
			utils.RespondInternalError(w, "Failed to generate certificate")
			return
		}

		_, _ = h.db.Exec(
			`UPDATE interns SET certificate_number = ?, certificate_issued_at = ? WHERE id = ?`,
			certNumber, issueDate, internID,
		)
	} else if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	if !intern.CertificateNumber.Valid {
		intern.CertificateNumber = sql.NullString{String: certNumber, Valid: true}
		intern.CertificateIssuedAt = sql.NullTime{Time: issueDate, Valid: true}
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 26)
	pdf.Cell(0, 20, "Sertifikat Magang")
	pdf.Ln(14)

	pdf.SetFont("Helvetica", "", 14)
	pdf.Cell(0, 10, "Diberikan kepada:")
	pdf.Ln(10)
	pdf.SetFont("Helvetica", "B", 22)
	pdf.Cell(0, 14, intern.FullName)
	pdf.Ln(12)

	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(0, 8, fmt.Sprintf("Nomor Sertifikat: %s", certNumber))
	pdf.Ln(6)
	if intern.CertificateIssuedAt.Valid {
		pdf.Cell(0, 8, fmt.Sprintf("Tanggal: %s", intern.CertificateIssuedAt.Time.Format("02/01/2006")))
	}
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "", 11)
	if finalScore.Valid {
		pdf.Cell(0, 6, fmt.Sprintf("Nilai Akhir: %.1f", finalScore.Float64))
	}

	filename := fmt.Sprintf("Sertifikat_%s_%s.pdf", sanitizeFilename(intern.FullName), time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	if err := pdf.Output(w); err != nil {
		utils.RespondInternalError(w, "Failed to generate PDF")
		return
	}
}

func sanitizeFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "file"
	}
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		if r == ' ' {
			return '_'
		}
		return '_'
	}, value)
}
