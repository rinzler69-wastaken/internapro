package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/gorilla/mux"
)

type TaskHandler struct {
	db *sql.DB
}

func NewTaskHandler(db *sql.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

type createTaskRequest struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Priority     string  `json:"priority"`      // low, medium, high
	StartDate    string  `json:"start_date"`    // YYYY-MM-DD
	Deadline     string  `json:"deadline"`      // YYYY-MM-DD
	DeadlineTime string  `json:"deadline_time"` // HH:MM
	AssignTo     string  `json:"assign_to"`     // all, selected
	InternIDs    []int64 `json:"intern_ids"`
}

type updateTaskRequest struct {
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	InternID      *int64  `json:"intern_id"`
	Priority      *string `json:"priority"`
	Status        *string `json:"status"`
	StartDate     *string `json:"start_date"`
	Deadline      *string `json:"deadline"`
	DeadlineTime  *string `json:"deadline_time"`
	AdminFeedback *string `json:"admin_feedback"`
}

type submitTaskRequest struct {
	SubmissionNotes string                  `json:"submission_notes"`
	Links           []models.SubmissionLink `json:"links"`
}

type reviewTaskRequest struct {
	Action   string `json:"action"` // approve, revision
	Score    *int   `json:"score"`
	Feedback string `json:"feedback"`
}

type updateStatusRequest struct {
	Status string `json:"status"` // pending, in_progress
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
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

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	priority := strings.TrimSpace(r.URL.Query().Get("priority"))
	internFilter := strings.TrimSpace(r.URL.Query().Get("intern_id"))

	var where []string
	var args []interface{}

	role := normalizeRole(claims.Role)

	// Interns can only see their tasks and not scheduled tasks
	if role == "intern" {
		internID, err := h.getInternIDForUser(claims.UserID)
		if err != nil {
			utils.RespondNotFound(w, "Intern not found")
			return
		}
		where = append(where, "t.intern_id = ?")
		args = append(args, internID)
		where = append(where, "t.status != 'scheduled'")
	} else {
		// Supervisors/pembimbing only see tasks for their own interns
		if role == "pembimbing" {
			where = append(where, "i.supervisor_id = ?")
			args = append(args, claims.UserID)
		}
		if internFilter != "" {
			if id, err := strconv.ParseInt(internFilter, 10, 64); err == nil {
				where = append(where, "t.intern_id = ?")
				args = append(args, id)
			}
		}
	}

	if search != "" {
		where = append(where, "t.title LIKE ?")
		args = append(args, "%"+search+"%")
	}
	if status != "" {
		where = append(where, "t.status = ?")
		args = append(args, status)
	}
	if priority != "" {
		where = append(where, "t.priority = ?")
		args = append(args, priority)
	}

	baseFrom := `
		FROM tasks t
		LEFT JOIN interns i ON t.intern_id = i.id
		LEFT JOIN users iu ON i.user_id = iu.id
		LEFT JOIN users au ON t.assigned_by = au.id
	`

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	var total int64
	countQuery := "SELECT COUNT(*) " + baseFrom + " " + whereClause
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		utils.RespondInternalError(w, "Failed to count tasks")
		return
	}

	query := `
		SELECT t.id, t.task_assignment_id, t.intern_id, t.assigned_by, t.title, t.description, t.priority, t.status,
		       t.start_date, t.deadline, t.deadline_time, t.started_at, t.submitted_at, t.completed_at, t.approved_at,
		       t.is_late, t.submission_notes, t.submission_links, t.score, t.admin_feedback, t.created_at, t.updated_at,
		       iu.name, au.name
	` + baseFrom + " " + whereClause + " ORDER BY t.created_at DESC LIMIT ? OFFSET ?"

	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch tasks")
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	utils.RespondPaginated(w, tasks, utils.CalculatePagination(page, limit, total))
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	query := `
		SELECT t.id, t.task_assignment_id, t.intern_id, t.assigned_by, t.title, t.description, t.priority, t.status,
		       t.start_date, t.deadline, t.deadline_time, t.started_at, t.submitted_at, t.completed_at, t.approved_at,
		       t.is_late, t.submission_notes, t.submission_links, t.score, t.admin_feedback, t.created_at, t.updated_at,
		       iu.name, au.name
		FROM tasks t
		LEFT JOIN interns i ON t.intern_id = i.id
		LEFT JOIN users iu ON i.user_id = iu.id
		LEFT JOIN users au ON t.assigned_by = au.id
		WHERE t.id = ?
	`

	t, err := scanTask(h.db.QueryRow(query, id))
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Task not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	// Interns can only access their own tasks
	if normalizeRole(claims.Role) == "intern" {
		internID, err := h.getInternIDForUser(claims.UserID)
		if err != nil || t.InternID == nil || *t.InternID != internID {
			utils.RespondForbidden(w, "You do not have access to this task")
			return
		}
	}

	utils.RespondSuccess(w, "Task retrieved", t)
}

