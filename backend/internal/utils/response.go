package utils

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// RespondSuccess sends a success response
func RespondSuccess(w http.ResponseWriter, message string, data interface{}) {
	RespondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// RespondCreated sends a created response
func RespondCreated(w http.ResponseWriter, message string, data interface{}) {
	RespondJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, Response{
		Success: false,
		Error:   message,
	})
}

// RespondBadRequest sends a bad request error
func RespondBadRequest(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusBadRequest, message)
}

// RespondUnauthorized sends an unauthorized error
func RespondUnauthorized(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusUnauthorized, message)
}

// RespondForbidden sends a forbidden error
func RespondForbidden(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusForbidden, message)
}

// RespondNotFound sends a not found error
func RespondNotFound(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusNotFound, message)
}

// RespondInternalError sends an internal server error
func RespondInternalError(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusInternalServerError, message)
}

// RespondPaginated sends a paginated response
func RespondPaginated(w http.ResponseWriter, data interface{}, pagination Pagination) {
	RespondJSON(w, http.StatusOK, PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
	})
}

// ParseJSON parses JSON request body
func ParseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit int, totalItems int64) Pagination {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	
	totalPages := int(totalItems) / limit
	if int(totalItems)%limit > 0 {
		totalPages++
	}
	
	return Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
