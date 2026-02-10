package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dsi_interna_sys/internal/middleware"
	"dsi_interna_sys/internal/models"
	"dsi_interna_sys/internal/utils"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type AssessmentHandler struct {
	db *sql.DB
}

type updateAssessmentPayload struct {
	QualityScore       *int    `json:"quality_score,omitempty"`
	SpeedScore         *int    `json:"speed_score,omitempty"`
	InitiativeScore    *int    `json:"initiative_score,omitempty"`
	TeamworkScore      *int    `json:"teamwork_score,omitempty"`
	CommunicationScore *int    `json:"communication_score,omitempty"`
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

	var assessments []map[string]interface{}
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
			assessments = append(assessments, presentAssessment(a))
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

	utils.RespondSuccess(w, "Assessment retrieved", presentAssessment(a))
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

	// Validate intern existence (and supervisor ownership)
	var supervisorID sql.NullInt64
	if err := h.db.QueryRow("SELECT supervisor_id FROM interns WHERE id = ?", req.InternID).Scan(&supervisorID); err != nil {
		if err == sql.ErrNoRows {
			utils.RespondBadRequest(w, "intern_id not found")
			return
		}
		utils.RespondInternalError(w, "Failed to validate intern")
		return
	}
	if normalizeRole(claims.Role) == "pembimbing" && supervisorID.Valid && supervisorID.Int64 != claims.UserID {
		utils.RespondForbidden(w, "You cannot assess an intern outside your supervision")
		return
	}

	// Resolve assessor ID to satisfy potential FK to supervisors table
	assessorCandidates := []int64{claims.UserID}
	if supID, ok, err := h.getSupervisorIDForUser(claims.UserID); err == nil && ok {
		assessorCandidates = append([]int64{supID}, assessorCandidates...)
	} else if err != nil {
		utils.RespondInternalError(w, "Failed to validate assessor")
		return
	}
	// Basic score validation to avoid DB constraint errors
	scores := []struct {
		name string
		val  int
	}{
		{"quality_score", req.QualityScore},
		{"speed_score", req.SpeedScore},
		{"initiative_score", req.InitiativeScore},
		{"teamwork_score", req.TeamworkScore},
		{"communication_score", req.CommunicationScore},
	}
	for _, s := range scores {
		if s.val < 0 || s.val > 100 {
			utils.RespondBadRequest(w, s.name+" must be between 0 and 100")
			return
		}
	}

	score := (req.QualityScore + req.SpeedScore + req.InitiativeScore + req.TeamworkScore + req.CommunicationScore) / 5
	aspect := req.Aspect
	if strings.TrimSpace(aspect) == "" {
		aspect = "overall"
	}
	assessmentDate := time.Now()
	if strings.TrimSpace(req.AssessmentDate) != "" {
		parsed, err := time.Parse("2006-01-02", req.AssessmentDate)
		if err != nil {
			utils.RespondBadRequest(w, "assessment_date must be in YYYY-MM-DD format")
			return
		}
		assessmentDate = parsed
	}

	var taskID sql.NullInt64
	if req.TaskID != nil && *req.TaskID > 0 {
		taskID = sql.NullInt64{Int64: *req.TaskID, Valid: true}
		var taskInternID sql.NullInt64
		err := h.db.QueryRow("SELECT intern_id FROM tasks WHERE id = ?", *req.TaskID).Scan(&taskInternID)
		if err == sql.ErrNoRows {
			utils.RespondBadRequest(w, "task_id not found")
			return
		}
		if err != nil {
			utils.RespondInternalError(w, "Failed to validate task")
			return
		}
		if taskInternID.Valid && taskInternID.Int64 != req.InternID {
			utils.RespondBadRequest(w, "task_id does not belong to the selected intern")
			return
		}
	}

	var lastErr error
	for idx, assessorID := range assessorCandidates {
		_, err := h.db.Exec(
			`INSERT INTO assessments (intern_id, task_id, assessed_by, score, aspect, quality_score, speed_score, initiative_score,
			                          teamwork_score, communication_score, strengths, improvements, comments, notes, assessment_date)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			req.InternID, taskID, assessorID, score, aspect,
			req.QualityScore, req.SpeedScore, req.InitiativeScore, req.TeamworkScore, req.CommunicationScore,
			nullIfEmpty(req.Strengths), nullIfEmpty(req.Improvements), nullIfEmpty(req.Comments), nullIfEmpty(req.Notes), assessmentDate,
		)
		if err == nil {
			utils.RespondCreated(w, "Assessment created", nil)
			return
		}
		lastErr = err
		// If FK fails, try next candidate
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1452 && idx < len(assessorCandidates)-1 {
			continue
		}
		break
	}

	log.Printf("create assessment failed: %v", lastErr)
	if mysqlErr, ok := lastErr.(*mysql.MySQLError); ok {
		switch mysqlErr.Number {
		case 1452: // foreign key constraint
			utils.RespondBadRequest(w, "Assessor must be a registered supervisor for this database schema")
			return
		case 3819, 4025: // check constraint
			utils.RespondBadRequest(w, "scores must be between 0 and 100")
			return
		case 1292: // incorrect date value
			utils.RespondBadRequest(w, "assessment_date is invalid")
			return
		}
	}
	utils.RespondInternalError(w, "Failed to create assessment")
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

// getSupervisorIDForUser returns supervisor.id for the given user if it exists.
func (h *AssessmentHandler) getSupervisorIDForUser(userID int64) (int64, bool, error) {
	var supID int64
	err := h.db.QueryRow("SELECT id FROM supervisors WHERE user_id = ? LIMIT 1", userID).Scan(&supID)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return supID, true, nil
}

func presentAssessment(a models.Assessment) map[string]interface{} {
	return map[string]interface{}{
		"id":                  a.ID,
		"intern_id":           a.InternID,
		"task_id":             nullIntToPtr(a.TaskID),
		"assessed_by":         a.AssessedBy,
		"score":               a.Score,
		"category":            a.Category,
		"aspect":              a.Aspect,
		"quality_score":       nullIntToPtr(a.QualityScore),
		"speed_score":         nullIntToPtr(a.SpeedScore),
		"initiative_score":    nullIntToPtr(a.InitiativeScore),
		"teamwork_score":      nullIntToPtr(a.TeamworkScore),
		"communication_score": nullIntToPtr(a.CommunicationScore),
		"strengths":           nullStringToPtr(a.Strengths),
		"improvements":        nullStringToPtr(a.Improvements),
		"comments":            nullStringToPtr(a.Comments),
		"notes":               nullStringToPtr(a.Notes),
		"assessment_date":     a.AssessmentDate.Format("2006-01-02"),
		"created_at":          a.CreatedAt,
		"updated_at":          a.UpdatedAt,
		"intern_name":         a.InternName,
		"assessor_name":       a.AssessorName,
		"task_title":          a.TaskTitle,
	}
}

func nullIntToPtr(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int64)
	return &i
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
