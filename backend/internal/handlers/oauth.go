package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
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
	var errUser error
	query := `SELECT id, name, email, role, avatar, google_id, provider, is_2fa_enabled, created_at
	          FROM users WHERE google_id = ? OR email = ? LIMIT 1`
	errUser = h.db.QueryRow(query, info.ID, info.Email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.Avatar, &user.GoogleID, &user.Provider, &user.Is2FAEnabled, &user.CreatedAt,
	)

	setupRequired := false
	if errUser == sql.ErrNoRows {
		// Create user
		res, err := h.db.Exec(
			"INSERT INTO users (name, email, role, google_id, provider, avatar) VALUES (?, ?, ?, ?, ?, ?)",
			info.Name, info.Email, "intern", info.ID, "google", info.Picture,
		)
		if err != nil {
			utils.RespondInternalError(w, "Failed to create user")
			return
		}
		userID, _ := res.LastInsertId()
		user.ID = userID
		user.Email = info.Email
		user.Role = "intern"
		user.Name = sql.NullString{String: info.Name, Valid: info.Name != ""}
		user.Avatar = sql.NullString{String: info.Picture, Valid: info.Picture != ""}
		user.Is2FAEnabled = false
		user.CreatedAt = time.Now()
		setupRequired = true

		// Create minimal intern profile
		start := time.Now()
		end := time.Now().AddDate(0, 3, 0)
		_, _ = h.db.Exec(
			`INSERT INTO interns (user_id, full_name, start_date, end_date, status)
			 VALUES (?, ?, ?, ?, 'pending')`,
			userID, info.Name, start, end,
		)
	} else if errUser != nil {
		utils.RespondInternalError(w, "Database error")
		return
	} else {
		// Update google_id/provider/avatar if missing
		if !user.GoogleID.Valid || user.GoogleID.String == "" {
			_, _ = h.db.Exec("UPDATE users SET google_id = ?, provider = 'google' WHERE id = ?", info.ID, user.ID)
		}
		if info.Picture != "" {
			_, _ = h.db.Exec("UPDATE users SET avatar = ? WHERE id = ?", info.Picture, user.ID)
		}
	}

	// Block pending intern/supervisor
	role := normalizeRole(user.Role)
	if role == "intern" {
		var status string
		if err := h.db.QueryRow("SELECT status FROM interns WHERE user_id = ?", user.ID).Scan(&status); err == nil {
			if status == "pending" {
				utils.RespondForbidden(w, "Account pending approval")
				return
			}
		}
	}
	if role == "pembimbing" {
		var status string
		if err := h.db.QueryRow("SELECT status FROM supervisors WHERE user_id = ?", user.ID).Scan(&status); err == nil {
			if status == "pending" {
				utils.RespondForbidden(w, "Account pending approval")
				return
			}
		}
	}

	jwtToken, err := h.generateToken(&user)
	if err != nil {
		utils.RespondInternalError(w, "Failed to generate token")
		return
	}

	// optional redirect to frontend
	redirect := r.URL.Query().Get("redirect")
	if strings.TrimSpace(redirect) != "" {
		redirectURL := config.Loaded.OAuth.FrontendURL + redirect
		q := "?token=" + jwtToken
		if setupRequired {
			q += "&setup_required=1"
		}
		http.Redirect(w, r, redirectURL+q, http.StatusFound)
		return
	}

	if rc, err := r.Cookie("oauth_redirect"); err == nil && rc.Value != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_redirect",
			Value:    "",
			MaxAge:   -1,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   config.Loaded.App.Env == "production",
		})
		redirectURL := config.Loaded.OAuth.FrontendURL + rc.Value
		q := "?token=" + jwtToken
		if setupRequired {
			q += "&setup_required=1"
		}
		http.Redirect(w, r, redirectURL+q, http.StatusFound)
		return
	}

	utils.RespondSuccess(w, "OAuth login successful", LoginResponse{
		Token:         jwtToken,
		User:          toUserResponse(user),
		Require2FA:    user.Is2FAEnabled,
		SetupRequired: setupRequired,
	})
}
