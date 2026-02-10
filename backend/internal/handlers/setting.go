package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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

// ensureSettingsTable creates the settings table if it does not exist.
func ensureSettingsTable(db *sql.DB) error {
	_, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS settings (" +
			"id BIGINT AUTO_INCREMENT PRIMARY KEY," +
			"`key` VARCHAR(255) UNIQUE NOT NULL," +
			"`value` TEXT," +
			"`type` VARCHAR(50) DEFAULT 'string'," +
			"description VARCHAR(255)," +
			"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP," +
			"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" +
		") ENGINE=InnoDB;",
	)
	return err
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

	if err := ensureSettingsTable(h.db); err != nil {
		log.Printf("settings ensure table failed: %v", err)
		utils.RespondInternalError(w, "Failed to prepare settings table: "+err.Error())
		return
	}

	// Select only core columns to tolerate older settings table schemas.
	rows, err := h.db.Query("SELECT id, `key`, `value` FROM settings")
	if err != nil {
		log.Printf("settings query failed: %v", err)
		utils.RespondInternalError(w, "Failed to fetch settings: "+err.Error())
		return
	}
	defer rows.Close()

	var settings []models.Setting
	for rows.Next() {
		var s models.Setting
		if err := rows.Scan(&s.ID, &s.Key, &s.Value); err == nil {
			// Type/Description/CreatedAt/UpdatedAt may not exist in older schemas.
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

	if err := ensureSettingsTable(h.db); err != nil {
		log.Printf("settings ensure table failed: %v", err)
		utils.RespondInternalError(w, "Settings table unavailable: "+err.Error())
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
			log.Printf("settings update failed for key %s: %v", key, err)
			utils.RespondInternalError(w, "Failed to update settings")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("settings commit failed: %v", err)
		utils.RespondInternalError(w, "Failed to commit settings")
		return
	}

	utils.RespondSuccess(w, "Settings updated", nil)
}
