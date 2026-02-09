package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/gorilla/mux"
)

type AssessmentHandler struct {
	db *sql.DB
}

type updateAssessmentPayload struct {
	QualityScore       *int   `json:"quality_score,omitempty"`
	SpeedScore         *int   `json:"speed_score,omitempty"`
	InitiativeScore    *int   `json:"initiative_score,omitempty"`
	TeamworkScore      *int   `json:"teamwork_score,omitempty"`
	CommunicationScore *int   `json:"communication_score,omitempty"`
	Strengths          *string `json:"strengths,omitempty"`
	Improvements       *string `json:"improvements,omitempty"`
	Comments           *string `json:"comments,omitempty"`
	Aspect             *string `json:"aspect,omitempty"`
	Notes              *string `json:"notes,omitempty"`
	AssessmentDate     *string `json:"assessment_date,omitempty"` // YYYY-MM-DD
}

func NewAssessmentHandler(db *sql.DB) *AssessmentHandler {
	return &AssessmentHandler{db: db}
}

func (h *AssessmentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
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
		limit = 15
	}
	offset := (page - 1) * limit

	internFilter := strings.TrimSpace(r.URL.Query().Get("intern_id"))
	where := []string{}
	args := []interface{}{}

	role := normalizeRole(claims.Role)
	if role == "intern" {
		var internID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&internID); err != nil {
			utils.RespondNotFound(w, "Intern not found")
			return
		}
		where = append(where, "a.intern_id = ?")
		args = append(args, internID)
	} else if internFilter != "" {
		if id, err := strconv.ParseInt(internFilter, 10, 64); err == nil {
			where = append(where, "a.intern_id = ?")
			args = append(args, id)
		}
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	baseFrom := `
		FROM assessments a
		LEFT JOIN interns i ON a.intern_id = i.id
		LEFT JOIN users iu ON i.user_id = iu.id
		LEFT JOIN users au ON a.assessed_by = au.id
		LEFT JOIN tasks t ON a.task_id = t.id
	`

	var total int64
	if err := h.db.QueryRow("SELECT COUNT(*) "+baseFrom+" "+whereClause, args...).Scan(&total); err != nil {
		utils.RespondInternalError(w, "Failed to count assessments")
		return
	}

	query := `
		SELECT a.id, a.intern_id, a.task_id, a.assessed_by, a.score, a.category, a.aspect,
		       a.quality_score, a.speed_score, a.initiative_score, a.teamwork_score, a.communication_score,
		       a.strengths, a.improvements, a.comments, a.notes, a.assessment_date, a.created_at, a.updated_at,
		       iu.name, au.name, t.title
	` + baseFrom + " " + whereClause + " ORDER BY a.created_at DESC LIMIT ? OFFSET ?"

	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to fetch assessments")
		return
	}
	defer rows.Close()

	var assessments []models.Assessment
	for rows.Next() {
		var a models.Assessment
		var internName, assessorName, taskTitle sql.NullString
		if err := rows.Scan(
			&a.ID, &a.InternID, &a.TaskID, &a.AssessedBy, &a.Score, &a.Category, &a.Aspect,
			&a.QualityScore, &a.SpeedScore, &a.InitiativeScore, &a.TeamworkScore, &a.CommunicationScore,
			&a.Strengths, &a.Improvements, &a.Comments, &a.Notes, &a.AssessmentDate, &a.CreatedAt, &a.UpdatedAt,
			&internName, &assessorName, &taskTitle,
		); err == nil {
			if internName.Valid {
				a.InternName = internName.String
			}
			if assessorName.Valid {
				a.AssessorName = assessorName.String
			}
			if taskTitle.Valid {
				a.TaskTitle = taskTitle.String
			}
			assessments = append(assessments, a)
		}
	}

	utils.RespondPaginated(w, assessments, utils.CalculatePagination(page, limit, total))
}

