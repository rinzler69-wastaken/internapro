package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Struct untuk payload pendaftaran
type RegisterInternRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	School          string `json:"school"`
	Department      string `json:"department"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	InstitutionID   *int64 `json:"institution_id,omitempty"`
	NIS             string `json:"nis,omitempty"`
	Phone           string `json:"phone,omitempty"`
	Address         string `json:"address,omitempty"`
	StartDate       string `json:"start_date,omitempty"`
	EndDate         string `json:"end_date,omitempty"`
	SupervisorID    *int64 `json:"supervisor_id,omitempty"`
}

// Register menangani pendaftaran magang mandiri (Public Endpoint)
func (h *InternHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterInternRequest

	// 1. Decode JSON Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 2. Validasi Input
	if req.Name == "" || req.Email == "" || req.Password == "" || req.School == "" || req.Department == "" {
		http.Error(w, `{"message": "Semua kolom wajib diisi"}`, http.StatusBadRequest)
		return
	}

	if req.Password != req.ConfirmPassword {
		http.Error(w, `{"message": "Konfirmasi password tidak cocok"}`, http.StatusBadRequest)
		return
	}

	// 3. Cek apakah email sudah terdaftar dan siapkan upsert
	var (
		existingID    int64
		existingRole  string
		existingHash  sql.NullString
		existingProv  sql.NullString
		userExistsErr error
	)
	userExistsErr = h.db.QueryRow("SELECT id, role, password_hash, provider FROM users WHERE email = ? LIMIT 1", req.Email).
		Scan(&existingID, &existingRole, &existingHash, &existingProv)

	emailExists := userExistsErr == nil
	if userExistsErr != nil && userExistsErr != sql.ErrNoRows {
		fmt.Println("ERROR Check Email:", userExistsErr)
		http.Error(w, `{"message": "Server error checking email"}`, http.StatusInternalServerError)
		return
	}

	if emailExists && strings.ToLower(existingRole) != "intern" {
		http.Error(w, `{"message": "Email sudah terdaftar"}`, http.StatusConflict)
		return
	}

	// 4. Cari / buat institusi berdasarkan nama sekolah (agar institution_id tidak NULL)
	schoolName := strings.TrimSpace(req.School)
	var institutionID sql.NullInt64
	var err error
	if schoolName != "" {
		var instID int64
		err = h.db.QueryRow("SELECT id FROM institutions WHERE name = ? LIMIT 1", schoolName).Scan(&instID)
		if err == sql.ErrNoRows {
			resInst, errInst := h.db.Exec("INSERT INTO institutions (name, created_at, updated_at) VALUES (?, NOW(), NOW())", schoolName)
			if errInst != nil {
				http.Error(w, `{"message": "Gagal membuat institusi"}`, http.StatusInternalServerError)
				return
			}
			instID, _ = resInst.LastInsertId()
		} else if err != nil {
			http.Error(w, `{"message": "Gagal mencari institusi"}`, http.StatusInternalServerError)
			return
		}
		institutionID = sql.NullInt64{Int64: instID, Valid: true}
	}

	// 5. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"message": "Gagal memproses password"}`, http.StatusInternalServerError)
		return
	}

	// =================================================================================
	// INSERT / UPDATE USERS
	// =================================================================================
	var userID int64
	if !emailExists {
		res, err := h.db.Exec(
			"INSERT INTO users (name, email, password_hash, role, created_at, provider) VALUES (?, ?, ?, 'intern', NOW(), 'google')",
			req.Name, req.Email, string(hashedPassword),
		)
		if err != nil {
			fmt.Println("ERROR SQL INSERT USER:", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Gagal mendaftarkan user: %v", err),
			})
			return
		}
		userID, _ = res.LastInsertId()
	} else {
		userID = existingID
		// Jika belum punya password (ex: OAuth), set password baru
		if !existingHash.Valid || existingHash.String == "" {
			if _, err := h.db.Exec("UPDATE users SET password_hash = ?, name = ? WHERE id = ?", string(hashedPassword), req.Name, userID); err != nil {
				http.Error(w, `{"message": "Gagal memperbarui password"}`, http.StatusInternalServerError)
				return
			}
		} else {
			// Tetap perbarui nama jika kosong
			if _, err := h.db.Exec("UPDATE users SET name = ? WHERE id = ?", req.Name, userID); err != nil {
				http.Error(w, `{"message": "Gagal memperbarui profil"}`, http.StatusInternalServerError)
				return
			}
		}
	}

	// =================================================================================
	// FIX 2: Insert ke tabel 'interns'
	// - Wajib isi 'full_name' (ambil dari req.Name)
	// - Wajib isi 'start_date' & 'end_date' (NOT NULL). Kita isi default tanggal hari ini dulu.
	//   Nanti admin yang update tanggal aslinya saat approval.
	// - Isi 'status' = 'pending'
	// - Set institution_id berdasarkan nama sekolah yang dikirim user
	// =================================================================================

	// Default date (hari ini) agar tidak error constraint
	defaultDate := time.Now().Format("2006-01-02")
	startDate := defaultDate
	endDate := defaultDate
	if strings.TrimSpace(req.StartDate) != "" {
		startDate = req.StartDate
	}
	if strings.TrimSpace(req.EndDate) != "" {
		endDate = req.EndDate
	}

	var institutionValue interface{}
	if institutionID.Valid {
		institutionValue = institutionID.Int64
	} else {
		institutionValue = nil
	}

	// Insert or update intern profile
	var internExists bool
	if err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM interns WHERE user_id = ?)", userID).Scan(&internExists); err != nil {
		http.Error(w, `{"message": "Gagal memeriksa profil magang"}`, http.StatusInternalServerError)
		return
	}

	if !internExists {
		_, err = h.db.Exec(
			`INSERT INTO interns (user_id, institution_id, supervisor_id, full_name, nis, phone, address, school, department, status, start_date, end_date, created_at) 
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?, NOW())`,
			userID, institutionValue, req.SupervisorID, req.Name, req.NIS, req.Phone, req.Address, req.School, req.Department, startDate, endDate,
		)
		if err != nil {
			fmt.Println("ERROR SQL INSERT INTERN:", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Gagal menyimpan data magang: %v", err),
			})
			return
		}
	} else {
		// Update pending profile fields, keep status pending
		_, err = h.db.Exec(
			`UPDATE interns 
			 SET institution_id = COALESCE(?, institution_id),
			     supervisor_id = COALESCE(?, supervisor_id),
			     full_name = ?, nis = ?, phone = ?, address = ?, school = ?, department = ?, start_date = ?, end_date = ?, updated_at = NOW()
			 WHERE user_id = ? AND status = 'pending'`,
			institutionValue, req.SupervisorID, req.Name, req.NIS, req.Phone, req.Address, req.School, req.Department, startDate, endDate, userID,
		)
		if err != nil {
			http.Error(w, `{"message": "Gagal memperbarui data magang"}`, http.StatusInternalServerError)
			return
		}
	}

	// 6. Response Sukses
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Pendaftaran berhasil. Silakan tunggu persetujuan admin.",
		"user_id": userID,
	})
}
