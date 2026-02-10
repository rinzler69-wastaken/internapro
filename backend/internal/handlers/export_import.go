package handlers

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dsi_interna_sys/internal/utils"

	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

type ExportImportHandler struct {
	db *sql.DB
}

func NewExportImportHandler(db *sql.DB) *ExportImportHandler {
	return &ExportImportHandler{db: db}
}

func (h *ExportImportHandler) ExportInterns(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	where := ""
	args := []interface{}{}
	if status != "" {
		where = "WHERE i.status = ?"
		args = append(args, status)
	}

	query := `
		SELECT i.id, i.full_name, u.email, i.nis, i.school, i.department, i.phone, i.address,
		       su.name as supervisor_name, i.start_date, i.end_date, i.status,
		       (SELECT COUNT(*) FROM tasks t WHERE t.intern_id = i.id) as total_tasks,
		       (SELECT COUNT(*) FROM tasks t WHERE t.intern_id = i.id AND t.status = 'completed') as completed_tasks,
		       (SELECT COUNT(*) FROM attendances a WHERE a.intern_id = i.id) as total_attendance,
		       (SELECT COUNT(*) FROM attendances a WHERE a.intern_id = i.id AND a.status IN ('present','late')) as present_attendance,
		       (SELECT AVG(score) FROM tasks t WHERE t.intern_id = i.id AND t.status = 'completed' AND t.score IS NOT NULL) as avg_score
		FROM interns i
		JOIN users u ON i.user_id = u.id
		LEFT JOIN users su ON i.supervisor_id = su.id
	` + where + ` ORDER BY i.created_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to export interns")
		return
	}
	defer rows.Close()

	headers := []string{
		"ID", "Nama", "Email", "NIS", "Sekolah", "Jurusan", "No. Telepon", "Alamat", "Pembimbing",
		"Tanggal Mulai", "Tanggal Selesai", "Status", "Total Tugas", "Tugas Selesai", "Tingkat Kehadiran (%)", "Skor Rata-rata",
	}

	var data [][]string
	for rows.Next() {
		var (
			id                          int64
			fullName, email             string
			nis, school, department     sql.NullString
			phone, address              sql.NullString
			supervisorName              sql.NullString
			startDate, endDate          sql.NullTime
			rowStatus                   string
			totalTasks, completedTasks  int64
			totalAttendance, presentAtt int64
			avgScore                    sql.NullFloat64
		)
		if err := rows.Scan(
			&id, &fullName, &email, &nis, &school, &department, &phone, &address,
			&supervisorName, &startDate, &endDate, &rowStatus,
			&totalTasks, &completedTasks, &totalAttendance, &presentAtt, &avgScore,
		); err != nil {
			continue
		}

		attendanceRate := 0.0
		if totalAttendance > 0 {
			attendanceRate = (float64(presentAtt) / float64(totalAttendance)) * 100
		}

		data = append(data, []string{
			strconv.FormatInt(id, 10),
			fullName,
			email,
			valueOrDash(nis),
			valueOrDash(school),
			valueOrDash(department),
			valueOrDash(phone),
			valueOrDash(address),
			valueOrDash(supervisorName),
			formatDate(startDate),
			formatDate(endDate),
			internStatusLabel(rowStatus),
			strconv.FormatInt(totalTasks, 10),
			strconv.FormatInt(completedTasks, 10),
			fmt.Sprintf("%.1f", attendanceRate),
			formatFloat(avgScore),
		})
	}

	filename := fmt.Sprintf("Data_Peserta_Magang_%s.xlsx", time.Now().Format("2006-01-02_150405"))
	if err := writeExcel(w, filename, headers, data); err != nil {
		utils.RespondInternalError(w, "Failed to generate export")
		return
	}
}

func (h *ExportImportHandler) ExportAttendances(w http.ResponseWriter, r *http.Request) {
	internID := strings.TrimSpace(r.URL.Query().Get("intern_id"))
	startDate := strings.TrimSpace(r.URL.Query().Get("start_date"))
	endDate := strings.TrimSpace(r.URL.Query().Get("end_date"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	where := []string{}
	args := []interface{}{}
	if internID != "" {
		where = append(where, "a.intern_id = ?")
		args = append(args, internID)
	}
	if startDate != "" {
		where = append(where, "a.date >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		where = append(where, "a.date <= ?")
		args = append(args, endDate)
	}
	if status != "" {
		where = append(where, "a.status = ?")
		args = append(args, status)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	query := `
		SELECT a.id, i.full_name, i.nis, a.date, a.check_in_time, a.check_out_time,
		       a.status, a.late_reason, a.notes, a.distance_meters
		FROM attendances a
		JOIN interns i ON a.intern_id = i.id
	` + whereClause + ` ORDER BY a.date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to export attendances")
		return
	}
	defer rows.Close()

	headers := []string{
		"ID", "Nama Siswa", "NIS", "Tanggal", "Hari", "Jam Masuk", "Jam Pulang", "Status", "Alasan Terlambat", "Catatan", "Jarak (meter)",
	}

	var data [][]string
	for rows.Next() {
		var (
			id         int64
			name       string
			nis        sql.NullString
			date       time.Time
			checkIn    sql.NullTime
			checkOut   sql.NullTime
			rowStatus  string
			lateReason sql.NullString
			notes      sql.NullString
			distance   sql.NullInt64
		)
		if err := rows.Scan(&id, &name, &nis, &date, &checkIn, &checkOut, &rowStatus, &lateReason, &notes, &distance); err != nil {
			continue
		}

		data = append(data, []string{
			strconv.FormatInt(id, 10),
			name,
			valueOrDash(nis),
			date.Format("02/01/2006"),
			dayNameID(date.Weekday()),
			formatTime(checkIn),
			formatTime(checkOut),
			attendanceStatusLabel(rowStatus),
			valueOrDash(lateReason),
			valueOrDash(notes),
			formatInt64(distance),
		})
	}

	filename := fmt.Sprintf("Data_Presensi_%s.xlsx", time.Now().Format("2006-01-02_150405"))
	if err := writeExcel(w, filename, headers, data); err != nil {
		utils.RespondInternalError(w, "Failed to generate export")
		return
	}
}