func (h *AssessmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	query := `
		SELECT a.id, a.intern_id, a.task_id, a.assessed_by, a.score, a.category, a.aspect,
		       a.quality_score, a.speed_score, a.initiative_score, a.teamwork_score, a.communication_score,
		       a.strengths, a.improvements, a.comments, a.notes, a.assessment_date, a.created_at, a.updated_at,
		       iu.name, au.name, t.title
		FROM assessments a
		LEFT JOIN interns i ON a.intern_id = i.id
		LEFT JOIN users iu ON i.user_id = iu.id
		LEFT JOIN users au ON a.assessed_by = au.id
		LEFT JOIN tasks t ON a.task_id = t.id
		WHERE a.id = ?
	`

	var a models.Assessment
	var internName, assessorName, taskTitle sql.NullString
	err := h.db.QueryRow(query, id).Scan(
		&a.ID, &a.InternID, &a.TaskID, &a.AssessedBy, &a.Score, &a.Category, &a.Aspect,
		&a.QualityScore, &a.SpeedScore, &a.InitiativeScore, &a.TeamworkScore, &a.CommunicationScore,
		&a.Strengths, &a.Improvements, &a.Comments, &a.Notes, &a.AssessmentDate, &a.CreatedAt, &a.UpdatedAt,
		&internName, &assessorName, &taskTitle,
	)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Assessment not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	if normalizeRole(claims.Role) == "intern" {
		var myInternID int64
		if err := h.db.QueryRow("SELECT id FROM interns WHERE user_id = ?", claims.UserID).Scan(&myInternID); err != nil || myInternID != a.InternID {
			utils.RespondForbidden(w, "You do not have access to this assessment")
			return
		}
	}

	if internName.Valid {
		a.InternName = internName.String
	}
	if assessorName.Valid {
		a.AssessorName = assessorName.String
	}
	if taskTitle.Valid {
		a.TaskTitle = taskTitle.String
	}

	utils.RespondSuccess(w, "Assessment retrieved", a)
}

func (h *AssessmentHandler) GetByInternID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	internID, _ := strconv.ParseInt(vars["id"], 10, 64)
	q := r.URL.Query()
	q.Set("intern_id", strconv.FormatInt(internID, 10))
	r.URL.RawQuery = q.Encode()
	h.GetAll(w, r)
}

