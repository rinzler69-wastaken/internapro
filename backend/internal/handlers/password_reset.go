package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"dsi_interna_sys/internal/config"
	"dsi_interna_sys/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type PasswordResetHandler struct {
	db *sql.DB
}

func NewPasswordResetHandler(db *sql.DB) *PasswordResetHandler {
	return &PasswordResetHandler{db: db}
}

func (h *PasswordResetHandler) RequestReset(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	email := strings.TrimSpace(strings.ToLower(payload.Email))
	if email == "" {
		utils.RespondBadRequest(w, "Email is required")
		return
	}

	// Always respond success to avoid leaking user existence
	var userCount int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&userCount)
	if userCount == 0 {
		utils.RespondSuccess(w, "If the email exists, a reset link will be sent", nil)
		return
	}

	token, err := utils.GenerateToken(32)
	if err != nil {
		utils.RespondInternalError(w, "Failed to generate token")
		return
	}

	hash := sha256.Sum256([]byte(token))
	hashStr := hex.EncodeToString(hash[:])

	// Clear old tokens
	_, _ = h.db.Exec("DELETE FROM password_resets WHERE email = ?", email)
	_, err = h.db.Exec(
		"INSERT INTO password_resets (email, token_hash, created_at) VALUES (?, ?, ?)",
		email, hashStr, time.Now(),
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to save reset token")
		return
	}

	resetURL := config.Loaded.OAuth.FrontendURL + "/reset-password?email=" + email + "&token=" + token
	subject := "Reset Password"
	body := "Gunakan tautan berikut untuk mengatur ulang password Anda:\n\n" + resetURL + "\n\nTautan ini berlaku selama 1 jam."

	sent := false
	if config.Loaded.SMTP.Host != "" && config.Loaded.SMTP.From != "" {
		if err := utils.SendMail(email, subject, body); err == nil {
			sent = true
		} else {
			log.Printf("Email sending failed to %s: %v", email, err)
		}
	}

	responsePayload := map[string]interface{}{
		"sent": sent,
	}
	if !sent && config.Loaded.App.Env != "production" {
		responsePayload["token"] = token
		responsePayload["reset_url"] = resetURL
	}

	utils.RespondSuccess(w, "If the email exists, a reset link will be sent", responsePayload)
}

func (h *PasswordResetHandler) Reset(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email                string `json:"email"`
		Token                string `json:"token"`
		Password             string `json:"password"`
		PasswordConfirmation string `json:"password_confirmation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	email := strings.TrimSpace(strings.ToLower(payload.Email))
	if email == "" || payload.Token == "" || payload.Password == "" {
		utils.RespondBadRequest(w, "Email, token, and password are required")
		return
	}
	if payload.Password != payload.PasswordConfirmation {
		utils.RespondBadRequest(w, "Password confirmation does not match")
		return
	}
	if len(payload.Password) < 6 {
		utils.RespondBadRequest(w, "Password must be at least 6 characters")
		return
	}

	// Check token
	hash := sha256.Sum256([]byte(payload.Token))
	hashStr := hex.EncodeToString(hash[:])

	var createdAt time.Time
	err := h.db.QueryRow(
		"SELECT created_at FROM password_resets WHERE email = ? AND token_hash = ?",
		email, hashStr,
	).Scan(&createdAt)
	if err == sql.ErrNoRows {
		utils.RespondBadRequest(w, "Invalid token")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	if time.Since(createdAt) > time.Hour {
		utils.RespondBadRequest(w, "Token expired")
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondInternalError(w, "Failed to hash password")
		return
	}

	if _, err := h.db.Exec("UPDATE users SET password_hash = ? WHERE email = ?", string(newHash), email); err != nil {
		utils.RespondInternalError(w, "Failed to update password")
		return
	}

	_, _ = h.db.Exec("DELETE FROM password_resets WHERE email = ?", email)

	utils.RespondSuccess(w, "Password updated", nil)
}