func (h *ExportImportHandler) ExportTasks(w http.ResponseWriter, r *http.Request) {
	internID := strings.TrimSpace(r.URL.Query().Get("intern_id"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	where := []string{}
	args := []interface{}{}
	if internID != "" {
		where = append(where, "t.intern_id = ?")
		args = append(args, internID)
	}
	if status != "" {
		where = append(where, "t.status = ?")
		args = append(args, status)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	query := `
		SELECT t.id, t.title, t.description, i.full_name, t.priority, t.status, t.deadline,
		       t.submitted_at, t.completed_at, t.is_late, t.score, t.admin_feedback, u.name, t.created_at
		FROM tasks t
		LEFT JOIN interns i ON t.intern_id = i.id
		LEFT JOIN users u ON t.assigned_by = u.id
	` + whereClause + ` ORDER BY t.created_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.RespondInternalError(w, "Failed to export tasks")
		return
	}
	defer rows.Close()

	headers := []string{
		"ID", "Judul", "Deskripsi", "Nama Siswa", "Prioritas", "Status", "Deadline", "Tanggal Submit",
		"Tanggal Selesai", "Tepat Waktu", "Nilai", "Feedback", "Dibuat Oleh", "Dibuat Pada",
	}

	var data [][]string
	for rows.Next() {
		var (
			id          int64
			title       string
			description sql.NullString
			internName  sql.NullString
			priority    string
			rowStatus   string
			deadline    sql.NullTime
			submittedAt sql.NullTime
			completedAt sql.NullTime
			isLate      bool
			score       sql.NullInt64
			feedback    sql.NullString
			assignedBy  sql.NullString
			createdAt   time.Time
		)
		if err := rows.Scan(&id, &title, &description, &internName, &priority, &rowStatus, &deadline, &submittedAt, &completedAt, &isLate, &score, &feedback, &assignedBy, &createdAt); err != nil {
			continue
		}

		ontime := "-"
		if rowStatus == "completed" {
			if isLate {
				ontime = "Terlambat"
			} else {
				ontime = "Tepat Waktu"
			}
		}

		data = append(data, []string{
			strconv.FormatInt(id, 10),
			title,
			valueOrDash(description),
			valueOrDash(internName),
			priorityLabel(priority),
			taskStatusLabel(rowStatus),
			formatDate(deadline),
			formatDateTime(submittedAt),
			formatDateTime(completedAt),
			ontime,
			formatInt64(score),
			valueOrDash(feedback),
			valueOrDash(assignedBy),
			createdAt.Format("02/01/2006 15:04"),
		})
	}

	filename := fmt.Sprintf("Data_Tugas_%s.xlsx", time.Now().Format("2006-01-02_150405"))
	if err := writeExcel(w, filename, headers, data); err != nil {
		utils.RespondInternalError(w, "Failed to generate export")
		return
	}
}

func (h *ExportImportHandler) DownloadTemplate(w http.ResponseWriter, r *http.Request) {
	headers := []string{"ID", "Nama", "Email", "NIS", "Sekolah", "Jurusan", "No. Telepon", "Alamat", "Pembimbing", "Tanggal Mulai", "Tanggal Selesai", "Status"}
	example := []string{"1", "Sabil Murti", "isabilmurti@gmail.com", "1234", "SMK N 9 Semarang", "PPLG", "0882003427575", "Jl. Bukit Cemara Permai IV, No. DN-28, Meteseh, Tembalang.", "Budi Santoso", "05/01/2026", "05/04/2026", "Aktif"}

	filename := "Template_Import_Peserta_Magang.xlsx"
	if err := writeExcel(w, filename, headers, [][]string{example}); err != nil {
		utils.RespondInternalError(w, "Failed to generate template")
		return
	}
}

func (h *ExportImportHandler) ImportInterns(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondBadRequest(w, "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.RespondBadRequest(w, "Missing file")
		return
	}
	defer file.Close()

	supervisorID := int64(0)
	if raw := r.FormValue("supervisor_id"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			supervisorID = parsed
		}
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	rows, err := readRows(file, ext)
	if err != nil {
		utils.RespondBadRequest(w, "Failed to read file")
		return
	}
	if len(rows) == 0 {
		utils.RespondBadRequest(w, "File is empty")
		return
	}

	startIdx := 0
	if isHeaderRow(rows[0]) {
		startIdx = 1
	}

	imported := 0
	skipped := 0
	errors := []string{}

	for i := startIdx; i < len(rows); i++ {
		row := rows[i]
		name := cell(row, 1)
		email := cell(row, 2)
		if name == "" || email == "" {
			skipped++
			continue
		}

		var exists int
		_ = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
		if exists > 0 {
			skipped++
			continue
		}

		resolvedSupervisorID := supervisorID
		if resolvedSupervisorID == 0 {
			supervisorName := cell(row, 8)
			if supervisorName != "" {
				_ = h.db.QueryRow(
					"SELECT id FROM users WHERE name LIKE ? AND role IN ('admin','pembimbing','supervisor') LIMIT 1",
					"%"+supervisorName+"%",
				).Scan(&resolvedSupervisorID)
			}
		}

		status := mapInternStatus(cell(row, 11))
		startDate, okStart := parseFlexibleDate(cell(row, 9))
		endDate, okEnd := parseFlexibleDate(cell(row, 10))
		if !okStart || !okEnd {
			errors = append(errors, fmt.Sprintf("Row %d: invalid start/end date", i+1))
			skipped++
			continue
		}

		password := "password123"
		if pwd := cell(row, 12); pwd != "" {
			password = pwd
		}

		hashed, err := hashPassword(password)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: failed to hash password", i+1))
			skipped++
			continue
		}

		tx, err := h.db.Begin()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: failed to start transaction", i+1))
			skipped++
			continue
		}

		res, err := tx.Exec(
			"INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, 'intern')",
			name, email, hashed,
		)
		if err != nil {
			tx.Rollback()
			errors = append(errors, fmt.Sprintf("Row %d: failed to create user", i+1))
			skipped++
			continue
		}
		userID, _ := res.LastInsertId()

		_, err = tx.Exec(
			`INSERT INTO interns (user_id, supervisor_id, full_name, nis, school, department, phone, address, start_date, end_date, status)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			userID,
			nullInt64(resolvedSupervisorID),
			name,
			cell(row, 3),
			cell(row, 4),
			cell(row, 5),
			cell(row, 6),
			cell(row, 7),
			startDate,
			endDate,
			status,
		)
		if err != nil {
			tx.Rollback()
			errors = append(errors, fmt.Sprintf("Row %d: failed to create intern", i+1))
			skipped++
			continue
		}

		if err := tx.Commit(); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: failed to commit", i+1))
			skipped++
			continue
		}

		imported++
	}

	payload := map[string]interface{}{
		"imported": imported,
		"skipped":  skipped,
	}
	if len(errors) > 0 {
		payload["errors"] = errors
	}

	utils.RespondSuccess(w, "Import completed", payload)
}

