package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/gorilla/mux"
)

type AgendaHandler struct {
	db *sql.DB
}

func NewAgendaHandler(db *sql.DB) *AgendaHandler {
	return &AgendaHandler{db: db}
}

func (h *AgendaHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	month := r.URL.Query().Get("month") // YYYY-MM
	date := r.URL.Query().Get("date")   // YYYY-MM-DD

	query := "SELECT id, user_id, title, description, date, time, is_notified, created_at, updated_at FROM agendas WHERE user_id = ?"
	args := []interface{}{claims.UserID}

	if date != "" {
		query += " AND date = ?"
		args = append(args, date)
	} else if month != "" {
		query += " AND DATE_FORMAT(date, '%Y-%m') = ?"
		args = append(args, month)
	}

	query += " ORDER BY date ASC, time ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch agendas: "+err.Error())
		return
	}
	defer rows.Close()

	var agendas []models.Agenda
	for rows.Next() {
		var a models.Agenda
		var desc sql.NullString
		var timeStr string
		var dateStr string

		if err := rows.Scan(&a.ID, &a.UserID, &a.Title, &desc, &dateStr, &timeStr, &a.IsNotified, &a.CreatedAt, &a.UpdatedAt); err != nil {
			utils.RespondInternalError(w, "Failed to scan agenda")
			return
		}
		a.Description = desc.String
		a.Date = dateStr
		a.Time = timeStr
		agendas = append(agendas, a)
	}

	utils.RespondSuccess(w, "Agendas fetched successfully", agendas)
}

func (h *AgendaHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Date        string `json:"date"`
		Time        string `json:"time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	if req.Title == "" || req.Date == "" || req.Time == "" {
		utils.RespondBadRequest(w, "Title, Date, and Time are required")
		return
	}

	res, err := h.db.Exec(
		"INSERT INTO agendas (user_id, title, description, date, time) VALUES (?, ?, ?, ?, ?)",
		claims.UserID, req.Title, req.Description, req.Date, req.Time,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to create agenda: "+err.Error())
		return
	}

	id, _ := res.LastInsertId()
	utils.RespondCreated(w, "Agenda created successfully", map[string]interface{}{"id": id})
}

func (h *AgendaHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Date        string `json:"date"`
		Time        string `json:"time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	_, err := h.db.Exec(
		"UPDATE agendas SET title = ?, description = ?, date = ?, time = ? WHERE id = ? AND user_id = ?",
		req.Title, req.Description, req.Date, req.Time, id, claims.UserID,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to update agenda: "+err.Error())
		return
	}

	utils.RespondSuccess(w, "Agenda updated successfully", nil)
}

func (h *AgendaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	_, err := h.db.Exec("DELETE FROM agendas WHERE id = ? AND user_id = ?", id, claims.UserID)
	if err != nil {
		utils.RespondInternalError(w, "Failed to delete agenda: "+err.Error())
		return
	}

	utils.RespondSuccess(w, "Agenda deleted successfully", nil)
}