func (h *TaskHandler) GetByInternID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	q := r.URL.Query()
	q.Set("intern_id", strconv.FormatInt(internID, 10))
	r.URL.RawQuery = q.Encode()
	h.GetAll(w, r)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// Only admin/pembimbing can create tasks
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can assign tasks")
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		utils.RespondBadRequest(w, "Title is required")
		return
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}
	if req.Priority != "low" && req.Priority != "medium" && req.Priority != "high" {
		utils.RespondBadRequest(w, "Invalid priority")
		return
	}
	if req.AssignTo == "" {
		req.AssignTo = "selected"
	}
	if req.AssignTo != "all" && req.AssignTo != "selected" {
		utils.RespondBadRequest(w, "Invalid assign_to value")
		return
	}

	startDate := time.Now()
	if req.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = parsed
		}
	}
	var deadline sql.NullTime
	if req.Deadline != "" {
		if parsed, err := time.Parse("2006-01-02", req.Deadline); err == nil {
			deadline = sql.NullTime{Time: parsed, Valid: true}
		}
	}
	var deadlineTime sql.NullString
	if req.DeadlineTime != "" {
		deadlineTime = sql.NullString{String: req.DeadlineTime + ":00", Valid: true}
	}

	isScheduled := startDate.After(time.Now())
	initialStatus := "pending"
	if isScheduled {
		initialStatus = "scheduled"
	}

	// Determine interns to assign
	type internRow struct {
		ID     int64
		UserID int64
	}
	var interns []internRow

	if req.AssignTo == "all" {
		rows, err := h.db.Query("SELECT id, user_id FROM interns WHERE status = 'active'")
		if err != nil {
			utils.RespondInternalError(w, "Failed to fetch interns")
			return
		}
		defer rows.Close()
		for rows.Next() {
			var it internRow
			if err := rows.Scan(&it.ID, &it.UserID); err == nil {
				interns = append(interns, it)
			}
		}
	} else {
		if len(req.InternIDs) == 0 {
			utils.RespondBadRequest(w, "intern_ids is required when assign_to is selected")
			return
		}
		query := "SELECT id, user_id FROM interns WHERE id IN (" + placeholders(len(req.InternIDs)) + ")"
		args := make([]interface{}, 0, len(req.InternIDs))
		for _, id := range req.InternIDs {
			args = append(args, id)
		}
		rows, err := h.db.Query(query, args...)
		if err != nil {
			utils.RespondInternalError(w, "Failed to fetch interns")
			return
		}
		defer rows.Close()
		for rows.Next() {
			var it internRow
			if err := rows.Scan(&it.ID, &it.UserID); err == nil {
				interns = append(interns, it)
			}
		}
	}

	if len(interns) == 0 {
		utils.RespondBadRequest(w, "No interns found for assignment")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	assignmentRes, err := tx.Exec(
		`INSERT INTO task_assignments (title, description, assigned_by, priority, start_date, deadline, deadline_time, assign_to_all)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.Title, nullIfEmpty(req.Description), claims.UserID, req.Priority, startDate, deadline, deadlineTime, req.AssignTo == "all",
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to create task assignment")
		return
	}
	assignmentID, _ := assignmentRes.LastInsertId()

	for _, it := range interns {
		if _, err := tx.Exec(
			"INSERT INTO task_assignment_interns (task_assignment_id, intern_id) VALUES (?, ?)",
			assignmentID, it.ID,
		); err != nil {
			utils.RespondInternalError(w, "Failed to attach interns")
			return
		}

		taskRes, err := tx.Exec(
			`INSERT INTO tasks (task_assignment_id, title, description, intern_id, assigned_by, priority, status, start_date, deadline, deadline_time)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			assignmentID, req.Title, nullIfEmpty(req.Description), it.ID, claims.UserID, req.Priority, initialStatus, startDate, deadline, deadlineTime,
		)
		if err != nil {
			utils.RespondInternalError(w, "Failed to create tasks")
			return
		}

		taskID, _ := taskRes.LastInsertId()

		if !isScheduled {
			_ = h.createNotification(it.UserID, models.NotificationTaskAssigned, "Tugas Baru: "+req.Title,
				"Anda mendapat tugas baru. Silakan cek detail tugas Anda.", "/tasks/"+strconv.FormatInt(taskID, 10), map[string]interface{}{"task_id": taskID})
		}
	}

	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to commit transaction")
		return
	}

	utils.RespondCreated(w, "Tasks created successfully", map[string]interface{}{
		"assignment_id": assignmentID,
		"count":         len(interns),
		"scheduled":     isScheduled,
	})
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can update tasks")
		return
	}

	vars := mux.Vars(r)
	taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// Load current task for status transitions
	var current struct {
		Status       string
		StartedAt    sql.NullTime
		Deadline     sql.NullTime
		DeadlineTime sql.NullString
	}
	err := h.db.QueryRow("SELECT status, started_at, deadline, deadline_time FROM tasks WHERE id = ?", taskID).
		Scan(&current.Status, &current.StartedAt, &current.Deadline, &current.DeadlineTime)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Task not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	updates := []string{}
	args := []interface{}{}

	if req.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		updates = append(updates, "description = ?")
		args = append(args, nullIfEmpty(*req.Description))
	}
	if req.InternID != nil {
		updates = append(updates, "intern_id = ?")
		args = append(args, sql.NullInt64{Int64: *req.InternID, Valid: *req.InternID != 0})
	}
	if req.Priority != nil {
		updates = append(updates, "priority = ?")
		args = append(args, *req.Priority)
	}
	if req.StartDate != nil {
		if parsed, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			updates = append(updates, "start_date = ?")
			args = append(args, parsed)
		}
	}
	if req.Deadline != nil {
		if parsed, err := time.Parse("2006-01-02", *req.Deadline); err == nil {
			updates = append(updates, "deadline = ?")
			args = append(args, parsed)
		}
	}
	if req.DeadlineTime != nil {
		if *req.DeadlineTime == "" {
			updates = append(updates, "deadline_time = NULL")
		} else {
			updates = append(updates, "deadline_time = ?")
			args = append(args, *req.DeadlineTime+":00")
		}
	}
	if req.AdminFeedback != nil {
		updates = append(updates, "admin_feedback = ?")
		args = append(args, nullIfEmpty(*req.AdminFeedback))
	}

	if req.Status != nil && *req.Status != current.Status {
		updates = append(updates, "status = ?")
		args = append(args, *req.Status)
		if *req.Status == "in_progress" && !current.StartedAt.Valid {
			updates = append(updates, "started_at = ?")
			args = append(args, time.Now())
		}
		if *req.Status == "completed" {
			updates = append(updates, "completed_at = ?")
			args = append(args, time.Now())
			updates = append(updates, "submitted_at = ?")
			args = append(args, time.Now())
			if h.isLate(current.Deadline, current.DeadlineTime, time.Now()) {
				updates = append(updates, "is_late = 1")
			}
		}
	}

	if len(updates) == 0 {
		utils.RespondBadRequest(w, "No updates provided")
		return
	}

	args = append(args, taskID)
	query := "UPDATE tasks SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	if _, err := h.db.Exec(query, args...); err != nil {
		utils.RespondInternalError(w, "Failed to update task")
		return
	}

	utils.RespondSuccess(w, "Task updated", nil)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can delete tasks")
		return
	}

	vars := mux.Vars(r)
	taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

	if _, err := h.db.Exec("DELETE FROM tasks WHERE id = ?", taskID); err != nil {
		utils.RespondInternalError(w, "Failed to delete task")
		return
	}

	utils.RespondSuccess(w, "Task deleted", nil)
}

