package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/gorilla/mux"
)

type NotificationHandler struct {
	db *sql.DB
}

func NewNotificationHandler(db *sql.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

func (h *NotificationHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	if err := h.db.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ?", claims.UserID).Scan(&total); err != nil {
		// Fail-soft: return empty list instead of 500 so UI keeps working
		utils.RespondPaginated(w, []models.Notification{}, utils.CalculatePagination(page, limit, 0))
		return
	}

	rows, err := h.db.Query(
		`SELECT id, user_id, type, title, message, icon, link, data, read_at, created_at, updated_at
		 FROM notifications
		 WHERE user_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		claims.UserID, limit, offset,
	)
	if err != nil {
		utils.RespondPaginated(w, []models.Notification{}, utils.CalculatePagination(page, limit, 0))
		return
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.Icon, &n.Link, &n.Data, &n.ReadAt, &n.CreatedAt, &n.UpdatedAt); err == nil {
			notifications = append(notifications, n)
		}
	}

	utils.RespondPaginated(w, notifications, utils.CalculatePagination(page, limit, total))
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	res, err := h.db.Exec(
		"UPDATE notifications SET read_at = ? WHERE id = ? AND user_id = ?",
		time.Now(), id, claims.UserID,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to update notification")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		utils.RespondNotFound(w, "Notification not found")
		return
	}

	utils.RespondSuccess(w, "Notification marked as read", nil)
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	if _, err := h.db.Exec(
		"UPDATE notifications SET read_at = ? WHERE user_id = ? AND read_at IS NULL",
		time.Now(), claims.UserID,
	); err != nil {
		utils.RespondInternalError(w, "Failed to mark all notifications")
		return
	}

	utils.RespondSuccess(w, "All notifications marked as read", nil)
}

func (h *NotificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	res, err := h.db.Exec("DELETE FROM notifications WHERE id = ? AND user_id = ?", id, claims.UserID)
	if err != nil {
		utils.RespondInternalError(w, "Failed to delete notification")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		utils.RespondNotFound(w, "Notification not found")
		return
	}

	utils.RespondSuccess(w, "Notification deleted", nil)
}
