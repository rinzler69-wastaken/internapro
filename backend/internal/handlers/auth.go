package handlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"net/http"
	"strings"
	"time"

	"dsi_interna_sys/internal/config"
	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"` // admin, pembimbing/supervisor, intern
	FullName string `json:"full_name"`

	// Optional fields for intern
	StudentID     string `json:"student_id,omitempty"`
	InstitutionID int64  `json:"institution_id,omitempty"`
	SupervisorID  int64  `json:"supervisor_id,omitempty"`
	StartDate     string `json:"start_date,omitempty"`
	EndDate       string `json:"end_date,omitempty"`

	// Optional fields for supervisor
	NIP      string `json:"nip,omitempty"`
	Position string `json:"position,omitempty"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code,omitempty"` // Required if 2FA enabled
}

// LoginResponse represents login response
type LoginResponse struct {
	Token         string       `json:"token"`
	User          UserResponse `json:"user"`
	Require2FA    bool         `json:"require_2fa"`
	SetupRequired bool         `json:"setup_required"` // <--- ADD THIS
}

// UserResponse is a safe JSON shape for frontend consumption
type UserResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	Avatar         string    `json:"avatar,omitempty"`
	Is2FAEnabled   bool      `json:"is_2fa_enabled"`
	CreatedAt      time.Time `json:"created_at"`
	InternID       int64     `json:"intern_id,omitempty"`       // For interns
	SupervisorID   int64     `json:"supervisor_id,omitempty"`   // For interns
	SupervisorName string    `json:"supervisor_name,omitempty"` // For interns
}

// Setup2FAResponse represents 2FA setup response
type Setup2FAResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

// Verify2FARequest represents 2FA verification request
type Verify2FARequest struct {
	Code string `json:"code"`
}

// Register creates a new user
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Role == "" || req.FullName == "" {
		utils.RespondBadRequest(w, "Missing required fields")
		return
	}

	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role == "supervisor" {
		role = "pembimbing"
	}

	// Validate role
	if role != "admin" && role != "pembimbing" && role != "intern" {
		utils.RespondBadRequest(w, "Invalid role. Must be admin, pembimbing, or intern")
		return
	}

	// Validate password length
	if len(req.Password) < 6 {
		utils.RespondBadRequest(w, "Password must be at least 6 characters")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondInternalError(w, "Failed to hash password")
		return
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Insert user
	result, err := tx.Exec(
		"INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, ?)",
		req.FullName, req.Email, string(hashedPassword), role,
	)
	if err != nil {
		utils.RespondBadRequest(w, "Email already exists")
		return
	}

	userID, _ := result.LastInsertId()

	// Create role-specific record
	switch role {
	case "pembimbing":
		_, err = tx.Exec(
			"INSERT INTO supervisors (user_id, full_name, nip, position) VALUES (?, ?, ?, ?)",
			userID, req.FullName, req.NIP, req.Position,
		)
	case "intern":
		if req.InstitutionID == 0 || req.SupervisorID == 0 {
			utils.RespondBadRequest(w, "Institution and supervisor are required for interns")
			return
		}

		startDate := time.Now()
		endDate := time.Now().AddDate(0, 3, 0) // Default 3 months

		if req.StartDate != "" {
			startDate, _ = time.Parse("2006-01-02", req.StartDate)
		}
		if req.EndDate != "" {
			endDate, _ = time.Parse("2006-01-02", req.EndDate)
		}

		_, err = tx.Exec(
			`INSERT INTO interns (user_id, institution_id, supervisor_id, full_name, student_id, start_date, end_date, status) 
			 VALUES (?, ?, ?, ?, ?, ?, ?, 'active')`,
			userID, req.InstitutionID, req.SupervisorID, req.FullName, req.StudentID, startDate, endDate,
		)
	}

	if err != nil {
		utils.RespondInternalError(w, "Failed to create user profile")
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to commit transaction")
		return
	}

	utils.RespondCreated(w, "User registered successfully", map[string]interface{}{
		"user_id": userID,
		"email":   req.Email,
		"role":    role,
	})
}