func (h *TaskHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can upload attachments")
		return
	}

	vars := mux.Vars(r)
	taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondBadRequest(w, "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.RespondBadRequest(w, "Missing file")
		return
	}
	defer file.Close()

	path, err := utils.UploadFile(file, header, "tasks")
	if err != nil {
		utils.RespondInternalError(w, "Upload failed: "+err.Error())
		return
	}

	_, err = h.db.Exec(
		`INSERT INTO task_attachments (task_id, file_name, file_path, file_type, file_size)
		 VALUES (?, ?, ?, ?, ?)`,
		taskID, header.Filename, path, utils.GetFileExtension(header.Filename), header.Size,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to save attachment")
		return
	}

	utils.RespondCreated(w, "Attachment uploaded", map[string]string{"path": path})
}

func (h *TaskHandler) MarkComplete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can mark complete")
		return
	}

	vars := mux.Vars(r)
	taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

	_, err := h.db.Exec("UPDATE tasks SET status = 'completed', completed_at = ? WHERE id = ?", time.Now(), taskID)
	if err != nil {
		utils.RespondInternalError(w, "Failed to update task")
		return
	}

	utils.RespondSuccess(w, "Task marked complete", nil)
}

// Submit task by intern
func (h *TaskHandler) Submit(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) != "intern" {
		utils.RespondForbidden(w, "Only interns can submit tasks")
		return
	}

	vars := mux.Vars(r)
	taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

	internID, err := h.getInternIDForUser(claims.UserID)
	if err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	var req submitTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}
	if len(req.Links) == 0 {
		utils.RespondBadRequest(w, "At least one submission link is required")
		return
	}
	for _, l := range req.Links {
		if strings.TrimSpace(l.Label) == "" || strings.TrimSpace(l.URL) == "" {
			utils.RespondBadRequest(w, "Each link must have label and url")
			return
		}
	}

	var deadline sql.NullTime
	var deadlineTime sql.NullString
	err = h.db.QueryRow("SELECT deadline, deadline_time FROM tasks WHERE id = ? AND intern_id = ?", taskID, internID).
		Scan(&deadline, &deadlineTime)
	if err == sql.ErrNoRows {
		utils.RespondForbidden(w, "You do not have access to this task")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	linksJSON, _ := json.Marshal(req.Links)
	now := time.Now()
	isLate := h.isLate(deadline, deadlineTime, now)

	_, err = h.db.Exec(
		`UPDATE tasks SET status = 'submitted', submitted_at = ?, submission_notes = ?, submission_links = ?, is_late = ?,
		        started_at = COALESCE(started_at, ?)
		 WHERE id = ?`,
		now, nullIfEmpty(req.SubmissionNotes), string(linksJSON), isLate, now, taskID,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to submit task")
		return
	}

	utils.RespondSuccess(w, "Task submitted", map[string]interface{}{"is_late": isLate})
}

