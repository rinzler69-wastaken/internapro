package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/gorilla/mux"
)

type OfficeHandler struct {
	db *sql.DB
}

func NewOfficeHandler(db *sql.DB) *OfficeHandler {
	return &OfficeHandler{db: db}
}

// ensureOfficeTable creates the office_locations table if it does not exist.
func ensureOfficeTable(db *sql.DB) error {
	_, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS office_locations (" +
			"id BIGINT AUTO_INCREMENT PRIMARY KEY," +
			"name VARCHAR(255) NOT NULL," +
			"latitude DECIMAL(10, 8) NOT NULL," +
			"longitude DECIMAL(11, 8) NOT NULL," +
			"radius_meters INT DEFAULT 100," +
			"address TEXT," +
			"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP," +
			"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" +
			") ENGINE=InnoDB;",
	)
	return err
}

func (h *OfficeHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok || middleware.NormalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Unauthorized")
		return
	}

	if err := ensureOfficeTable(h.db); err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	var payload struct {
		Name         string  `json:"name"`
		Latitude     float64 `json:"latitude"`
		Longitude    float64 `json:"longitude"`
		RadiusMeters int     `json:"radius_meters"`
		Address      string  `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	res, err := h.db.Exec(
		"INSERT INTO office_locations (name, latitude, longitude, radius_meters, address) VALUES (?, ?, ?, ?, ?)",
		payload.Name, payload.Latitude, payload.Longitude, payload.RadiusMeters, payload.Address,
	)
	if err != nil {
		log.Printf("Failed to create office location: %v", err)
		utils.RespondInternalError(w, "Failed to create location")
		return
	}

	id, _ := res.LastInsertId()
	utils.RespondSuccess(w, "Location created", map[string]int64{"id": id})
}

func (h *OfficeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	if err := ensureOfficeTable(h.db); err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	rows, err := h.db.Query("SELECT id, name, latitude, longitude, radius_meters, address, created_at, updated_at FROM office_locations ORDER BY name ASC")
	if err != nil {
		log.Printf("Failed to fetch office locations: %v", err)
		utils.RespondInternalError(w, "Failed to fetch locations")
		return
	}
	defer rows.Close()

	var locations []models.OfficeLocation
	for rows.Next() {
		var loc models.OfficeLocation
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.Latitude, &loc.Longitude, &loc.RadiusMeters, &loc.Address, &loc.CreatedAt, &loc.UpdatedAt); err != nil {
			continue
		}
		locations = append(locations, loc)
	}

	utils.RespondSuccess(w, "Locations retrieved", locations)
}

func (h *OfficeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok || middleware.NormalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid ID")
		return
	}

	// Check if this is the active office
	var activeIdStr string
	err = h.db.QueryRow("SELECT `value` FROM settings WHERE `key` = 'active_office_id'").Scan(&activeIdStr)
	if err == nil {
		activeId, _ := strconv.ParseInt(activeIdStr, 10, 64)
		if activeId == id {
			utils.RespondBadRequest(w, "Cannot delete the currently active office location")
			return
		}
	}

	_, err = h.db.Exec("DELETE FROM office_locations WHERE id = ?", id)
	if err != nil {
		log.Printf("Failed to delete office location: %v", err)
		utils.RespondInternalError(w, "Failed to delete location")
		return
	}

	utils.RespondSuccess(w, "Location deleted", nil)
}

func (h *OfficeHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok || middleware.NormalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Unauthorized")
		return
	}

	var payload struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// Verify ID exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM office_locations WHERE id = ?)", payload.ID).Scan(&exists)
	if err != nil || !exists {
		utils.RespondBadRequest(w, "Location not found or database error")
		return
	}

	val := fmt.Sprintf("%d", payload.ID)
	// Update settings
	_, err = h.db.Exec("INSERT INTO settings (`key`, `value`) VALUES ('active_office_id', ?) ON DUPLICATE KEY UPDATE `value` = VALUES(`value`)", val)
	if err != nil {
		log.Printf("Failed to set active office: %v", err)
		utils.RespondInternalError(w, "Failed to update settings")
		return
	}

	utils.RespondSuccess(w, "Active office updated", nil)
}
