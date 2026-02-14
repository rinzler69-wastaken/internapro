package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type SupervisorHandler struct {
	db *sql.DB
}

func NewSupervisorHandler(db *sql.DB) *SupervisorHandler {
	return &SupervisorHandler{db: db}
}

type createSupervisorRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	NIP         string `json:"nip"`
	Phone       string `json:"phone"`
	Position    string `json:"position"`
	Address     string `json:"address"`
	Institution string `json:"institution"`
	Status      string `json:"status"`
}

type updateSupervisorRequest struct {
	Name        *string `json:"name"`
	Email       *string `json:"email"`
	Password    *string `json:"password"`
	NIP         *string `json:"nip"`
	Phone       *string `json:"phone"`
	Position    *string `json:"position"`
	Address     *string `json:"address"`
	Institution *string `json:"institution"`
	Status      *string `json:"status"`
}

func (h *SupervisorHandler) GetAll(w http.ResponseWriter, r *http.Request) {
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

	where := []string{}
	args := []interface{}{}

	if status != "" {
		where = append(where, "s.status = ?")
		args = append(args, status)
	}
	if search != "" {
		where = append(where, "(u.name LIKE ? OR u.email LIKE ? OR s.institution LIKE ? OR s.nip LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
	}

	// RESTRICTION: Only Admin can see pending supervisors
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok || normalizeRole(claims.Role) != "admin" {
		where = append(where, "s.status != 'pending'")
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM supervisors s JOIN users u ON s.user_id = u.id " + whereClause
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		utils.RespondInternalError(w, "Failed to count supervisors")
		return
	}

	query := `
		SELECT s.id, s.user_id, s.full_name, s.nip, s.phone, s.position, s.address, s.institution,
		       s.status, s.created_at, s.updated_at, u.email, u.name, u.avatar,
		       (SELECT COUNT(*) FROM interns i WHERE i.supervisor_id = s.user_id) as interns_count
		FROM supervisors s
		JOIN users u ON s.user_id = u.id
	` + whereClause + ` ORDER BY s.created_at DESC LIMIT ? OFFSET ?`

	args = append(args, limit, offset)
	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch supervisors")
		return
	}
	defer rows.Close()

	supervisors := []models.Supervisor{}
	for rows.Next() {
		var s models.Supervisor
		var nip, phone, position, address, institution sql.NullString
		var email, userName string
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.FullName, &nip, &phone, &position, &address, &institution,
			&s.Status, &s.CreatedAt, &s.UpdatedAt, &email, &userName, &s.Avatar, &s.InternsCount,
		); err != nil {
			continue
		}
		if strings.TrimSpace(s.FullName) == "" {
			s.FullName = userName
		}
		s.Email = email
		s.NIP = ptrStringFromNull(nip)
		s.Phone = ptrStringFromNull(phone)
		s.Position = ptrStringFromNull(position)
		s.Address = ptrStringFromNull(address)
		s.Institution = ptrStringFromNull(institution)
		supervisors = append(supervisors, s)
	}

	utils.RespondPaginated(w, supervisors, utils.CalculatePagination(page, limit, total))
}

func (h *SupervisorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	var s models.Supervisor
	var nip, phone, position, address, institution sql.NullString
	var email, userName string
	query := `
		SELECT s.id, s.user_id, s.full_name, s.nip, s.phone, s.position, s.address, s.institution,
		       s.status, s.created_at, s.updated_at, u.email, u.name, u.avatar,
		       (SELECT COUNT(*) FROM interns i WHERE i.supervisor_id = s.user_id) as interns_count
		FROM supervisors s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = ?
	`
	if err := h.db.QueryRow(query, id).Scan(
		&s.ID, &s.UserID, &s.FullName, &nip, &phone, &position, &address, &institution,
		&s.Status, &s.CreatedAt, &s.UpdatedAt, &email, &userName, &s.Avatar, &s.InternsCount,
	); err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Supervisor not found")
		return
	} else if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}
	if strings.TrimSpace(s.FullName) == "" {
		s.FullName = userName
	}
	s.Email = email
	s.NIP = ptrStringFromNull(nip)
	s.Phone = ptrStringFromNull(phone)
	s.Position = ptrStringFromNull(position)
	s.Address = ptrStringFromNull(address)
	s.Institution = ptrStringFromNull(institution)

	utils.RespondSuccess(w, "Supervisor retrieved", s)
}

