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

type LeaveHandler struct {
	db *sql.DB
}

func NewLeaveHandler(db *sql.DB) *LeaveHandler {
	return &LeaveHandler{db: db}
}

// Create handles the submission of Sick/Permission letters
// It expects multipart/form-data
func (h *LeaveHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// 1. Parse Multipart Form (10MB max memory)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondBadRequest(w, "Failed to parse form data")
		return
	}

	// 2. Extract Fields
	leaveType := r.FormValue("leave_type")
	reason := r.FormValue("reason")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")

	// 3. Validate Inputs
	if leaveType != "sick" && leaveType != "permission" && leaveType != "other" {
		utils.RespondBadRequest(w, "Invalid leave type. Must be: sick, permission, other")
		return
	}
	if reason == "" {
		utils.RespondBadRequest(w, "Reason is required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid start_date format (YYYY-MM-DD)")
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid end_date format (YYYY-MM-DD)")
		return
	}
	if endDate.Before(startDate) {
		utils.RespondBadRequest(w, "End date cannot be before start date")
		return
	}

	// 4. Handle File Upload (Proof)
	var attachmentPath sql.NullString
	file, header, err := r.FormFile("attachment")

	// If Sick, file is mandatory (Logic ported from Laravel)
	if leaveType == "sick" && err != nil {
		utils.RespondBadRequest(w, "Attachment (Doctor's Note) is required for sick leave")
		return
	}

	if err == nil {
		defer file.Close()
		// Use existing utils.UploadFile
		path, err := utils.UploadFile(file, header, "leaves")
		if err != nil {
			utils.RespondInternalError(w, "Failed to upload file: "+err.Error())
			return
		}
		attachmentPath = sql.NullString{String: path, Valid: true}
	}

	// 5. Get Intern ID
	var internID int64
	err = h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&internID)
	if err != nil {
		utils.RespondNotFound(w, "Intern profile not found")
		return
	}

	// 6. Insert into DB
	query := `
		INSERT INTO leave_requests (intern_id, leave_type, start_date, end_date, reason, attachment_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'pending', NOW(), NOW())
	`
	res, err := h.db.Exec(query, internID, leaveType, startDate, endDate, reason, attachmentPath)
	if err != nil {
		utils.RespondInternalError(w, "Database error: "+err.Error())
		return
	}

	id, _ := res.LastInsertId()
	utils.RespondCreated(w, "Leave request submitted successfully", map[string]int64{"id": id})
}

func (h *LeaveHandler) GetByInternID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	// Security: Interns can only see their own leaves
	if claims.Role == "intern" {
		var myInternID int64
		h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myInternID)
		if internID != myInternID {
			utils.RespondForbidden(w, "You can only view your own requests")
			return
		}
	}

	rows, err := h.db.Query(`
		SELECT id, leave_type, start_date, end_date, reason, attachment_path, status, created_at 
		FROM leave_requests 
		WHERE intern_id = ? 
		ORDER BY created_at DESC`, internID)
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}
	defer rows.Close()

	var leaves []models.LeaveRequest
	for rows.Next() {
		var l models.LeaveRequest
		var att sql.NullString
		if err := rows.Scan(&l.ID, &l.LeaveType, &l.StartDate, &l.EndDate, &l.Reason, &att, &l.Status, &l.CreatedAt); err != nil {
			continue
		}
		if att.Valid {
			path := att.String
			l.AttachmentPath = &path
		}
		l.InternID = internID
		leaves = append(leaves, l)
	}

	utils.RespondSuccess(w, "Leave requests retrieved", leaves)
}

// Approve handles Supervisor approval
func (h *LeaveHandler) Approve(w http.ResponseWriter, r *http.Request) {
	h.updateStatus(w, r, "approved")
}

// Reject handles Supervisor rejection
func (h *LeaveHandler) Reject(w http.ResponseWriter, r *http.Request) {
	h.updateStatus(w, r, "rejected")
}

