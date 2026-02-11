package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"dsi_interna_sys/internal/config"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"time"

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
	if errUser == sql.ErrNoRows {
		// Auto-register as "new_user"
		// We insert the user with role 'new_user' and allow them to complete profile later
		res, err := h.db.Exec(
			"INSERT INTO users (name, email, role, google_id, provider, avatar, created_at) VALUES (?, ?, 'new_user', ?, 'google', ?, NOW())",
			info.Name, info.Email, info.ID, info.Picture,
		)
		if err != nil {
			utils.RespondInternalError(w, "Failed to create new user: "+err.Error())
			return
		}
		userID, _ := res.LastInsertId()

		// Fill the user struct for token generation
		user.ID = userID
		user.Name = sql.NullString{String: info.Name, Valid: true}
		user.Email = info.Email
		user.Role = "new_user"
		user.Avatar = sql.NullString{String: info.Picture, Valid: true}
		user.GoogleID = sql.NullString{String: info.ID, Valid: true}
		user.Provider = sql.NullString{String: "google", Valid: true}
		user.Is2FAEnabled = false
		user.CreatedAt = time.Now()

		setupRequired = true // Effectively they need setup (profile completion)
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
		// Always update avatar if available from Google
		if info.Picture != "" && (!user.Avatar.Valid || user.Avatar.String != info.Picture) {
			updates = append(updates, "avatar = ?")
			args = append(args, info.Picture)
		}

		if len(updates) > 0 {
			query := "UPDATE users SET " + strings.Join(updates, ", ") + " WHERE id = ?"
			args = append(args, user.ID)
			_, _ = h.db.Exec(query, args...)
		}
	}

	// Generate Token
	jwtToken, err := h.generateToken(&user)
	if err != nil {
		utils.RespondInternalError(w, "Failed to generate token")
		return
	}

	// Always redirect to frontend. Frontend determines next step based on user role ("new_user")
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

func sendRegisterRedirect(w http.ResponseWriter, r *http.Request, email, name string) {
	params := url.Values{}
	if email != "" {
		params.Set("email", email)
	}
	if name != "" {
		params.Set("name", name)
	}
	params.Set("oauth", "google_unregistered")

	redirectURL := config.Loaded.OAuth.FrontendURL + "/register?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
