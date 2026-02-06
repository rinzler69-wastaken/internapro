package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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
		       iu.name, cu.name
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
		var internName, createdByName sql.NullString
		if err := rows.Scan(
			&rep.ID, &rep.InternID, &rep.CreatedBy, &rep.Title, &rep.Content, &rep.Type,
			&rep.PeriodStart, &rep.PeriodEnd, &rep.Status, &feedback, &rep.CreatedAt, &rep.UpdatedAt,
			&internName, &createdByName,
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
		       iu.name, cu.name
		FROM reports r
		LEFT JOIN interns i ON r.intern_id = i.id
		LEFT JOIN users iu ON i.user_id = iu.id
		LEFT JOIN users cu ON r.created_by = cu.id
		WHERE r.id = ?
	`

	var rep models.Report
	var feedback sql.NullString
	var internName, createdByName sql.NullString
	err := h.db.QueryRow(query, id).Scan(
		&rep.ID, &rep.InternID, &rep.CreatedBy, &rep.Title, &rep.Content, &rep.Type,
		&rep.PeriodStart, &rep.PeriodEnd, &rep.Status, &feedback, &rep.CreatedAt, &rep.UpdatedAt,
		&internName, &createdByName,
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

	utils.RespondSuccess(w, "Report retrieved", rep)
}

func (h *ReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can create reports")
		return
	}

	var req createReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
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

	_, err = h.db.Exec(
		`INSERT INTO reports (intern_id, created_by, title, content, type, period_start, period_end, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 'submitted')`,
		req.InternID, claims.UserID, req.Title, req.Content, req.Type, start, end,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to create report")
		return
	}

	utils.RespondCreated(w, "Report created", nil)
}

func (h *ReportHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can update reports")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

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

	args = append(args, id)
	if _, err := h.db.Exec("UPDATE reports SET "+strings.Join(updates, ", ")+" WHERE id = ?", args...); err != nil {
		utils.RespondInternalError(w, "Failed to update report")
		return
	}

	utils.RespondSuccess(w, "Report updated", nil)
}

func (h *ReportHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can delete reports")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	if _, err := h.db.Exec("DELETE FROM reports WHERE id = ?", id); err != nil {
		utils.RespondInternalError(w, "Failed to delete report")
		return
	}

	utils.RespondSuccess(w, "Report deleted", nil)
}

func (h *ReportHandler) AddFeedback(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can add feedback")
		return
	}

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

	if normalizeRole(claims.Role) == "intern" {
		var myID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myID); err != nil || myID != internID {
			utils.RespondForbidden(w, "You do not have access to this report")
			return
		}
	}

	// Intern info
	var intern struct {
		ID        int64
		FullName  string
		Email     string
		StartDate time.Time
		EndDate   time.Time
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
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(0, 10, "Laporan Intern")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Nama: %s", intern.FullName))
	pdf.Ln(5)
	pdf.Cell(0, 6, fmt.Sprintf("Email: %s", intern.Email))
	pdf.Ln(5)
	pdf.Cell(0, 6, fmt.Sprintf("Periode: %s - %s", intern.StartDate.Format("02/01/2006"), intern.EndDate.Format("02/01/2006")))
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 7, "Ringkasan Tugas")
	pdf.Ln(6)
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(0, 5, fmt.Sprintf("Total: %d | Selesai: %d | Dalam Proses: %d | Pending: %d | Revisi: %d", taskStats.Total, taskStats.Completed, taskStats.InProgress, taskStats.Pending, taskStats.Revision))
	pdf.Ln(5)
	pdf.Cell(0, 5, fmt.Sprintf("Tepat Waktu: %d | Terlambat: %d | Rata-rata Nilai: %.1f", taskStats.CompletedOnTime, taskStats.CompletedLate, taskStats.AverageScore))
	pdf.Ln(7)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 7, "Ringkasan Presensi")
	pdf.Ln(6)
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(0, 5, fmt.Sprintf("Total: %d | Hadir: %d | Terlambat: %d | Tidak Hadir: %d | Sakit: %d | Izin: %d", attendanceStats.Total, attendanceStats.Present, attendanceStats.Late, attendanceStats.Absent, attendanceStats.Sick, attendanceStats.Permission))
	pdf.Ln(5)
	pdf.Cell(0, 5, fmt.Sprintf("Kehadiran: %.1f%%", attendanceStats.Percentage))
	pdf.Ln(7)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 7, "Ringkasan Penilaian")
	pdf.Ln(6)
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(0, 5, fmt.Sprintf("Kualitas: %.1f | Kecepatan: %.1f | Inisiatif: %.1f | Kerjasama: %.1f | Komunikasi: %.1f", assessmentStats.Quality, assessmentStats.Speed, assessmentStats.Initiative, assessmentStats.Teamwork, assessmentStats.Communication))
	pdf.Ln(5)
	pdf.Cell(0, 5, fmt.Sprintf("Skor Keseluruhan: %.1f", assessmentStats.Overall))
	pdf.Ln(7)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 7, "Progress Magang")
	pdf.Ln(6)
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(0, 5, fmt.Sprintf("Durasi: %d hari | Hari berjalan: %d | Progress: %.1f%%", durationDays, daysCompleted, progress))

	filename := fmt.Sprintf("Laporan_%s_%s.pdf", sanitizeFilename(intern.FullName), time.Now().Format("2006-01-02"))
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