func (h *SupervisorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createSupervisorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		utils.RespondBadRequest(w, "Name, email, and password are required")
		return
	}

	if len(req.Password) < 6 {
		utils.RespondBadRequest(w, "Password must be at least 6 characters")
		return
	}

	var exists int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&exists)
	if exists > 0 {
		utils.RespondBadRequest(w, "Email already exists")
		return
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "pending" {
		utils.RespondBadRequest(w, "Invalid status")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondInternalError(w, "Failed to hash password")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, ?)", req.Name, req.Email, string(hashed), "pembimbing")
	if err != nil {
		utils.RespondInternalError(w, "Failed to create user")
		return
	}
	userID, _ := res.LastInsertId()

	supRes, err := tx.Exec(
		`INSERT INTO supervisors (user_id, full_name, nip, phone, position, address, institution, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, req.Name, nullIfEmpty(req.NIP), nullIfEmpty(req.Phone), nullIfEmpty(req.Position), nullIfEmpty(req.Address), nullIfEmpty(req.Institution), status,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to create supervisor")
		return
	}
	supervisorID, _ := supRes.LastInsertId()

	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to commit transaction")
		return
	}

	utils.RespondCreated(w, "Supervisor created", map[string]interface{}{
		"user_id":       userID,
		"supervisor_id": supervisorID,
	})
}

func (h *SupervisorHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	var req updateSupervisorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	var userID int64
	var currentEmail string
	if err := h.db.QueryRow("SELECT user_id FROM supervisors WHERE id = ?", id).Scan(&userID); err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Supervisor not found")
		return
	} else if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}
	_ = h.db.QueryRow("SELECT email FROM users WHERE id = ?", userID).Scan(&currentEmail)

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	userUpdates := []string{}
	userArgs := []interface{}{}
	if req.Name != nil {
		userUpdates = append(userUpdates, "name = ?")
		userArgs = append(userArgs, *req.Name)
	}
	if req.Email != nil {
		newEmail := strings.TrimSpace(*req.Email)
		if newEmail != "" && newEmail != currentEmail {
			var exists int
			_ = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", newEmail).Scan(&exists)
			if exists > 0 {
				utils.RespondBadRequest(w, "Email already exists")
				return
			}
		}
		userUpdates = append(userUpdates, "email = ?")
		userArgs = append(userArgs, newEmail)
	}
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		if len(*req.Password) < 6 {
			utils.RespondBadRequest(w, "Password must be at least 6 characters")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondInternalError(w, "Failed to hash password")
			return
		}
		userUpdates = append(userUpdates, "password_hash = ?")
		userArgs = append(userArgs, string(hash))
	}

	if len(userUpdates) > 0 {
		userArgs = append(userArgs, userID)
		if _, err := tx.Exec("UPDATE users SET "+strings.Join(userUpdates, ", ")+" WHERE id = ?", userArgs...); err != nil {
			utils.RespondInternalError(w, "Failed to update user")
			return
		}
	}

	supUpdates := []string{}
	supArgs := []interface{}{}
	if req.Name != nil {
		supUpdates = append(supUpdates, "full_name = ?")
		supArgs = append(supArgs, *req.Name)
	}
	if req.NIP != nil {
		supUpdates = append(supUpdates, "nip = ?")
		supArgs = append(supArgs, nullIfEmpty(*req.NIP))
	}
	if req.Phone != nil {
		supUpdates = append(supUpdates, "phone = ?")
		supArgs = append(supArgs, nullIfEmpty(*req.Phone))
	}
	if req.Position != nil {
		supUpdates = append(supUpdates, "position = ?")
		supArgs = append(supArgs, nullIfEmpty(*req.Position))
	}
	if req.Address != nil {
		supUpdates = append(supUpdates, "address = ?")
		supArgs = append(supArgs, nullIfEmpty(*req.Address))
	}
	if req.Institution != nil {
		supUpdates = append(supUpdates, "institution = ?")
		supArgs = append(supArgs, nullIfEmpty(*req.Institution))
	}
	if req.Status != nil {
		status := strings.TrimSpace(*req.Status)
		if status != "" {
			if status != "active" && status != "pending" {
				utils.RespondBadRequest(w, "Invalid status")
				return
			}
			supUpdates = append(supUpdates, "status = ?")
			supArgs = append(supArgs, status)
		}
	}

	if len(supUpdates) > 0 {
		supArgs = append(supArgs, id)
		if _, err := tx.Exec("UPDATE supervisors SET "+strings.Join(supUpdates, ", ")+" WHERE id = ?", supArgs...); err != nil {
			utils.RespondInternalError(w, "Failed to update supervisor")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to commit transaction")
		return
	}

	utils.RespondSuccess(w, "Supervisor updated", nil)
}

func (h *SupervisorHandler) Approve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	res, err := h.db.Exec("UPDATE supervisors SET status = 'active' WHERE id = ?", id)
	if err != nil {
		utils.RespondInternalError(w, "Failed to approve supervisor")
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		utils.RespondNotFound(w, "Supervisor not found")
		return
	}

	utils.RespondSuccess(w, "Supervisor approved", nil)
}

func (h *SupervisorHandler) Reject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	var userID int64
	if err := h.db.QueryRow("SELECT user_id FROM supervisors WHERE id = ?", id).Scan(&userID); err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Supervisor not found")
		return
	} else if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Manual Cascade: Delete assessments (because ON DELETE RESTRICT)
	if _, err := tx.Exec("DELETE FROM assessments WHERE assessed_by = ?", userID); err != nil {
		utils.RespondInternalError(w, "Failed to delete supervisor's assessments")
		return
	}

	if _, err := tx.Exec("DELETE FROM supervisors WHERE id = ?", id); err != nil {
		utils.RespondInternalError(w, "Failed to delete supervisor")
		return
	}
	if _, err := tx.Exec("DELETE FROM users WHERE id = ?", userID); err != nil {
		utils.RespondInternalError(w, "Failed to delete user")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to commit transaction")
		return
	}

	utils.RespondSuccess(w, "Supervisor rejected", nil)
}

func (h *SupervisorHandler) GetAllPublic(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT s.user_id, u.name
		FROM supervisors s
		JOIN users u ON s.user_id = u.id
		WHERE s.status = 'active'
		ORDER BY u.name ASC
	`)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch supervisors")
		return
	}
	defer rows.Close()

	supervisors := []map[string]interface{}{}
	for rows.Next() {
		var userID int64
		var name string
		if err := rows.Scan(&userID, &name); err != nil {
			continue
		}
		supervisors = append(supervisors, map[string]interface{}{
			"user_id": userID,
			"name":    name,
		})
	}

	utils.RespondSuccess(w, "Supervisors retrieved", supervisors)
}