// Review task by admin/pembimbing
func (h *TaskHandler) Review(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can review tasks")
		return
	}

	vars := mux.Vars(r)
	taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var req reviewTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}
	if req.Action != "approve" && req.Action != "revision" {
		utils.RespondBadRequest(w, "Invalid action")
		return
	}
	if req.Action == "approve" && req.Score == nil {
		utils.RespondBadRequest(w, "Score is required for approval")
		return
	}

	var internUserID int64
	err := h.db.QueryRow(
		`SELECT i.user_id FROM tasks t
		 LEFT JOIN interns i ON t.intern_id = i.id
		 WHERE t.id = ?`, taskID,
	).Scan(&internUserID)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Task not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	if req.Action == "approve" {
		_, err = h.db.Exec(
			`UPDATE tasks SET status = 'completed', completed_at = ?, approved_at = ?, score = ?, admin_feedback = ?
			 WHERE id = ?`,
			time.Now(), time.Now(), *req.Score, nullIfEmpty(req.Feedback), taskID,
		)
		if err != nil {
			utils.RespondInternalError(w, "Failed to approve task")
			return
		}
		_ = h.createNotification(internUserID, models.NotificationTaskApproved, "Tugas Disetujui",
			"Tugas Anda telah disetujui. Nilai: "+strconv.Itoa(*req.Score), "/tasks/"+strconv.FormatInt(taskID, 10),
			map[string]interface{}{"task_id": taskID, "score": *req.Score})
	} else {
		_, err = h.db.Exec(
			`UPDATE tasks SET status = 'revision', admin_feedback = ?, score = NULL, approved_at = ?
			 WHERE id = ?`,
			nullIfEmpty(req.Feedback), time.Now(), taskID,
		)
		if err != nil {
			utils.RespondInternalError(w, "Failed to request revision")
			return
		}
		_ = h.createNotification(internUserID, models.NotificationTaskRevision, "Perlu Revisi",
			"Tugas Anda memerlukan revisi. Silakan cek feedback pembimbing.", "/tasks/"+strconv.FormatInt(taskID, 10),
			map[string]interface{}{"task_id": taskID})
	}

	utils.RespondSuccess(w, "Review processed", nil)
}