func (h *LeaveHandler) updateStatus(w http.ResponseWriter, r *http.Request, status string) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// Only Supervisors/Admins can approve
	if claims.Role == "intern" {
		utils.RespondForbidden(w, "Interns cannot approve requests")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	query := `
		UPDATE leave_requests 
		SET status = ?, approved_by = ?, approved_at = NOW(), updated_at = NOW() 
		WHERE id = ?`

	_, err := h.db.Exec(query, status, claims.UserID, id) // Note: approved_by stores UserID of supervisor
	if err != nil {
		utils.RespondInternalError(w, "Failed to update status")
		return
	}

	utils.RespondSuccess(w, "Leave request "+status, nil)
}

// GetAll lists leave requests (admin/pembimbing)
func (h *LeaveHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can view all leaves")
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

	status := r.URL.Query().Get("status")
	internFilter := r.URL.Query().Get("intern_id")

	where := []string{}
	args := []interface{}{}
	if status != "" {
		where = append(where, "l.status = ?")
		args = append(args, status)
	}
	if internFilter != "" {
		if id, err := strconv.ParseInt(internFilter, 10, 64); err == nil {
			where = append(where, "l.intern_id = ?")
			args = append(args, id)
		}
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	baseFrom := `
		FROM leave_requests l
		LEFT JOIN interns i ON l.intern_id = i.id
		LEFT JOIN users iu ON i.user_id = iu.id
		LEFT JOIN users au ON l.approved_by = au.id
	`

	var total int64
	if err := h.db.QueryRow("SELECT COUNT(*) "+baseFrom+" "+whereClause, args...).Scan(&total); err != nil {
		utils.RespondInternalError(w, "Failed to count leave requests")
		return
	}

	query := `
		SELECT l.id, l.intern_id, l.leave_type, l.start_date, l.end_date, l.reason, l.attachment_path, l.status,
		       l.approved_by, l.approved_at, l.created_at, l.updated_at,
		       iu.name, au.name
	` + baseFrom + " " + whereClause + " ORDER BY l.created_at DESC LIMIT ? OFFSET ?"

	args = append(args, limit, offset)
	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch leave requests")
		return
	}
	defer rows.Close()

	var leaves []models.LeaveRequest
	for rows.Next() {
		var l models.LeaveRequest
		var attachment sql.NullString
		var approverID sql.NullInt64
		var approverAt sql.NullTime
		var internName, approverName sql.NullString

		if err := rows.Scan(
			&l.ID, &l.InternID, &l.LeaveType, &l.StartDate, &l.EndDate, &l.Reason, &attachment, &l.Status,
			&approverID, &approverAt, &l.CreatedAt, &l.UpdatedAt,
			&internName, &approverName,
		); err == nil {
			if attachment.Valid {
				path := attachment.String
				l.AttachmentPath = &path
			}
			if approverID.Valid {
				id := approverID.Int64
				l.ApprovedBy = &id
			}
			if approverAt.Valid {
				t := approverAt.Time
				l.ApprovedAt = &t
			}
			if internName.Valid {
				l.InternName = internName.String
			}
			if approverName.Valid {
				l.ApproverName = approverName.String
			}
			leaves = append(leaves, l)
		}
	}

	utils.RespondPaginated(w, leaves, utils.CalculatePagination(page, limit, total))
}

