package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"
)

type SettingHandler struct {
	db *sql.DB
}

func NewSettingHandler(db *sql.DB) *SettingHandler {
	return &SettingHandler{db: db}
}

func (h *SettingHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if middleware.NormalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can view settings")
		return
	}

	rows, err := h.db.Query("SELECT id, `key`, `value`, `type`, description, created_at, updated_at FROM settings")
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch settings")
		return
	}
	defer rows.Close()

	var settings []models.Setting
	for rows.Next() {
		var s models.Setting
		if err := rows.Scan(&s.ID, &s.Key, &s.Value, &s.Type, &s.Description, &s.CreatedAt, &s.UpdatedAt); err == nil {
			settings = append(settings, s)
		}
	}

	utils.RespondSuccess(w, "Settings retrieved", settings)
}

func (h *SettingHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if middleware.NormalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can update settings")
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		utils.RespondInternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	for key, value := range payload {
		val := fmt.Sprint(value)
		query := "INSERT INTO settings (`key`, `value`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `value` = VALUES(`value`)"
		if _, err := tx.Exec(query, key, val); err != nil {
			utils.RespondInternalError(w, "Failed to update settings")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.RespondInternalError(w, "Failed to commit settings")
		return
	}

	utils.RespondSuccess(w, "Settings updated", nil)
}