// Update status by intern (pending/in_progress)
func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) != "intern" {
		utils.RespondForbidden(w, "Only interns can update task status")
		return
	}

	vars := mux.Vars(r)
	taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}
	if req.Status != "pending" && req.Status != "in_progress" {
		utils.RespondBadRequest(w, "Invalid status")
		return
	}

	internID, err := h.getInternIDForUser(claims.UserID)
	if err != nil {
		utils.RespondNotFound(w, "Intern not found")
		return
	}

	query := "UPDATE tasks SET status = ?"
	args := []interface{}{req.Status}
	if req.Status == "in_progress" {
		query += ", started_at = COALESCE(started_at, ?)"
		args = append(args, time.Now())
	}
	query += " WHERE id = ? AND intern_id = ?"
	args = append(args, taskID, internID)

	if _, err := h.db.Exec(query, args...); err != nil {
		utils.RespondInternalError(w, "Failed to update status")
		return
	}

	utils.RespondSuccess(w, "Status updated", nil)
}

// Search interns (for task assignment)
func (h *TaskHandler) SearchInterns(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	args := []interface{}{}
	where := "WHERE i.status = 'active'"

	if query != "" {
		where += " AND (u.name LIKE ? OR i.school LIKE ? OR i.department LIKE ?)"
		like := "%" + query + "%"
		args = append(args, like, like, like)
	}

	rows, err := h.db.Query(
		`SELECT i.id, u.name, i.school, i.department
		 FROM interns i
		 LEFT JOIN users u ON i.user_id = u.id
		`+where+` LIMIT 20`, args...,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to search interns")
		return
	}
	defer rows.Close()

	type result struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		School     string `json:"school"`
		Department string `json:"department"`
		Label      string `json:"label"`
	}

	results := []result{}
	for rows.Next() {
		var r result
		if err := rows.Scan(&r.ID, &r.Name, &r.School, &r.Department); err == nil {
			r.Label = r.Name + " - " + r.School + " (" + r.Department + ")"
			results = append(results, r)
		}
	}

	utils.RespondSuccess(w, "Interns found", results)
}