func (h *LeaveHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	query := `
		SELECT l.id, l.intern_id, l.leave_type, l.start_date, l.end_date, l.reason, l.attachment_path, l.status,
		       l.approved_by, l.approved_at, l.created_at, l.updated_at,
		       iu.name, au.name
		FROM leave_requests l
		LEFT JOIN interns i ON l.intern_id = i.id
		LEFT JOIN users iu ON i.user_id = iu.id
		LEFT JOIN users au ON l.approved_by = au.id
		WHERE l.id = ?
	`

	var l models.LeaveRequest
	var attachment sql.NullString
	var approverID sql.NullInt64
	var approverAt sql.NullTime
	var internName, approverName sql.NullString

	err := h.db.QueryRow(query, id).Scan(
		&l.ID, &l.InternID, &l.LeaveType, &l.StartDate, &l.EndDate, &l.Reason, &attachment, &l.Status,
		&approverID, &approverAt, &l.CreatedAt, &l.UpdatedAt,
		&internName, &approverName,
	)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Leave request not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	// Access check for interns
	if normalizeRole(claims.Role) == "intern" {
		var myInternID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myInternID); err != nil || myInternID != l.InternID {
			utils.RespondForbidden(w, "You do not have access to this request")
			return
		}
	}

	if attachment.Valid {
		path := attachment.String
		l.AttachmentPath = &path
	}
	if approverID.Valid {
		id := approverID.Int64
		l.ApprovedBy = &id
	}
	if approverAt.Valid {
		t := approverAt.Time
		l.ApprovedAt = &t
	}
	if internName.Valid {
		l.InternName = internName.String
	}
	if approverName.Valid {
		l.ApproverName = approverName.String
	}

	utils.RespondSuccess(w, "Leave request retrieved", l)
}

func (h *LeaveHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	// Load existing request
	var existing struct {
		InternID int64
		Status   string
	}
	err := h.db.QueryRow("SELECT intern_id, status FROM leave_requests WHERE id = ?", id).Scan(&existing.InternID, &existing.Status)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Leave request not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	if normalizeRole(claims.Role) == "intern" {
		var myInternID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myInternID); err != nil || myInternID != existing.InternID {
			utils.RespondForbidden(w, "You do not have access to this request")
			return
		}
		if existing.Status != "pending" {
			utils.RespondForbidden(w, "Only pending requests can be updated")
			return
		}
	}

	var payload struct {
		LeaveType string `json:"leave_type"`
		Reason    string `json:"reason"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	updates := []string{}
	args := []interface{}{}
	if payload.LeaveType != "" {
		updates = append(updates, "leave_type = ?")
		args = append(args, payload.LeaveType)
	}
	if payload.Reason != "" {
		updates = append(updates, "reason = ?")
		args = append(args, payload.Reason)
	}
	if payload.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", payload.StartDate); err == nil {
			updates = append(updates, "start_date = ?")
			args = append(args, parsed)
		}
	}
	if payload.EndDate != "" {
		if parsed, err := time.Parse("2006-01-02", payload.EndDate); err == nil {
			updates = append(updates, "end_date = ?")
			args = append(args, parsed)
		}
	}

	if len(updates) == 0 {
		utils.RespondBadRequest(w, "No updates provided")
		return
	}

	args = append(args, id)
	if _, err := h.db.Exec("UPDATE leave_requests SET "+strings.Join(updates, ", ")+" WHERE id = ?", args...); err != nil {
		utils.RespondInternalError(w, "Failed to update request")
		return
	}

	utils.RespondSuccess(w, "Leave request updated", nil)
}

func (h *LeaveHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	// Ensure access
	var internID int64
	if err := h.db.QueryRow("SELECT intern_id FROM leave_requests WHERE id = ?", id).Scan(&internID); err != nil {
		utils.RespondNotFound(w, "Leave request not found")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		var myInternID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myInternID); err != nil || myInternID != internID {
			utils.RespondForbidden(w, "You do not have access to this request")
			return
		}
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondBadRequest(w, "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("attachment")
	if err != nil {
		utils.RespondBadRequest(w, "Missing attachment")
		return
	}
	defer file.Close()

	path, err := utils.UploadFile(file, header, "leaves")
	if err != nil {
		utils.RespondInternalError(w, "Failed to upload attachment")
		return
	}

	if _, err := h.db.Exec("UPDATE leave_requests SET attachment_path = ? WHERE id = ?", path, id); err != nil {
		utils.RespondInternalError(w, "Failed to update attachment")
		return
	}

	utils.RespondSuccess(w, "Attachment uploaded", map[string]string{"path": path})
}