func (h *AssessmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can create assessments")
		return
	}

	var req models.CreateAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	if req.InternID == 0 {
		utils.RespondBadRequest(w, "intern_id is required")
		return
	}

	score := (req.QualityScore + req.SpeedScore + req.InitiativeScore + req.TeamworkScore + req.CommunicationScore) / 5
	aspect := req.Aspect
	if strings.TrimSpace(aspect) == "" {
		aspect = "overall"
	}
	assessmentDate := req.AssessmentDate
	if assessmentDate.IsZero() {
		assessmentDate = time.Now()
	}

	var taskID sql.NullInt64
	if req.TaskID != nil && *req.TaskID > 0 {
		taskID = sql.NullInt64{Int64: *req.TaskID, Valid: true}
	}

	_, err := h.db.Exec(
		`INSERT INTO assessments (intern_id, task_id, assessed_by, score, aspect, quality_score, speed_score, initiative_score,
		                          teamwork_score, communication_score, strengths, improvements, comments, notes, assessment_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.InternID, taskID, claims.UserID, score, aspect,
		req.QualityScore, req.SpeedScore, req.InitiativeScore, req.TeamworkScore, req.CommunicationScore,
		nullIfEmpty(req.Strengths), nullIfEmpty(req.Improvements), nullIfEmpty(req.Comments), nullIfEmpty(req.Notes), assessmentDate,
	)
	if err != nil {
		utils.RespondInternalError(w, "Failed to create assessment")
		return
	}

	utils.RespondCreated(w, "Assessment created", nil)
}

func (h *AssessmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can update assessments")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	var req updateAssessmentPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	// Load existing scores
	var existing struct {
		Quality       sql.NullInt64
		Speed         sql.NullInt64
		Initiative    sql.NullInt64
		Teamwork      sql.NullInt64
		Communication sql.NullInt64
	}
	err := h.db.QueryRow(
		`SELECT quality_score, speed_score, initiative_score, teamwork_score, communication_score
		 FROM assessments WHERE id = ?`, id,
	).Scan(&existing.Quality, &existing.Speed, &existing.Initiative, &existing.Teamwork, &existing.Communication)
	if err == sql.ErrNoRows {
		utils.RespondNotFound(w, "Assessment not found")
		return
	}
	if err != nil {
		utils.RespondInternalError(w, "Database error")
		return
	}

	quality := int64OrZero(existing.Quality)
	speed := int64OrZero(existing.Speed)
	initiative := int64OrZero(existing.Initiative)
	teamwork := int64OrZero(existing.Teamwork)
	communication := int64OrZero(existing.Communication)

	updates := []string{}
	args := []interface{}{}

	if req.QualityScore != nil {
		quality = int64(*req.QualityScore)
		updates = append(updates, "quality_score = ?")
		args = append(args, *req.QualityScore)
	}
	if req.SpeedScore != nil {
		speed = int64(*req.SpeedScore)
		updates = append(updates, "speed_score = ?")
		args = append(args, *req.SpeedScore)
	}
	if req.InitiativeScore != nil {
		initiative = int64(*req.InitiativeScore)
		updates = append(updates, "initiative_score = ?")
		args = append(args, *req.InitiativeScore)
	}
	if req.TeamworkScore != nil {
		teamwork = int64(*req.TeamworkScore)
		updates = append(updates, "teamwork_score = ?")
		args = append(args, *req.TeamworkScore)
	}
	if req.CommunicationScore != nil {
		communication = int64(*req.CommunicationScore)
		updates = append(updates, "communication_score = ?")
		args = append(args, *req.CommunicationScore)
	}

	if req.Strengths != nil {
		updates = append(updates, "strengths = ?")
		args = append(args, nullIfEmpty(*req.Strengths))
	}
	if req.Improvements != nil {
		updates = append(updates, "improvements = ?")
		args = append(args, nullIfEmpty(*req.Improvements))
	}
	if req.Comments != nil {
		updates = append(updates, "comments = ?")
		args = append(args, nullIfEmpty(*req.Comments))
	}
	if req.Notes != nil {
		updates = append(updates, "notes = ?")
		args = append(args, nullIfEmpty(*req.Notes))
	}
	if req.Aspect != nil {
		updates = append(updates, "aspect = ?")
		args = append(args, *req.Aspect)
	}
	if req.AssessmentDate != nil && *req.AssessmentDate != "" {
		if parsed, err := time.Parse("2006-01-02", *req.AssessmentDate); err == nil {
			updates = append(updates, "assessment_date = ?")
			args = append(args, parsed)
		}
	}

	// recompute score if any criteria updated
	score := int((quality + speed + initiative + teamwork + communication) / 5)
	updates = append(updates, "score = ?")
	args = append(args, score)

	if len(updates) == 0 {
		utils.RespondBadRequest(w, "No updates provided")
		return
	}

	args = append(args, id)
	if _, err := h.db.Exec("UPDATE assessments SET "+strings.Join(updates, ", ")+" WHERE id = ?", args...); err != nil {
		utils.RespondInternalError(w, "Failed to update assessment")
		return
	}

	utils.RespondSuccess(w, "Assessment updated", nil)
}

func (h *AssessmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}
	if normalizeRole(claims.Role) == "intern" {
		utils.RespondForbidden(w, "Only admin or pembimbing can delete assessments")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	if _, err := h.db.Exec("DELETE FROM assessments WHERE id = ?", id); err != nil {
		utils.RespondInternalError(w, "Failed to delete assessment")
		return
	}

	utils.RespondSuccess(w, "Assessment deleted", nil)
}
