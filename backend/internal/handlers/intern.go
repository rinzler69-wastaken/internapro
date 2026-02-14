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
	"golang.org/x/crypto/bcrypt"
)

type InternHandler struct {
	db *sql.DB
}

func NewInternHandler(db *sql.DB) *InternHandler {
	return &InternHandler{db: db}
}

type createInternRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password,omitempty"`
	FullName      string `json:"full_name"`
	NIS           string `json:"nis,omitempty"`
	StudentID     string `json:"student_id,omitempty"`
	School        string `json:"school"`
	Department    string `json:"department"`
	Phone         string `json:"phone,omitempty"`
	Address       string `json:"address,omitempty"`
	StartDate     string `json:"start_date"` // YYYY-MM-DD
	EndDate       string `json:"end_date"`   // YYYY-MM-DD
	Status        string `json:"status,omitempty"`
	InstitutionID *int64 `json:"institution_id,omitempty"`
	SupervisorID  *int64 `json:"supervisor_id,omitempty"`
}

type updateInternRequest struct {
	Email         *string `json:"email,omitempty"`
	Password      *string `json:"password,omitempty"`
	FullName      *string `json:"full_name,omitempty"`
	NIS           *string `json:"nis,omitempty"`
	StudentID     *string `json:"student_id,omitempty"`
	School        *string `json:"school,omitempty"`
	Department    *string `json:"department,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Address       *string `json:"address,omitempty"`
	StartDate     *string `json:"start_date,omitempty"`
	EndDate       *string `json:"end_date,omitempty"`
	Status        *string `json:"status,omitempty"`
	InstitutionID *int64  `json:"institution_id,omitempty"`
	SupervisorID  *int64  `json:"supervisor_id,omitempty"`
}

func (h *InternHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	role := normalizeRole(claims.Role)
	if role == "intern" {
		internID, err := h.getInternIDForUser(claims.UserID)
		if err != nil {
			utils.RespondNotFound(w, "Intern not found")
			return
		}
		r = r.Clone(r.Context())
		q := r.URL.Query()
		q.Set("intern_id", strconv.FormatInt(internID, 10))
		r.URL.RawQuery = q.Encode()
		h.GetByID(w, r)
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
	supervisorFilter := strings.TrimSpace(r.URL.Query().Get("supervisor_id"))

	where := []string{}
	args := []interface{}{}

	if status != "" {
		where = append(where, "i.status = ?")
		args = append(args, status)
	}
	if role == "pembimbing" {
		where = append(where, "i.supervisor_id = ?")
		args = append(args, claims.UserID)
	} else if supervisorFilter != "" {
		if id, err := strconv.ParseInt(supervisorFilter, 10, 64); err == nil {
			where = append(where, "i.supervisor_id = ?")
			args = append(args, id)
		}
	}
	if search != "" {
		where = append(where, "(i.full_name LIKE ? OR i.school LIKE ? OR i.department LIKE ? OR u.email LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
	}

	// RESTRICTION: Only Admin can see pending interns
	if role != "admin" {
		where = append(where, "i.status != 'pending'")
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	baseFrom := `
		FROM interns i
		JOIN users u ON i.user_id = u.id
		LEFT JOIN users su ON i.supervisor_id = su.id
		LEFT JOIN institutions inst ON i.institution_id = inst.id
	`

	var total int64
	countQuery := "SELECT COUNT(*) " + baseFrom + " " + whereClause
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		utils.RespondInternalError(w, "Failed to count interns")
		return
	}

	query := `
		SELECT i.id, i.user_id, i.institution_id, i.supervisor_id, i.full_name,
		       COALESCE(i.nis,''), COALESCE(i.student_id,''), COALESCE(i.school,''), COALESCE(i.department,''),
		       i.date_of_birth, COALESCE(i.gender,''), COALESCE(i.phone,''), COALESCE(i.address,''),
		       i.start_date, i.end_date, i.status, i.certificate_number, i.certificate_issued_at, i.created_at, i.updated_at,
		       u.email, u.avatar, COALESCE(su.name,''), COALESCE(inst.name,'')
	` + baseFrom + " " + whereClause + " ORDER BY i.created_at DESC LIMIT ? OFFSET ?"

	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch interns")
		return
	}
	defer rows.Close()

	var interns []models.InternWithDetails
	for rows.Next() {
		var i models.Intern
		var email, supervisorName, institutionName string
		var avatar sql.NullString

		if err := rows.Scan(
			&i.ID, &i.UserID, &i.InstitutionID, &i.SupervisorID, &i.FullName,
			&i.NIS, &i.StudentID, &i.School, &i.Department,
			&i.DateOfBirth, &i.Gender, &i.Phone, &i.Address,
			&i.StartDate, &i.EndDate, &i.Status, &i.CertificateNumber, &i.CertificateIssuedAt, &i.CreatedAt, &i.UpdatedAt,
			&email, &avatar, &supervisorName, &institutionName,
		); err == nil {
			interns = append(interns, models.InternWithDetails{
				Intern:          i,
				SupervisorName:  supervisorName,
				InstitutionName: institutionName,
				Email:           email,
				Avatar:          avatar,
			})
		}
	}

	utils.RespondPaginated(w, presentInterns(interns), utils.CalculatePagination(page, limit, total))
}

func (h *InternHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		idStr = r.URL.Query().Get("intern_id")
	}
	internID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid intern id")
		return
	}

	// Interns can only access their own profile
	if normalizeRole(claims.Role) == "intern" {
		myID, err := h.getInternIDForUser(claims.UserID)
		if err != nil || myID != internID {
			utils.RespondForbidden(w, "You do not have access to this intern")
			return
		}
	}

	query := `
		SELECT i.id, i.user_id, i.institution_id, i.supervisor_id, i.full_name,
		       COALESCE(i.nis,''), COALESCE(i.student_id,''), COALESCE(i.school,''), COALESCE(i.department,''),
		       i.date_of_birth, COALESCE(i.gender,''), COALESCE(i.phone,''), COALESCE(i.address,''),
		       i.start_date, i.end_date, i.status, i.certificate_number, i.certificate_issued_at, i.created_at, i.updated_at,
		       u.email, u.avatar, COALESCE(su.name,''), COALESCE(inst.name,'')
		FROM interns i
		JOIN users u ON i.user_id = u.id
		LEFT JOIN users su ON i.supervisor_id = su.id
		LEFT JOIN institutions inst ON i.institution_id = inst.id
		WHERE i.id = ?
	`

	var i models.Intern
	var email, supervisorName, institutionName string
	var avatar sql.NullString
	err = h.db.QueryRow(query, internID).Scan(
		&i.ID, &i.UserID, &i.InstitutionID, &i.SupervisorID, &i.FullName,
		&i.NIS, &i.StudentID, &i.School, &i.Department,
		&i.DateOfBirth, &i.Gender, &i.Phone, &i.Address,
		&i.StartDate, &i.EndDate, &i.Status, &i.CertificateNumber, &i.CertificateIssuedAt, &i.CreatedAt, &i.UpdatedAt,
		&email, &avatar, &supervisorName, &institutionName,
	)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Intern not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	result := models.InternWithDetails{
		Intern:          i,
		SupervisorName:  supervisorName,
		InstitutionName: institutionName,
		Email:           email,
		Avatar:          avatar,
	}

	utils.RespondSuccess(w, "Intern retrieved", presentIntern(result))
}

func (h *InternHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can create interns")
		return
	}

	var req createInternRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.FullName) == "" || strings.TrimSpace(req.StartDate) == "" || strings.TrimSpace(req.EndDate) == "" {
		utils.RespondBadRequest(w, "Email, full_name, start_date, and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid start_date")
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid end_date")
		return
	}

	status := req.Status
	if status == "" {
		status = "pending" // Default to pending for approval
	}

	var passwordHash sql.NullString
	if strings.TrimSpace(req.Password) != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondInternalError(w, "Failed to hash password")
			return
		}
		passwordHash = sql.NullString{String: string(hashed), Valid: true}
	}

	// Check for existing "new_user" via token
	var existingUserID int64
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if claims, err := middleware.ParseToken(tokenString); err == nil {
			if claims.Role == "new_user" {
				existingUserID = claims.UserID
			}
		}
	}

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var userID int64

	if existingUserID > 0 {
		// Upgrade existing user
		userID = existingUserID

		updates := []string{"role = 'intern'"}
		args := []interface{}{}

		if req.FullName != "" {
			updates = append(updates, "name = ?")
			args = append(args, req.FullName)
		}

		// Only update password if provided
		if passwordHash.Valid {
			updates = append(updates, "password_hash = ?")
			args = append(args, passwordHash.String)
		}

		// We do NOT update email here to avoid hijacking if token is compromised but email is different?
		// Actually, if they own the token, they own the account.
		// If they change email in form, we should probably update it, but let's stick to the one in token/DB for safety unless we re-verify.
		// For simplicity, we trust the token's user.

		query := "UPDATE users SET " + strings.Join(updates, ", ") + " WHERE id = ?"
		args = append(args, userID)

		if _, err := tx.Exec(query, args...); err != nil {
			utils.RespondInternalError(w, "Failed to upgrade user profile")
			return
		}
	} else {
		// Insert new user
		res, err := tx.Exec(
			"INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, 'intern')",
			req.FullName, req.Email, nullIfEmptySQL(passwordHash),
		)
		if err != nil {
			utils.RespondBadRequest(w, "Email already exists or invalid user payload")
			return
		}
		userID, _ = res.LastInsertId()
	}

	// Auto-resolve Institution ID from School name if not provided
	var institutionID *int64 = req.InstitutionID
	if institutionID == nil && strings.TrimSpace(req.School) != "" {
		var id int64
		err := tx.QueryRow("SELECT id FROM institutions WHERE name = ?", req.School).Scan(&id)
		if err == nil {
			institutionID = &id
		} else if err == sql.ErrNoRows {
			res, err := tx.Exec("INSERT INTO institutions (name) VALUES (?)", req.School)
			if err == nil {
				lid, _ := res.LastInsertId()
				institutionID = &lid
			}
		}
	}

	_, err = tx.Exec(
		`INSERT INTO interns (user_id, institution_id, supervisor_id, full_name, nis, student_id, school, department,
		                      phone, address, start_date, end_date, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, nullableInt(institutionID), nullableInt(req.SupervisorID), req.FullName, req.NIS, req.StudentID,
		req.School, req.Department, req.Phone, req.Address, startDate, endDate, status,
	)
	if err != nil {
		utils.RespondBadRequest(w, "Failed to create intern: "+err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to commit transaction")
		return
	}

	// Create notification for admins about new intern registration
	if status == "pending" {
		_, _ = h.db.Exec(`
			INSERT INTO notifications (user_id, type, title, message, link, created_at)
			SELECT u.id, 'info', 'Pendaftaran Magang Baru', ?, '/interns', NOW()
			FROM users u WHERE u.role = 'admin'
		`, "Peserta magang baru "+req.FullName+" telah mendaftar dan menunggu persetujuan.")
	}

	utils.RespondCreated(w, "Intern created. Your account is pending admin approval.", map[string]interface{}{"user_id": userID})
}

func (h *InternHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can update interns")
		return
	}

	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var req updateInternRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// Get user_id for this intern
	var userID int64
	err := h.db.QueryRow("SELECT user_id FROM interns WHERE id = ?", internID).Scan(&userID)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Intern not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	userUpdates := []string{}
	userArgs := []interface{}{}
	if req.Email != nil {
		userUpdates = append(userUpdates, "email = ?")
		userArgs = append(userArgs, *req.Email)
	}
	if req.FullName != nil {
		userUpdates = append(userUpdates, "name = ?")
		userArgs = append(userArgs, *req.FullName)
	}
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondInternalError(w, "Failed to hash password")
			return
		}
		userUpdates = append(userUpdates, "password_hash = ?")
		userArgs = append(userArgs, string(hashed))
	}
	if len(userUpdates) > 0 {
		userArgs = append(userArgs, userID)
		if _, err := tx.Exec("UPDATE users SET "+strings.Join(userUpdates, ", ")+" WHERE id = ?", userArgs...); err != nil {
			utils.RespondInternalError(w, "Failed to update user")
			return
		}
	}

	internUpdates := []string{}
	internArgs := []interface{}{}
	if req.FullName != nil {
		internUpdates = append(internUpdates, "full_name = ?")
		internArgs = append(internArgs, *req.FullName)
	}
	if req.NIS != nil {
		internUpdates = append(internUpdates, "nis = ?")
		internArgs = append(internArgs, *req.NIS)
	}
	if req.StudentID != nil {
		internUpdates = append(internUpdates, "student_id = ?")
		internArgs = append(internArgs, *req.StudentID)
	}
	if req.School != nil {
		internUpdates = append(internUpdates, "school = ?")
		internArgs = append(internArgs, *req.School)
	}
	if req.Department != nil {
		internUpdates = append(internUpdates, "department = ?")
		internArgs = append(internArgs, *req.Department)
	}
	if req.Phone != nil {
		internUpdates = append(internUpdates, "phone = ?")
		internArgs = append(internArgs, *req.Phone)
	}
	if req.Address != nil {
		internUpdates = append(internUpdates, "address = ?")
		internArgs = append(internArgs, *req.Address)
	}
	if req.StartDate != nil {
		if parsed, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			internUpdates = append(internUpdates, "start_date = ?")
			internArgs = append(internArgs, parsed)
		}
	}
	if req.EndDate != nil {
		if parsed, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			internUpdates = append(internUpdates, "end_date = ?")
			internArgs = append(internArgs, parsed)
		}
	}
	if req.Status != nil {
		internUpdates = append(internUpdates, "status = ?")
		internArgs = append(internArgs, *req.Status)
	}
	if req.InstitutionID != nil {
		internUpdates = append(internUpdates, "institution_id = ?")
		internArgs = append(internArgs, nullableInt(req.InstitutionID))
	}
	if req.SupervisorID != nil {
		internUpdates = append(internUpdates, "supervisor_id = ?")
		internArgs = append(internArgs, nullableInt(req.SupervisorID))
	}

	if len(internUpdates) > 0 {
		internArgs = append(internArgs, internID)
		if _, err := tx.Exec("UPDATE interns SET "+strings.Join(internUpdates, ", ")+" WHERE id = ?", internArgs...); err != nil {
			utils.RespondInternalError(w, "Failed to update intern")
			return
		}
	}

	// If intern is approved/activated, ensure the linked user is no longer marked as new_user
	if req.Status != nil && strings.EqualFold(*req.Status, "active") {
		if _, err := tx.Exec("UPDATE users SET role = 'intern' WHERE id = ? AND role = 'new_user'", userID); err != nil {
			utils.RespondInternalError(w, "Failed to update user role")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to commit transaction")
		return
	}

	utils.RespondSuccess(w, "Intern updated", nil)
}