// Login authenticates user and returns JWT token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// Find user
	var user models.User
	err := h.db.QueryRow(
		"SELECT id, name, email, password_hash, role, totp_secret, is_2fa_enabled, provider FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.TOTPSecret, &user.Is2FAEnabled, &user.Provider)

	if err != nil {
		utils.RespondUnauthorized(w, "Invalid email or password")
		return
	}

	// Verify password
	if !user.PasswordHash.Valid {
		// Allow first-time password setup for accounts without a password (e.g., created via OAuth)
		if req.Password == "" {
			utils.RespondUnauthorized(w, "Password login not available for this account")
			return
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondInternalError(w, "Failed to process password")
			return
		}
		if _, err := h.db.Exec("UPDATE users SET password_hash = ?, provider = COALESCE(provider, 'local') WHERE id = ?", string(hashed), user.ID); err != nil {
			utils.RespondInternalError(w, "Failed to save password")
			return
		}
		user.PasswordHash = sql.NullString{String: string(hashed), Valid: true}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.Password)); err != nil {
		utils.RespondUnauthorized(w, "Invalid email or password")
		return
	}

	// 1. Check if 2FA is Enabled (Existing Logic)
	if user.Is2FAEnabled {
		if req.TOTPCode == "" {
			utils.RespondSuccess(w, "2FA Code Required", LoginResponse{
				Require2FA: true,
			})
			return
		}

		if !user.TOTPSecret.Valid {
			utils.RespondUnauthorized(w, "2FA is enabled but no secret is configured")
			return
		}
		valid := totp.Validate(req.TOTPCode, user.TOTPSecret.String)
		if !valid {
			utils.RespondUnauthorized(w, "Invalid 2FA code")
			return
		}
	}

	// Normalize role, auto-upgrade new_user if intern data exists
	role := strings.ToLower(strings.TrimSpace(user.Role))
	if role == "new_user" {
		var status string
		err := h.db.QueryRow("SELECT status FROM interns WHERE user_id = ? LIMIT 1", user.ID).Scan(&status)
		if err == nil {
			// Upgrade role in DB and in-memory; status checks below will gate access
			role = "intern"
			user.Role = "intern"
			_, _ = h.db.Exec("UPDATE users SET role = 'intern' WHERE id = ?", user.ID)
		}
	}

	// 2. Block accounts that are not approved/active (intern & pembimbing/supervisor)
	if role == "intern" {
		var status string
		err := h.db.QueryRow("SELECT status FROM interns WHERE user_id = ? LIMIT 1", user.ID).Scan(&status)
		if err == sql.ErrNoRows {
			utils.RespondForbidden(w, "Profil magang belum lengkap. Hubungi admin untuk melengkapi data.")
			return
		}
		if err != nil {
			utils.RespondInternalError(w, "Gagal memeriksa status magang")
			return
		}

		if status != "active" {
			if status == "pending" {
				utils.RespondForbidden(w, "Akun Anda belum disetujui admin. Silakan tunggu persetujuan sebelum login.")
				return
			}
			utils.RespondForbidden(w, "Akun magang tidak aktif. Hubungi admin.")
			return
		}
	}

	if role == "pembimbing" || role == "supervisor" {
		var status string
		err := h.db.QueryRow("SELECT status FROM supervisors WHERE user_id = ? LIMIT 1", user.ID).Scan(&status)
		if err == sql.ErrNoRows {
			utils.RespondForbidden(w, "Profil pembimbing belum lengkap. Hubungi admin.")
			return
		}
		if err != nil {
			utils.RespondInternalError(w, "Gagal memeriksa status pembimbing")
			return
		}
		if status == "pending" {
			utils.RespondForbidden(w, "Akun pembimbing belum disetujui admin. Silakan tunggu persetujuan.")
			return
		}
	}

	// Generate JWT token
	token, err := h.generateToken(&user)
	if err != nil {
		utils.RespondInternalError(w, "Failed to generate token")
		return
	}

	// 2. Determine if Setup is Required (New Logic)
	// If 2FA is NOT enabled, we mark SetupRequired as true
	setupRequired := !user.Is2FAEnabled

	resp := toUserResponse(user)
	enrichInternResponse(h.db, &resp)

	utils.RespondSuccess(w, "Login successful", LoginResponse{
		Token:         token,
		User:          resp,
		Require2FA:    false,
		SetupRequired: setupRequired, // <--- SEND THE FLAG
	})
}