func (h *SupervisorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	var userID int64
	if err := h.db.QueryRow("SELECT user_id FROM supervisors WHERE id = ?", id).Scan(&userID); err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Supervisor not found")
		return
	} else if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Manual Cascade: Delete assessments (because ON DELETE RESTRICT)
	if _, err := tx.Exec("DELETE FROM assessments WHERE assessed_by = ?", userID); err != nil {
		utils.RespondInternalError(w, "Failed to delete supervisor's assessments")
		return
	}

	if _, err := tx.Exec("DELETE FROM supervisors WHERE id = ?", id); err != nil {
		utils.RespondInternalError(w, "Failed to delete supervisor")
		return
	}
	if _, err := tx.Exec("DELETE FROM users WHERE id = ?", userID); err != nil {
		utils.RespondInternalError(w, "Failed to delete user")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to commit transaction")
		return
	}

	utils.RespondSuccess(w, "Supervisor deleted", nil)
}

// Register handles public supervisor self-registration (similar to intern registration)
func (h *SupervisorHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string `json:"name"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
		NIP             string `json:"nip"`
		Phone           string `json:"phone"`
		Institution     string `json:"institution"`
		Address         string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" ||
		strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Institution) == "" {
		utils.RespondBadRequest(w, "Name, email, password, and institution are required")
		return
	}

	// Validate password
	if len(req.Password) < 6 {
		utils.RespondBadRequest(w, "Password must be at least 6 characters")
		return
	}

	if req.Password != req.ConfirmPassword {
		utils.RespondBadRequest(w, "Passwords do not match")
		return
	}

	// Check if email already exists
	var exists int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&exists)
	if exists > 0 {
		utils.RespondBadRequest(w, "Email already registered")
		return
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

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var userID int64

	if existingUserID > 0 {
		userID = existingUserID
		// Upgrade
		updates := []string{"role = 'pembimbing'"}
		args := []interface{}{}

		if req.Name != "" {
			updates = append(updates, "name = ?")
			args = append(args, req.Name)
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondInternalError(w, "Failed to hash password")
			return
		}
		updates = append(updates, "password_hash = ?")
		args = append(args, string(hashed))

		query := "UPDATE users SET " + strings.Join(updates, ", ") + " WHERE id = ?"
		args = append(args, userID)

		if _, err := tx.Exec(query, args...); err != nil {
			utils.RespondInternalError(w, "Failed to upgrade user profile")
			return
		}
	} else {
		// Hash password
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondInternalError(w, "Failed to hash password")
			return
		}

		// Create user with pembimbing role
		res, err := tx.Exec(
			"INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, ?)",
			req.Name, req.Email, string(hashed), "pembimbing",
		)
		if err != nil {
			utils.RespondInternalError(w, "Failed to create user account")
			return
		}
		userID, _ = res.LastInsertId()
	}

	// Create supervisor record with pending status
	_, err = tx.Exec(
		`INSERT INTO supervisors (user_id, full_name, nip, phone, address, institution, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, req.Name, nullIfEmpty(req.NIP), nullIfEmpty(req.Phone),
		nullIfEmpty(req.Address), req.Institution, "pending",
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to create supervisor profile")
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to complete registration")
		return
	}

	// Create notification for admins about new supervisor registration
	_, _ = h.db.Exec(`
		INSERT INTO notifications (user_id, type, title, message, link, created_at)
		SELECT u.id, 'info', 'Pendaftaran Pembimbing Baru', ?, '/supervisors', NOW()
		FROM users u WHERE u.role = 'admin'
	`, "Pembimbing baru "+req.Name+" telah mendaftar dan menunggu persetujuan.")

	utils.RespondCreated(w, "Registration successful. Your account is pending admin approval.", map[string]interface{}{
		"user_id": userID,
		"status":  "pending",
	})
}
