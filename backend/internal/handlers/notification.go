package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

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
	queryCount := "SELECT COUNT(*) FROM notifications WHERE user_id = ?"
	queryData := `SELECT id, user_id, type, title, message, icon, link, is_read, created_at 
	             FROM notifications WHERE user_id = ?`

	isReadParam := r.URL.Query().Get("is_read")
	var argsCount []interface{}
	var argsData []interface{}
	argsCount = append(argsCount, claims.UserID)
	argsData = append(argsData, claims.UserID)

	if isReadParam != "" {
		isRead, _ := strconv.ParseBool(isReadParam)
		queryCount += " AND is_read = ?"
		queryData += " AND is_read = ?"
		argsCount = append(argsCount, isRead)
		argsData = append(argsData, isRead)
	}

	if err := h.db.QueryRow(queryCount, argsCount...).Scan(&total); err != nil {
		utils.RespondPaginated(w, []models.Notification{}, utils.CalculatePagination(page, limit, 0))
		return
	}

	queryData += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	argsData = append(argsData, limit, offset)

	rows, err := h.db.Query(queryData, argsData...)
	if err != nil {
		utils.RespondPaginated(w, []models.Notification{}, utils.CalculatePagination(page, limit, 0))
		return
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		var icon, link sql.NullString
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &icon, &link, &n.IsRead, &n.CreatedAt); err == nil {
			n.Icon = icon.String
			n.Link = link.String
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
		"UPDATE notifications SET is_read = 1 WHERE id = ? AND user_id = ?",
		id, claims.UserID,
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
		"UPDATE notifications SET is_read = 1 WHERE user_id = ? AND is_read = 0",
		claims.UserID,
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
