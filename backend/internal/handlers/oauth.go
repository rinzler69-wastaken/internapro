package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dsi_interna_sys/internal/config"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func googleOAuthConfig() *oauth2.Config {
	cfg := config.Loaded
	return &oauth2.Config{
		ClientID:     cfg.OAuth.GoogleClientID,
		ClientSecret: cfg.OAuth.GoogleClientSecret,
		RedirectURL:  cfg.OAuth.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// StartGoogleOAuth returns a Google OAuth URL or redirects to it
func (h *AuthHandler) StartGoogleOAuth(w http.ResponseWriter, r *http.Request) {
	if config.Loaded.OAuth.GoogleClientID == "" || config.Loaded.OAuth.GoogleClientSecret == "" || config.Loaded.OAuth.GoogleRedirectURL == "" {
		utils.RespondInternalError(w, "Google OAuth is not configured")
		return
	}

	state, err := utils.GenerateToken(24)
	if err != nil {
		utils.RespondInternalError(w, "Failed to generate state")
		return
	}

	redirectPath := strings.TrimSpace(r.URL.Query().Get("redirect_path"))
	if redirectPath != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_redirect",
			Value:    redirectPath,
			MaxAge:   600,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   config.Loaded.App.Env == "production",
		})
	}

	// store state in cookie (10 min)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   600,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   config.Loaded.App.Env == "production",
	})

	url := googleOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline)

	if r.URL.Query().Get("redirect") == "1" {
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	utils.RespondSuccess(w, "OAuth URL generated", map[string]string{"url": url})
}

// HandleGoogleCallback handles OAuth callback and returns JWT
func (h *AuthHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value == "" || cookie.Value != state {
		utils.RespondBadRequest(w, "Invalid OAuth state")
		return
	}

	if code == "" {
		utils.RespondBadRequest(w, "Missing code")
		return
	}

	token, err := googleOAuthConfig().Exchange(context.Background(), code)
	if err != nil {
		utils.RespondInternalError(w, "Failed to exchange code")
		return
	}

	client := googleOAuthConfig().Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch user info")
		return
	}
	defer resp.Body.Close()

	var info struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		utils.RespondInternalError(w, "Failed to parse user info")
		return
	}

	if info.Email == "" {
		utils.RespondBadRequest(w, "Email not available from Google")
		return
	}

	var user models.User
	query := `SELECT id, name, email, role, avatar, google_id, provider, is_2fa_enabled, created_at
	          FROM users WHERE google_id = ? OR email = ? LIMIT 1`
	errUser := h.db.QueryRow(query, info.ID, info.Email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.Avatar, &user.GoogleID, &user.Provider, &user.Is2FAEnabled, &user.CreatedAt,
	)

	setupRequired := false
	avatarPath := ""
	if info.Picture != "" {
		if saved, err := saveRemoteAvatar(info.Picture); err == nil && saved != "" {
			avatarPath = saved
		} else {
			// fall back to remote URL so frontend can still show something
			avatarPath = info.Picture
		}
	}

	if errUser == sql.ErrNoRows {
		// Do NOT create a DB row. Redirect to register with prefill data.
		sendRegisterRedirect(w, r, info.Email, info.Name, info.Picture, info.ID)
		return
	} else if errUser != nil {
		utils.RespondInternalError(w, "Database error")
		return
	} else {
		// Update google_id/provider/avatar if missing
		updates := []string{}
		args := []interface{}{}

		if !user.GoogleID.Valid || user.GoogleID.String == "" {
			updates = append(updates, "google_id = ?")
			args = append(args, info.ID)
		}
		if !user.Provider.Valid || user.Provider.String == "" {
			updates = append(updates, "provider = 'google'")
		}
		// Always update avatar if available from Google (prefer saved local path)
		if avatarPath != "" && (!user.Avatar.Valid || user.Avatar.String != avatarPath) {
			updates = append(updates, "avatar = ?")
			args = append(args, avatarPath)
		}

		if len(updates) > 0 {
			query := "UPDATE users SET " + strings.Join(updates, ", ") + " WHERE id = ?"
			args = append(args, user.ID)
			_, _ = h.db.Exec(query, args...)
		}

		// If user was pending completion but already has an intern profile, upgrade role
		if strings.EqualFold(user.Role, "new_user") {
			var status string
			if err := h.db.QueryRow("SELECT status FROM interns WHERE user_id = ? LIMIT 1", user.ID).Scan(&status); err == nil {
				if status != "" {
					newRole := "intern"
					user.Role = newRole
					_, _ = h.db.Exec("UPDATE users SET role = ? WHERE id = ?", newRole, user.ID)
				}
			}
		}
	}

	normalizedRole := strings.ToLower(strings.TrimSpace(user.Role))
	if normalizedRole == "supervisor" {
		normalizedRole = "pembimbing"
		user.Role = "pembimbing"
	}

	// Enforce the same approval/status gates as password login.
	switch normalizedRole {
	case "intern":
		var internStatus string
		err := h.db.QueryRow("SELECT status FROM interns WHERE user_id = ? LIMIT 1", user.ID).Scan(&internStatus)
		if err == sql.ErrNoRows {
			sendRegisterRedirect(w, r, info.Email, info.Name, info.Picture, info.ID)
			return
		}
		if err != nil {
			utils.RespondInternalError(w, "Gagal memeriksa status magang")
			return
		}

		switch strings.ToLower(strings.TrimSpace(internStatus)) {
		case "active":
			// allow login
		case "pending":
			sendFrontendRedirect(w, r, "/waiting-approval")
			return
		default:
			sendFrontendRedirect(w, r, "/login?error="+url.QueryEscape("Akun magang tidak aktif. Hubungi admin."))
			return
		}
	case "pembimbing":
		var supervisorStatus string
		err := h.db.QueryRow("SELECT status FROM supervisors WHERE user_id = ? LIMIT 1", user.ID).Scan(&supervisorStatus)
		if err == sql.ErrNoRows {
			sendFrontendRedirect(w, r, "/login?error="+url.QueryEscape("Profil pembimbing belum lengkap. Hubungi admin."))
			return
		}
		if err != nil {
			utils.RespondInternalError(w, "Gagal memeriksa status pembimbing")
			return
		}

		switch strings.ToLower(strings.TrimSpace(supervisorStatus)) {
		case "active":
			// allow login
		case "pending":
			sendFrontendRedirect(w, r, "/waiting-approval")
			return
		default:
			sendFrontendRedirect(w, r, "/login?error="+url.QueryEscape("Akun pembimbing tidak aktif. Hubungi admin."))
			return
		}
	}

	// Generate Token
	jwtToken, err := h.generateToken(&user)
	if err != nil {
		utils.RespondInternalError(w, "Failed to generate token")
		return
	}

	redirect := r.URL.Query().Get("redirect")
	if strings.TrimSpace(redirect) == "" {
		if rc, err := r.Cookie("oauth_redirect"); err == nil && rc.Value != "" {
			redirect = rc.Value
			// Clear cookie
			http.SetCookie(w, &http.Cookie{
				Name:   "oauth_redirect",
				Value:  "",
				MaxAge: -1,
				Path:   "/",
			})
		}
	}

	// Force new_user to /register
	if user.Role == "new_user" {
		redirect = "/register"
	}

	if strings.TrimSpace(redirect) == "" {
		redirect = "/dashboard"
	}

	// Construct Redirect URL with Token
	finalURL := config.Loaded.OAuth.FrontendURL + redirect
	q := "?token=" + jwtToken
	if setupRequired {
		q += "&setup_required=1"
	}
	// If new_user, the frontend auth check will handle redirection to /register
	finalURL += q

	http.Redirect(w, r, finalURL, http.StatusFound)
}