// Setup2FA generates TOTP secret and QR code image
func (h *AuthHandler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "INTERNA",
		AccountName: claims.Email,
	})
	if err != nil {
		utils.RespondInternalError(w, "Failed to generate 2FA secret")
		return
	}

	// Generate the QR Code Image
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		utils.RespondInternalError(w, "Failed to generate QR image")
		return
	}

	err = png.Encode(&buf, img)
	if err != nil {
		utils.RespondInternalError(w, "Failed to encode QR image")
		return
	}

	qrCodeBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	// Save secret to database
	res, err := h.db.Exec(
		"UPDATE users SET totp_secret = ? WHERE id = ?",
		key.Secret(), claims.UserID,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to save 2FA secret")
		return
	}

	// CHECK ROWS AFFECTED
	rows, _ := res.RowsAffected()
	if rows == 0 {
		utils.RespondInternalError(w, "Failed to initialize 2FA: User not found")
		return
	}

	utils.RespondSuccess(w, "2FA setup initiated. Scan the QR code with Google Authenticator", Setup2FAResponse{
		Secret: key.Secret(),
		QRCode: qrCodeBase64,
	})
}

// Verify2FA verifies and enables 2FA
func (h *AuthHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	var req Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// 1. Get the secret from the DB (Verify persistence)
	var secret sql.NullString
	err := h.db.QueryRow("SELECT totp_secret FROM users WHERE id = ?", claims.UserID).Scan(&secret)
	if err != nil {
		utils.RespondInternalError(w, "User not found")
		return
	}

	if !secret.Valid || secret.String == "" {
		// CRITICAL: This means Setup2FA didn't save the secret!
		utils.RespondBadRequest(w, "2FA Setup not initialized. Please try generating the QR code again.")
		return
	}

	// 2. Verify code
	valid := totp.Validate(req.Code, secret.String)
	if !valid {
		utils.RespondBadRequest(w, "Invalid 2FA code")
		return
	}

	// 3. Enable 2FA (Strict Update)
	res, err := h.db.Exec("UPDATE users SET is_2fa_enabled = 1 WHERE id = ?", claims.UserID)
	if err != nil {
		utils.RespondInternalError(w, "Database error enabling 2FA")
		return
	}

	// 4. CHECK ROWS AFFECTED (The Fix)
	rows, err := res.RowsAffected()
	if err != nil {
		utils.RespondInternalError(w, "Failed to verify database update")
		return
	}

	if rows == 0 {
		// If we are here, it means the Update found NO matching user to update.
		// This usually means claims.UserID is wrong or the user was deleted.
		utils.RespondInternalError(w, "Failed to enable 2FA: User record could not be updated")
		return
	}

	utils.RespondSuccess(w, "2FA enabled successfully", nil)
}

// Disable2FA disables 2FA for user
func (h *AuthHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	_, err := h.db.Exec(
		"UPDATE users SET is_2fa_enabled = FALSE, totp_secret = NULL WHERE id = ?",
		claims.UserID,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to disable 2FA")
		return
	}

	utils.RespondSuccess(w, "2FA disabled successfully", nil)
}

// GetCurrentUser returns current user information
// GetCurrentUser returns current user information
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	// DISABLE CACHING: Force browser to ask server every time
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	var user models.User
	err := h.db.QueryRow(
		"SELECT id, name, email, role, avatar, is_2fa_enabled, created_at FROM users WHERE id = ?",
		claims.UserID,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.Avatar, &user.Is2FAEnabled, &user.CreatedAt)

	if err != nil {
		utils.RespondNotFound(w, "User not found")
		return
	}

	resp := toUserResponse(user)
	enrichInternResponse(h.db, &resp)

	utils.RespondSuccess(w, "User retrieved successfully", resp)
}

// Logout handles logout (mainly client-side token removal)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// In a stateless JWT system, logout is mainly handled client-side
	// You could implement token blacklisting here if needed
	utils.RespondSuccess(w, "Logged out successfully", nil)
}

// generateToken creates a new JWT token
func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	cfg := config.Loaded

	claims := middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWT.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

func toUserResponse(user models.User) UserResponse {
	resp := UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		Role:         user.Role,
		Is2FAEnabled: user.Is2FAEnabled,
		CreatedAt:    user.CreatedAt,
	}
	if user.Name.Valid {
		resp.Name = user.Name.String
	}
	if user.Avatar.Valid {
		resp.Avatar = prependUploadPath(user.Avatar.String)
	}
	return resp
}

func prependUploadPath(path string) string {
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

func enrichInternResponse(db *sql.DB, resp *UserResponse) {
	if resp.Role == "intern" {
		var internID, supID int64
		var supName string
		_ = db.QueryRow(`
			SELECT i.id, i.supervisor_id, u.name 
			FROM interns i 
			LEFT JOIN users u ON i.supervisor_id = u.id 
			WHERE i.user_id = ?`, resp.ID).Scan(&internID, &supID, &supName)
		resp.InternID = internID
		resp.SupervisorID = supID
		resp.SupervisorName = supName
	}
}