// Task Assignments (grouped)
func (h *TaskHandler) GetAssignments(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can view assignments")
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

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	priority := strings.TrimSpace(r.URL.Query().Get("priority"))

	where := []string{}
	args := []interface{}{}
	if search != "" {
		where = append(where, "ta.title LIKE ?")
		args = append(args, "%"+search+"%")
	}
	if priority != "" {
		where = append(where, "ta.priority = ?")
		args = append(args, priority)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	var total int64
	countQuery := `
		SELECT COUNT(*) FROM task_assignments ta
	` + whereClause
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		utils.RespondInternalError(w, "Failed to count assignments")
		return
	}

	query := `
		SELECT ta.id, ta.title, ta.description, ta.assigned_by, ta.priority, ta.start_date, ta.deadline, ta.deadline_time,
		       ta.assign_to_all, ta.created_at, ta.updated_at, u.name, COUNT(t.id) as tasks_count
		FROM task_assignments ta
		LEFT JOIN tasks t ON t.task_assignment_id = ta.id
		LEFT JOIN users u ON ta.assigned_by = u.id
	` + whereClause + `
		GROUP BY ta.id
		HAVING tasks_count > 0
		ORDER BY ta.created_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch assignments")
		return
	}
	defer rows.Close()

	type assignmentWithStats struct {
		models.TaskAssignment
		Stats map[string]interface{} `json:"stats"`
	}

	var assignments []assignmentWithStats
	for rows.Next() {
		var a models.TaskAssignment
		var description sql.NullString
		var startDate sql.NullTime
		var deadline sql.NullTime
		var deadlineTime sql.NullString
		var assignedByName sql.NullString
		if err := rows.Scan(
			&a.ID, &a.Title, &description, &a.AssignedBy, &a.Priority, &startDate, &deadline, &deadlineTime,
			&a.AssignToAll, &a.CreatedAt, &a.UpdatedAt, &assignedByName, &a.TasksCount,
		); err == nil {
			a.Description = ptrStringFromNull(description)
			a.StartDate = ptrTimeFromNull(startDate)
			a.Deadline = ptrTimeFromNull(deadline)
			a.DeadlineTime = ptrStringFromNull(deadlineTime)
			if assignedByName.Valid {
				a.AssignedByName = assignedByName.String
			}
			stats := h.assignmentStats(a.ID)
			assignments = append(assignments, assignmentWithStats{TaskAssignment: a, Stats: stats})
		}
	}

	utils.RespondPaginated(w, assignments, utils.CalculatePagination(page, limit, total))
}

func (h *TaskHandler) GetAssignmentByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can view assignments")
		return
	}

	vars := mux.Vars(r)
	assignmentID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var assignment models.TaskAssignment
	var description sql.NullString
	var startDate sql.NullTime
	var deadline sql.NullTime
	var deadlineTime sql.NullString
	var assignedByName sql.NullString
	err := h.db.QueryRow(
		`SELECT ta.id, ta.title, ta.description, ta.assigned_by, ta.priority, ta.start_date, ta.deadline, ta.deadline_time,
		        ta.assign_to_all, ta.created_at, ta.updated_at, u.name
		 FROM task_assignments ta
		 LEFT JOIN users u ON ta.assigned_by = u.id
		 WHERE ta.id = ?`, assignmentID,
	).Scan(&assignment.ID, &assignment.Title, &description, &assignment.AssignedBy, &assignment.Priority,
		&startDate, &deadline, &deadlineTime, &assignment.AssignToAll,
		&assignment.CreatedAt, &assignment.UpdatedAt, &assignedByName)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Assignment not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}
	assignment.Description = ptrStringFromNull(description)
	assignment.StartDate = ptrTimeFromNull(startDate)
	assignment.Deadline = ptrTimeFromNull(deadline)
	assignment.DeadlineTime = ptrStringFromNull(deadlineTime)
	if assignedByName.Valid {
		assignment.AssignedByName = assignedByName.String
	}

	// Fetch tasks under this assignment
	rows, err := h.db.Query(
		`SELECT t.id, t.task_assignment_id, t.intern_id, t.assigned_by, t.title, t.description, t.priority, t.status,
		        t.start_date, t.deadline, t.deadline_time, t.started_at, t.submitted_at, t.completed_at, t.approved_at,
		        t.is_late, t.submission_notes, t.submission_links, t.score, t.admin_feedback, t.created_at, t.updated_at,
		        iu.name, au.name
		 FROM tasks t
		 LEFT JOIN interns i ON t.intern_id = i.id
		 LEFT JOIN users iu ON i.user_id = iu.id
		 LEFT JOIN users au ON t.assigned_by = au.id
		 WHERE t.task_assignment_id = ?`, assignmentID,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch tasks")
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	stats := h.assignmentStats(assignmentID)
	utils.RespondSuccess(w, "Assignment retrieved", map[string]interface{}{
		"assignment": assignment,
		"tasks":      tasks,
		"stats":      stats,
	})
}

// Helpers
type sqlScanner interface {
	Scan(dest ...interface{}) error
}

func scanTask(scanner sqlScanner) (models.Task, error) {
	var t models.Task
	var taskAssignmentID sql.NullInt64
	var internID sql.NullInt64
	var description sql.NullString
	var startDate sql.NullTime
	var deadline sql.NullTime
	var deadlineTime sql.NullString
	var startedAt sql.NullTime
	var submittedAt sql.NullTime
	var completedAt sql.NullTime
	var approvedAt sql.NullTime
	var submissionNotes sql.NullString
	var submissionLinks sql.NullString
	var score sql.NullInt64
	var adminFeedback sql.NullString
	var internName sql.NullString
	var assignedByName sql.NullString

	if err := scanner.Scan(
		&t.ID, &taskAssignmentID, &internID, &t.AssignedBy, &t.Title, &description, &t.Priority, &t.Status,
		&startDate, &deadline, &deadlineTime, &startedAt, &submittedAt, &completedAt, &approvedAt,
		&t.IsLate, &submissionNotes, &submissionLinks, &score, &adminFeedback, &t.CreatedAt, &t.UpdatedAt,
		&internName, &assignedByName,
	); err != nil {
		return t, err
	}

	t.TaskAssignmentID = ptrInt64FromNull(taskAssignmentID)
	t.InternID = ptrInt64FromNull(internID)
	t.Description = ptrStringFromNull(description)
	t.StartDate = ptrTimeFromNull(startDate)
	t.Deadline = ptrTimeFromNull(deadline)
	t.DeadlineTime = ptrStringFromNull(deadlineTime)
	t.StartedAt = ptrTimeFromNull(startedAt)
	t.SubmittedAt = ptrTimeFromNull(submittedAt)
	t.CompletedAt = ptrTimeFromNull(completedAt)
	t.ApprovedAt = ptrTimeFromNull(approvedAt)
	t.SubmissionNotes = ptrStringFromNull(submissionNotes)
	t.Score = ptrIntFromNull(score)
	t.AdminFeedback = ptrStringFromNull(adminFeedback)

	if submissionLinks.Valid {
		var links []models.SubmissionLink
		_ = json.Unmarshal([]byte(submissionLinks.String), &links)
		t.SubmissionLinks = links
	}
	if internName.Valid {
		t.InternName = internName.String
	}
	if assignedByName.Valid {
		t.AssignedByName = assignedByName.String
	}

	return t, nil
}

func (h *TaskHandler) getInternIDForUser(userID int64) (int64, error) {
	var internID int64
	err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", userID).Scan(&internID)
	return internID, err
}

func (h *TaskHandler) createNotification(userID int64, ntype, title, message, link string, data map[string]interface{}) error {
	var dataStr sql.NullString
	if data != nil {
		if b, err := json.Marshal(data); err == nil {
			dataStr = sql.NullString{String: string(b), Valid: true}
		}
	}
	_, err := h.db.Exec(
		`INSERT INTO notifications (user_id, type, title, message, link, data)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, ntype, title, message, nullIfEmpty(link), dataStr,
	)
	return err
}

