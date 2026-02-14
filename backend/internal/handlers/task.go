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
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	SubmissionMethod   string  `json:"submission_method"` // links, files, both
	Priority           string  `json:"priority"`          // low, medium, high
	StartDate          string  `json:"start_date"`        // YYYY-MM-DD
	Deadline           string  `json:"deadline"`          // YYYY-MM-DD
	DeadlineTime       string  `json:"deadline_time"`     // HH:MM
	AssignTo           string  `json:"assign_to"`         // all, selected
	InternIDs          []int64 `json:"intern_ids"`
	AssignerID         int64   `json:"assigner_id"`
	CustomAssignerName string  `json:"custom_assigner_name"`
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
		// Supervisors/pembimbing only see tasks for their own interns OR tasks they assigned
		if claims.Role == "pembimbing" || claims.Role == "supervisor" {
			where = append(where, "(i.supervisor_id = ? OR t.assigned_by = ? OR t.assigner_id = ?)")
			args = append(args, claims.UserID, claims.UserID, claims.UserID)
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
		LEFT JOIN users su ON t.assigner_id = su.id
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
		SELECT t.id, t.task_assignment_id, t.intern_id, t.assigned_by, t.title, t.description, t.submission_method, t.priority, t.status,
		       t.start_date, t.deadline, t.deadline_time, t.started_at, t.submitted_at, t.completed_at, t.approved_at,
		       t.is_late, t.submission_notes, t.submission_links, t.submission_file, t.score, t.admin_feedback, t.created_at, t.updated_at,
		       t.is_unscheduled, COALESCE(t.assigner_id, t.assigned_by) as assigner_id, t.custom_assigner_name, iu.name, au.name, su.name, su.role
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
		SELECT t.id, t.task_assignment_id, t.intern_id, t.assigned_by, t.title, t.description, t.submission_method, t.priority, t.status,
		       t.start_date, t.deadline, t.deadline_time, t.started_at, t.submitted_at, t.completed_at, t.approved_at,
		       t.is_late, t.submission_notes, t.submission_links, t.submission_file, t.score, t.admin_feedback, t.created_at, t.updated_at,
		       t.is_unscheduled, COALESCE(t.assigner_id, t.assigned_by) as assigner_id, t.custom_assigner_name, iu.name, au.name, su.name, su.role
		FROM tasks t
		LEFT JOIN interns i ON t.intern_id = i.id
		LEFT JOIN users iu ON i.user_id = iu.id
		LEFT JOIN users au ON t.assigned_by = au.id
		LEFT JOIN users su ON COALESCE(t.assigner_id, t.assigned_by) = su.id
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

	role := normalizeRole(claims.Role)
	isIntern := role == "intern"

	if isIntern {
		var allowInternLogging string
		err := h.db.QueryRow("SELECT value FROM settings WHERE `key` = 'ALLOW_INTERN_UNSCHEDULED_LOGGING'").Scan(&allowInternLogging)
		if err != nil || allowInternLogging != "true" {
			utils.RespondForbidden(w, "Pencatatan tugas oleh intern sedang dinonaktifkan")
			return
		}
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
	if req.SubmissionMethod == "" {
		req.SubmissionMethod = "both"
	}
	if req.SubmissionMethod != "links" && req.SubmissionMethod != "files" && req.SubmissionMethod != "both" {
		utils.RespondBadRequest(w, "Invalid submission_method")
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

	if isIntern {
		// Intern can only assign to themselves
		id, err := h.getInternIDForUser(claims.UserID)
		if err != nil {
			utils.RespondNotFound(w, "Intern profile not found")
			return
		}
		interns = append(interns, internRow{ID: id, UserID: claims.UserID})
	} else if req.AssignTo == "all" {
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

		// Calculate assignerID for the task record
		var dbAssignerID sql.NullInt64
		if isIntern {
			if req.AssignerID > 0 {
				dbAssignerID = sql.NullInt64{Int64: req.AssignerID, Valid: true}
			}
		} else {
			dbAssignerID = sql.NullInt64{Int64: claims.UserID, Valid: true}
		}

		taskRes, err := tx.Exec(
			`INSERT INTO tasks (task_assignment_id, title, description, submission_method, intern_id, assigned_by, priority, status, start_date, deadline, deadline_time, is_unscheduled, assigner_id, custom_assigner_name)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			assignmentID, req.Title, nullIfEmpty(req.Description), req.SubmissionMethod, it.ID, claims.UserID, req.Priority, initialStatus, startDate, deadline, deadlineTime,
			isIntern, dbAssignerID, nullIfEmpty(req.CustomAssignerName),
		)
		if err != nil {
			utils.RespondInternalError(w, "Failed to create tasks")
			return
		}

		taskID, _ := taskRes.LastInsertId()

		// For intern reporting, send notification to the assigner
		if isIntern && req.AssignerID > 0 {
			_ = createNotification(h.db, req.AssignerID, models.NotificationTaskAssigned, "Laporan Tugas Baru: "+req.Title,
				"Intern telah melaporkan tugas baru untuk di-review.", "/tasks/"+strconv.FormatInt(taskID, 10), map[string]interface{}{"task_id": taskID})
		} else if !isScheduled {
			_ = createNotification(h.db, it.UserID, models.NotificationTaskAssigned, "Tugas Baru: "+req.Title,
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
	vars := mux.Vars(r)
	taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

	// Load task for permission checks
	// We need assigner_id AND assigner's role
	var current struct {
		Status        string
		StartedAt     sql.NullTime
		Deadline      sql.NullTime
		DeadlineTime  sql.NullString
		IsUnscheduled bool
		InternID      int64
		AssignerID    sql.NullInt64
		AssignerRole  sql.NullString
	}
	// Join users table to get assigner's role
	query := `
		SELECT t.status, t.started_at, t.deadline, t.deadline_time, t.is_unscheduled, t.intern_id, COALESCE(t.assigner_id, t.assigned_by) as assigner_id, u.role
		FROM tasks t
		LEFT JOIN users u ON COALESCE(t.assigner_id, t.assigned_by) = u.id
		WHERE t.id = ?
	`
	err := h.db.QueryRow(query, taskID).
		Scan(&current.Status, &current.StartedAt, &current.Deadline, &current.DeadlineTime, &current.IsUnscheduled, &current.InternID, &current.AssignerID, &current.AssignerRole)

	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Task not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	role := normalizeRole(claims.Role)
	assignerID := int64OrZero(current.AssignerID)
	assignerRole := ""
	if current.AssignerRole.Valid {
		assignerRole = normalizeRole(current.AssignerRole.String)
	}

	// Permission Checks
	if role == "intern" {
		// Interns: Only self-reported tasks they own
		ownInternID, _ := h.getInternIDForUser(claims.UserID)
		if !current.IsUnscheduled || current.InternID != ownInternID {
			utils.RespondForbidden(w, "You can only update your own self-reported tasks")
			return
		}
	} else if role == "supervisor" || role == "pembimbing" {
		// Supervisors: Only tasks they assigned
		// (IsUnscheduled tasks assigned to them count as assigned by them in this logic?
		// No, IsUnscheduled tasks have them as AssignerID if they are the target)
		if assignerID != claims.UserID {
			utils.RespondForbidden(w, "You can only update tasks you assigned")
			return
		}
	} else if role == "admin" || role == "super_admin" {
		// Admins: Own tasks OR Supervisor tasks
		// Allowed if: Assigner is Self OR Assigner is Supervisor
		// Denied if: Assigner is another Admin (unless it's self) ??
		// Rule: "admin can edit... its own... but not supervisors [created by superadmin?]" -> "supervisors can... not admins"
		// Rule: "admins can still edit and delete supervisor-created task entries"

		isOwnTask := assignerID == claims.UserID
		isSupervisorTask := assignerRole == "supervisor" || assignerRole == "pembimbing"

		if !isOwnTask && !isSupervisorTask {
			// This means the task was created by another Admin (or potentially System/Unknown, but realistically another Admin)
			// And we are an Admin.
			// "admin can edit... its own... assigned to any intern" -> Implies they CANNOT edit other admins' tasks?
			// The prompt says "supervisors... not admins". And "admins can... supervisor-created".
			// It doesn't explicitly ban Admin vs Admin, but "its own" suggests restrictiveness.
			// However, usually Admins have broad power.
			// Let's stick to: Own OR Supervisor-created.
			// If AssignerRole is 'admin' and AssignerID != Claims.UserID -> Forbidden.

			utils.RespondForbidden(w, "You can only edit your own tasks or supervisor tasks")
			return
		}
	}

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// Interns cannot update status to 'completed' via this endpoint (usually handled by Submit/MarkComplete)
	// But let's allow common updates for now based on previous logic

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
	query = "UPDATE tasks SET " + strings.Join(updates, ", ") + " WHERE id = ?"
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

	vars := mux.Vars(r)
	taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

	// Load task for permission check
	var current struct {
		IsUnscheduled bool
		InternID      int64
		AssignerID    sql.NullInt64
		AssignerRole  sql.NullString
	}
	// Join users table to get assigner's role
	query := `
		SELECT t.is_unscheduled, t.intern_id, COALESCE(t.assigner_id, t.assigned_by) as assigner_id, u.role
		FROM tasks t
		LEFT JOIN users u ON COALESCE(t.assigner_id, t.assigned_by) = u.id
		WHERE t.id = ?
	`
	err := h.db.QueryRow(query, taskID).
		Scan(&current.IsUnscheduled, &current.InternID, &current.AssignerID, &current.AssignerRole)

	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Task not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	role := normalizeRole(claims.Role)
	assignerID := int64OrZero(current.AssignerID)
	assignerRole := ""
	if current.AssignerRole.Valid {
		assignerRole = normalizeRole(current.AssignerRole.String)
	}

	if role == "intern" {
		ownInternID, _ := h.getInternIDForUser(claims.UserID)
		if !current.IsUnscheduled || current.InternID != ownInternID {
			utils.RespondForbidden(w, "You can only delete your own self-reported tasks")
			return
		}
	} else if role == "supervisor" || role == "pembimbing" {
		if assignerID != claims.UserID {
			utils.RespondForbidden(w, "You can only delete tasks you assigned")
			return
		}
	} else if role == "admin" || role == "super_admin" {
		isOwnTask := assignerID == claims.UserID
		isSupervisorTask := assignerRole == "supervisor" || assignerRole == "pembimbing"

		if !isOwnTask && !isSupervisorTask {
			utils.RespondForbidden(w, "You can only delete your own tasks or supervisor tasks")
			return
		}
	}

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
	var submissionFilePath sql.NullString

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.RespondBadRequest(w, "Failed to parse form data")
			return
		}
		req.SubmissionNotes = r.FormValue("submission_notes")
		linksStr := r.FormValue("links")
		if linksStr != "" {
			if err := json.Unmarshal([]byte(linksStr), &req.Links); err != nil {
				utils.RespondBadRequest(w, "Invalid links format")
				return
			}
		}

		file, header, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			path, err := utils.UploadFile(file, header, "tasks")
			if err != nil {
				utils.RespondInternalError(w, "Upload failed: "+err.Error())
				return
			}
			submissionFilePath = sql.NullString{String: path, Valid: true}
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondBadRequest(w, "Invalid request body")
			return
		}
	}

	// Fetch task submission_method to validate requirements
	var submissionMethod sql.NullString
	var deadline sql.NullTime
	var deadlineTime sql.NullString
	err = h.db.QueryRow("SELECT submission_method, deadline, deadline_time FROM tasks WHERE id = ? AND intern_id = ?", taskID, internID).
		Scan(&submissionMethod, &deadline, &deadlineTime)
	if err == sql.ErrNoRows {
		utils.RespondForbidden(w, "You do not have access to this task")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	method := "both"
	if submissionMethod.Valid {
		method = submissionMethod.String
	}

	// Validate based on submission method
	hasLinks := len(req.Links) > 0
	hasFile := submissionFilePath.Valid

	switch method {
	case "links":
		if !hasLinks {
			utils.RespondBadRequest(w, "Link pengumpulan wajib diisi")
			return
		}
	case "files":
		if !hasFile {
			utils.RespondBadRequest(w, "File pengumpulan wajib diunggah")
			return
		}
	case "both":
		if !hasLinks && !hasFile {
			utils.RespondBadRequest(w, "Minimal satu link atau file harus diisi")
			return
		}
	}

	for _, l := range req.Links {
		if strings.TrimSpace(l.Label) == "" || strings.TrimSpace(l.URL) == "" {
			utils.RespondBadRequest(w, "Each link must have label and url")
			return
		}
	}

	linksJSON, _ := json.Marshal(req.Links)
	now := time.Now()
	isLate := h.isLate(deadline, deadlineTime, now)

	_, err = h.db.Exec(
		`UPDATE tasks SET status = 'submitted', submitted_at = ?, submission_notes = ?, submission_links = ?, submission_file = ?, is_late = ?,
		        started_at = COALESCE(started_at, ?)
		 WHERE id = ?`,
		now, nullIfEmpty(req.SubmissionNotes), string(linksJSON), submissionFilePath, isLate, now, taskID,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to submit task")
		return
	}

	// Notify Supervisor
	var supervisorID int64
	err = h.db.QueryRow("SELECT assigned_by FROM tasks WHERE id = ?", taskID).Scan(&supervisorID)
	if err == nil {
		_ = createNotification(h.db, supervisorID, models.NotificationTaskSubmitted, "Tugas Dikumpulkan",
			"Seorang intern telah mengumpulkan tugas. Silakan periksa.", "/tasks/"+strconv.FormatInt(taskID, 10),
			map[string]interface{}{"task_id": taskID})
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
	var assignerID sql.NullInt64
	err := h.db.QueryRow(
		`SELECT i.user_id, t.assigner_id FROM tasks t
		 LEFT JOIN interns i ON t.intern_id = i.id
		 WHERE t.id = ?`, taskID,
	).Scan(&internUserID, &assignerID)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Task not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	rawRole := strings.ToLower(claims.Role)
	isAdmin := rawRole == "admin" || rawRole == "super_admin"
	isAssigner := assignerID.Valid && assignerID.Int64 == claims.UserID

	if !isAdmin && !isAssigner {
		utils.RespondForbidden(w, "Only the assigned admin or a super-administrator can review this task")
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
		_ = createNotification(h.db, internUserID, models.NotificationTaskApproved, "Tugas Disetujui",
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
		_ = createNotification(h.db, internUserID, models.NotificationTaskRevision, "Perlu Revisi",
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
		`SELECT t.id, t.task_assignment_id, t.intern_id, t.assigned_by, t.title, t.description, t.submission_method, t.priority, t.status,
		        t.start_date, t.deadline, t.deadline_time, t.started_at, t.submitted_at, t.completed_at, t.approved_at,
		        t.is_late, t.submission_notes, t.submission_links, t.submission_file, t.score, t.admin_feedback, t.created_at, t.updated_at,
		        iu.name, au.name, su.name
		 FROM tasks t
		 LEFT JOIN interns i ON t.intern_id = i.id
		 LEFT JOIN users iu ON i.user_id = iu.id
		 LEFT JOIN users au ON t.assigned_by = au.id
		 LEFT JOIN users su ON t.assigner_id = su.id
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
	var submissionMethod sql.NullString
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
	var submissionFile sql.NullString
	var internName sql.NullString
	var assignedByName sql.NullString
	var assignerID sql.NullInt64
	var customAssignerName sql.NullString

	var assignerName sql.NullString
	var assignerRole sql.NullString

	if err := scanner.Scan(
		&t.ID, &taskAssignmentID, &internID, &t.AssignedBy, &t.Title, &description, &submissionMethod, &t.Priority, &t.Status,
		&startDate, &deadline, &deadlineTime, &startedAt, &submittedAt, &completedAt, &approvedAt,
		&t.IsLate, &submissionNotes, &submissionLinks, &submissionFile, &score, &adminFeedback, &t.CreatedAt, &t.UpdatedAt,
		&t.IsUnscheduled, &assignerID, &customAssignerName, &internName, &assignedByName, &assignerName, &assignerRole,
	); err != nil {
		return t, err
	}

	t.TaskAssignmentID = ptrInt64FromNull(taskAssignmentID)
	t.InternID = ptrInt64FromNull(internID)
	t.Description = ptrStringFromNull(description)
	t.SubmissionMethod = ptrStringFromNull(submissionMethod)
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
	t.SubmissionFile = ptrStringFromNull(submissionFile)
	t.AssignerID = ptrInt64FromNull(assignerID)
	t.CustomAssignerName = ptrStringFromNull(customAssignerName)

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
	if assignerName.Valid {
		t.AssignerName = assignerName.String
	}
	if assignerRole.Valid {
		t.AssignerRole = assignerRole.String
	}

	return t, nil
}

func (h *TaskHandler) getInternIDForUser(userID int64) (int64, error) {
	var internID int64
	err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", userID).Scan(&internID)
	return internID, err
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