func sendRegisterRedirect(w http.ResponseWriter, r *http.Request, email, name, avatar, googleID string) {
	params := url.Values{}
	if email != "" {
		params.Set("email", email)
	}
	if name != "" {
		params.Set("name", name)
	}
	if avatar != "" {
		params.Set("avatar", avatar)
	}
	if googleID != "" {
		params.Set("google_id", googleID)
	}
	params.Set("oauth", "google_unregistered")

	redirectURL := config.Loaded.OAuth.FrontendURL + "/register?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func sendFrontendRedirect(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, config.Loaded.OAuth.FrontendURL+path, http.StatusFound)
}

// saveRemoteAvatar downloads an image from a remote URL and stores it in the uploads/avatars directory.
// Returns the relative path (e.g., "avatars/abc.jpg") or an empty string on failure.
func saveRemoteAvatar(pictureURL string) (string, error) {
	if pictureURL == "" {
		return "", nil
	}

	cfg := config.Loaded

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(pictureURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("avatar fetch returned status %d", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, cfg.Upload.MaxSize)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read avatar: %w", err)
	}

	// Detect content type to pick extension
	contentType := http.DetectContentType(data)
	ext := ".jpg"
	switch {
	case strings.Contains(contentType, "png"):
		ext = ".png"
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		ext = ".jpg"
	}

	// Ensure upload dir exists
	uploadDir := filepath.Join(cfg.Upload.Dir, "avatars")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload dir: %w", err)
	}

	filename := fmt.Sprintf("google_%d%s", time.Now().UnixNano(), ext)
	fullPath := filepath.Join(uploadDir, filename)

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save avatar: %w", err)
	}

	// Return relative path for serving
	return filepath.Join("avatars", filename), nil
}