func (h *TaskHandler) assignmentStats(assignmentID int64) map[string]interface{} {
	row := h.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'completed' AND is_late = 0 THEN 1 ELSE 0 END) as completed_on_time,
			SUM(CASE WHEN status = 'completed' AND is_late = 1 THEN 1 ELSE 0 END) as completed_late,
			SUM(CASE WHEN status = 'submitted' THEN 1 ELSE 0 END) as submitted,
			SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
			SUM(CASE WHEN status = 'revision' THEN 1 ELSE 0 END) as revision,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN status = 'scheduled' THEN 1 ELSE 0 END) as scheduled,
			AVG(CASE WHEN status = 'completed' THEN score ELSE NULL END) as average_score
		FROM tasks
		WHERE task_assignment_id = ?
	`, assignmentID)

	var total, completed, completedOnTime, completedLate, submitted, inProgress, revision, pending, scheduled sql.NullInt64
	var avgScore sql.NullFloat64
	_ = row.Scan(&total, &completed, &completedOnTime, &completedLate, &submitted, &inProgress, &revision, &pending, &scheduled, &avgScore)

	progress := 0.0
	if total.Valid && total.Int64 > 0 && completed.Valid {
		progress = (float64(completed.Int64) / float64(total.Int64)) * 100
	}

	stats := map[string]interface{}{
		"total":               int64OrZero(total),
		"completed":           int64OrZero(completed),
		"completed_on_time":   int64OrZero(completedOnTime),
		"completed_late":      int64OrZero(completedLate),
		"submitted":           int64OrZero(submitted),
		"in_progress":         int64OrZero(inProgress),
		"revision":            int64OrZero(revision),
		"pending":             int64OrZero(pending),
		"scheduled":           int64OrZero(scheduled),
		"progress_percentage": int(progress + 0.5),
		"average_score":       floatOrZero(avgScore),
	}
	return stats
}

func (h *TaskHandler) isLate(deadline sql.NullTime, deadlineTime sql.NullString, now time.Time) bool {
	if !deadline.Valid {
		return false
	}
	dt := time.Date(deadline.Time.Year(), deadline.Time.Month(), deadline.Time.Day(), 23, 59, 59, 0, now.Location())
	if deadlineTime.Valid {
		if t, err := time.Parse("15:04:05", deadlineTime.String); err == nil {
			dt = time.Date(deadline.Time.Year(), deadline.Time.Month(), deadline.Time.Day(), t.Hour(), t.Minute(), t.Second(), 0, now.Location())
		}
	}
	return now.After(dt)
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	sb := strings.Builder{}
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("?")
	}
	return sb.String()
}

func ptrStringFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func ptrTimeFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}

func ptrInt64FromNull(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func ptrIntFromNull(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int64)
	return &v
}

func nullIfEmpty(value string) sql.NullString {
	if strings.TrimSpace(value) == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: value, Valid: true}
}

func int64OrZero(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	}
	return 0
}

func floatOrZero(v sql.NullFloat64) float64 {
	if v.Valid {
		return v.Float64
	}
	return 0
}
