package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type ProfileHandler struct {
	db *sql.DB
}

func NewProfileHandler(db *sql.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	var user models.User
	if err := h.db.QueryRow(
		"SELECT id, email, role, name, avatar, is_2fa_enabled, created_at FROM users WHERE id = ?",
		claims.UserID,
	).Scan(&user.ID, &user.Email, &user.Role, &user.Name, &user.Avatar, &user.Is2FAEnabled, &user.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			utils.RespondUnauthorized(w, "User not found or session expired")
			return
		}
		utils.RespondInternalError(w, "Database error")
		return
	}

	var intern models.Intern
	var gender, status sql.NullString

	err := h.db.QueryRow(
		`SELECT id, user_id, institution_id, supervisor_id, full_name, nis, student_id, school, department, date_of_birth,
		        gender, phone, address, start_date, end_date, status, certificate_number, certificate_issued_at, created_at, updated_at
		 FROM interns WHERE user_id = ?`,
		claims.UserID,
	).Scan(
		&intern.ID, &intern.UserID, &intern.InstitutionID, &intern.SupervisorID, &intern.FullName, &intern.NIS, &intern.StudentID,
		&intern.School, &intern.Department, &intern.DateOfBirth, &gender, &intern.Phone, &intern.Address,
		&intern.StartDate, &intern.EndDate, &status, &intern.CertificateNumber, &intern.CertificateIssuedAt,
		&intern.CreatedAt, &intern.UpdatedAt,
	)

	if gender.Valid {
		intern.Gender = gender.String
	}
	if status.Valid {
		intern.Status = status.String
	}
	resp := toUserResponse(user)
	enrichInternResponse(h.db, &resp)

	if err == sql.ErrNoRows {
		utils.RespondSuccess(w, "Profile retrieved", map[string]interface{}{
			"user":   resp,
			"intern": intern,
		})
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	utils.RespondSuccess(w, "Profile retrieved", map[string]interface{}{
		"user":   resp,
		"intern": intern,
	})
}

func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	contentType := r.Header.Get("Content-Type")
	name := ""
	email := ""
	avatarPath := ""

	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.RespondBadRequest(w, "Failed to parse form data")
			return
		}
		name = strings.TrimSpace(r.FormValue("name"))
		email = strings.TrimSpace(r.FormValue("email"))

		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			path, err := utils.UploadFile(file, header, "avatars")
			if err != nil {
				utils.RespondBadRequest(w, "Upload failed: "+err.Error())
				return
			}
			avatarPath = path
		}
	} else {
		var payload struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			utils.RespondBadRequest(w, "Invalid request body")
			return
		}
		name = strings.TrimSpace(payload.Name)
		email = strings.TrimSpace(payload.Email)
	}

	if name == "" || email == "" {
		utils.RespondBadRequest(w, "Name and email are required")
		return
	}

	var currentEmail string
	_ = h.db.QueryRow("SELECT email FROM users WHERE id = ?", claims.UserID).Scan(&currentEmail)
	if email != currentEmail {
		var exists int
		_ = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
		if exists > 0 {
			utils.RespondBadRequest(w, "Email already exists")
			return
		}
	}

	updates := []string{"name = ?", "email = ?"}
	args := []interface{}{name, email}
	if avatarPath != "" {
		updates = append(updates, "avatar = ?")
		args = append(args, avatarPath)
	}
	args = append(args, claims.UserID)

	query := "UPDATE users SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	if _, err := h.db.Exec(query, args...); err != nil {
		utils.RespondInternalError(w, "Failed to update profile")
		return
	}

	utils.RespondSuccess(w, "Profile updated", nil)
}

func (h *ProfileHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	var payload struct {
		CurrentPassword string `json:"current_password"`
		Password        string `json:"password"`
		PasswordConfirm string `json:"password_confirmation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}
	if payload.CurrentPassword == "" || payload.Password == "" {
		utils.RespondBadRequest(w, "Current password and new password are required")
		return
	}
	if payload.Password != payload.PasswordConfirm {
		utils.RespondBadRequest(w, "Password confirmation does not match")
		return
	}
	if len(payload.Password) < 6 {
		utils.RespondBadRequest(w, "Password must be at least 6 characters")
		return
	}

	var hashed string
	if err := h.db.QueryRow("SELECT password_hash FROM users WHERE id = ?", claims.UserID).Scan(&hashed); err != nil {
		utils.RespondInternalError(w, "User not found")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(payload.CurrentPassword)); err != nil {
		utils.RespondBadRequest(w, "Current password is incorrect")
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondInternalError(w, "Failed to hash password")
		return
	}

	if _, err := h.db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(newHash), claims.UserID); err != nil {
		utils.RespondInternalError(w, "Failed to update password")
		return
	}

	utils.RespondSuccess(w, "Password updated", nil)
}