func (h *InternHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can delete interns")
		return
	}

	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)

	// Get related user_id so we can remove credentials as well
	var userID int64
	if err := h.db.QueryRow("SELECT user_id FROM interns WHERE id = ?", internID).Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			utils.RespondNotFound(w, "Intern not found")
			return
		}
		utils.RespondInternalError(w, "Failed to find intern")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Manual Cascade: Delete related data to ensure cleanup
	// We delete these explicitly to ensure data integrity and not rely solely on DB cascades
	var deleteQueries = []string{
		"DELETE FROM attendances WHERE intern_id = ?",
		"DELETE FROM leave_requests WHERE intern_id = ?",
		"DELETE FROM assessments WHERE intern_id = ?",
		"DELETE FROM reports WHERE intern_id = ?",
		"DELETE FROM certificates WHERE intern_id = ?",
		"DELETE FROM task_assignment_interns WHERE intern_id = ?",
		"DELETE FROM tasks WHERE intern_id = ?",
	}

	for _, query := range deleteQueries {
		if _, err := tx.Exec(query, internID); err != nil {
			utils.RespondInternalError(w, "Failed to delete related intern data")
			return
		}
	}

	// Delete user first (cascades to interns via FK ON DELETE CASCADE)
	if _, err := tx.Exec("DELETE FROM users WHERE id = ?", userID); err != nil {
		utils.RespondInternalError(w, "Failed to delete user")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to commit deletion")
		return
	}

	utils.RespondSuccess(w, "Intern and credentials deleted", nil)
}