// Helpers

func writeExcel(w http.ResponseWriter, filename string, headers []string, rows [][]string) error {
	file := excelize.NewFile()
	sheet := file.GetSheetName(0)

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		file.SetCellValue(sheet, cell, header)
	}

	for rIdx, row := range rows {
		for cIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, rIdx+2)
			file.SetCellValue(sheet, cell, value)
		}
	}

	style, _ := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"4B5563"}},
	})
	endCell, _ := excelize.CoordinatesToCellName(len(headers), 1)
	file.SetCellStyle(sheet, "A1", endCell, style)

	buf, err := file.WriteToBuffer()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	_, err = w.Write(buf.Bytes())
	return err
}

func readRows(file multipart.File, ext string) ([][]string, error) {
	switch ext {
	case ".csv":
		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1
		return reader.ReadAll()
	case ".xlsx", ".xls":
		f, err := excelize.OpenReader(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		sheet := f.GetSheetName(0)
		return f.GetRows(sheet)
	default:
		return nil, fmt.Errorf("unsupported file type")
	}
}

func isHeaderRow(row []string) bool {
	if len(row) == 0 {
		return false
	}
	joined := strings.ToLower(strings.Join(row, " "))
	return strings.Contains(joined, "nama") && strings.Contains(joined, "email")
}

func cell(row []string, idx int) string {
	if idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func parseFlexibleDate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}

	if num, err := strconv.ParseFloat(value, 64); err == nil {
		if t, err := excelize.ExcelDateToTime(num, false); err == nil {
			return t, true
		}
	}

	value = strings.ReplaceAll(value, "-", "/")
	layouts := []string{"02/01/2006", "2/1/2006", "2006/01/02", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, true
		}
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func internStatusLabel(status string) string {
	switch status {
	case "active":
		return "Aktif"
	case "completed":
		return "Selesai"
	case "cancelled":
		return "Dibatalkan"
	case "pending":
		return "Pending"
	default:
		return status
	}
}

func taskStatusLabel(status string) string {
	switch status {
	case "pending":
		return "Menunggu"
	case "in_progress":
		return "Dalam Proses"
	case "submitted":
		return "Disubmit"
	case "revision":
		return "Revisi"
	case "completed":
		return "Selesai"
	default:
		return status
	}
}

func priorityLabel(priority string) string {
	switch priority {
	case "low":
		return "Rendah"
	case "medium":
		return "Sedang"
	case "high":
		return "Tinggi"
	case "urgent":
		return "Mendesak"
	default:
		if priority == "" {
			return "Sedang"
		}
		return priority
	}
}

func attendanceStatusLabel(status string) string {
	switch status {
	case "present":
		return "Hadir"
	case "late":
		return "Terlambat"
	case "absent":
		return "Tidak Hadir"
	case "sick":
		return "Sakit"
	case "permission":
		return "Izin"
	default:
		return status
	}
}

func mapInternStatus(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	switch v {
	case "aktif", "active":
		return "active"
	case "selesai", "completed":
		return "completed"
	case "pending":
		return "pending"
	case "dibatalkan", "cancelled":
		return "cancelled"
	default:
		return "active"
	}
}

func valueOrDash(v sql.NullString) string {
	if v.Valid && strings.TrimSpace(v.String) != "" {
		return v.String
	}
	return "-"
}

func formatDate(v sql.NullTime) string {
	if v.Valid {
		return v.Time.Format("02/01/2006")
	}
	return "-"
}

func formatDateTime(v sql.NullTime) string {
	if v.Valid {
		return v.Time.Format("02/01/2006 15:04")
	}
	return "-"
}

func formatTime(v sql.NullTime) string {
	if v.Valid {
		return v.Time.Format("15:04")
	}
	return "-"
}

func formatInt64(v sql.NullInt64) string {
	if v.Valid {
		return strconv.FormatInt(v.Int64, 10)
	}
	return "-"
}

func formatFloat(v sql.NullFloat64) string {
	if v.Valid {
		return fmt.Sprintf("%.1f", v.Float64)
	}
	return "0"
}

func dayNameID(day time.Weekday) string {
	switch day {
	case time.Sunday:
		return "Minggu"
	case time.Monday:
		return "Senin"
	case time.Tuesday:
		return "Selasa"
	case time.Wednesday:
		return "Rabu"
	case time.Thursday:
		return "Kamis"
	case time.Friday:
		return "Jumat"
	case time.Saturday:
		return "Sabtu"
	default:
		return "-"
	}
}

// HashPassword wraps bcrypt in utils for import usage
// We keep a small helper to safely convert optional int64
func nullInt64(val int64) sql.NullInt64 {
	if val == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: val, Valid: true}
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