// Helpers
func (h *InternHandler) getInternIDForUser(userID int64) (int64, error) {
	var internID int64
	err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", userID).Scan(&internID)
	return internID, err
}

func nullableInt(val *int64) sql.NullInt64 {
	if val == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *val, Valid: *val != 0}
}

func nullIfEmptySQL(val sql.NullString) interface{} {
	if val.Valid {
		return val.String
	}
	return nil
}

// --- Presentation helpers to avoid sql.Null* leaking as objects ---
func presentInterns(list []models.InternWithDetails) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(list))
	for _, it := range list {
		out = append(out, presentIntern(it))
	}
	return out
}

func presentIntern(it models.InternWithDetails) map[string]interface{} {
	return map[string]interface{}{
		"id":                    it.ID,
		"user_id":               it.UserID,
		"institution_id":        nullInt64ToPtr(it.InstitutionID),
		"supervisor_id":         nullInt64ToPtr(it.SupervisorID),
		"full_name":             it.FullName,
		"nis":                   nullStringToPtr(it.NIS),
		"student_id":            nullStringToPtr(it.StudentID),
		"school":                nullStringToPtr(it.School),
		"department":            nullStringToPtr(it.Department),
		"date_of_birth":         nullTimeToPtr(it.DateOfBirth),
		"gender":                it.Gender,
		"phone":                 nullStringToPtr(it.Phone),
		"address":               nullStringToPtr(it.Address),
		"start_date":            it.StartDate,
		"end_date":              it.EndDate,
		"status":                it.Status,
		"certificate_number":    nullStringToPtr(it.CertificateNumber),
		"certificate_issued_at": nullTimeToPtr(it.CertificateIssuedAt),
		"created_at":            it.CreatedAt,
		"updated_at":            it.UpdatedAt,
		"email":                 it.Email,
		"supervisor_name":       it.SupervisorName,
		"institution_name":      it.InstitutionName,
		"avatar":                nullStringToPtr(it.Avatar),
	}
}

func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}
func nullInt64ToPtr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}
func nullTimeToPtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}
