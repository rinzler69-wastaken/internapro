package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type supervisorSeed struct {
	Name  string
	Email string
	NIP   string
}

type internSeed struct {
	ID     int64
	UserID int64
	Status string
}

type taskSeed struct {
	ID       int64
	InternID int64
	Status   string
}

type seedCounts struct {
	Interns                int
	TaskAssignments        int
	AttendanceDays         int
	ReportsMinWeeks        int
	ReportsMaxWeeks        int
	NotificationsPerUser   int
	AssessmentsSampleRatio int
}

var tasksHasTargetDate *bool

var firstNames = []string{
	"Ahmad", "Budi", "Citra", "Dewi", "Eka", "Fadli", "Gita", "Hendra", "Indah", "Joko",
	"Kartika", "Lukman", "Maya", "Naufal", "Olivia", "Prasetyo", "Qori", "Rizki", "Sari", "Taufik",
	"Umi", "Vino", "Wulan", "Yoga", "Zahra", "Aditya", "Bella", "Cahyo", "Diana", "Erwin",
	"Fitri", "Galih", "Hani", "Irfan", "Julia",
}

var lastNames = []string{
	"Pratama", "Wijaya", "Kusuma", "Sari", "Nugroho", "Permana", "Santoso", "Putra", "Wati", "Hidayat",
	"Ramadhan", "Lestari", "Setiawan", "Utami", "Saputra", "Dewi", "Kurniawan", "Putri", "Firmansyah", "Handayani",
	"Ramadhani", "Anggraini", "Prasetya", "Maharani", "Arifin", "Susanti", "Wahyudi", "Puspita", "Haryanto", "Safitri",
}

var schools = []string{
	"SMK Negeri 1 Jakarta",
	"SMK Negeri 2 Bandung",
	"SMK Telkom Malang",
	"SMK Informatika Surabaya",
	"Politeknik Negeri Jakarta",
	"Universitas Indonesia",
	"Institut Teknologi Bandung",
	"Universitas Gadjah Mada",
	"Universitas Brawijaya",
	"Politeknik Elektronika Negeri Surabaya",
	"SMK Negeri 4 Malang",
	"SMK Prakarya Internasional",
	"SMK Wikrama Bogor",
	"Universitas Bina Nusantara",
	"Universitas Telkom",
}

var departments = []string{
	"Rekayasa Perangkat Lunak",
	"Teknik Komputer dan Jaringan",
	"Multimedia",
	"Sistem Informasi",
	"Teknik Informatika",
	"Manajemen Informatika",
	"Desain Grafis",
	"Animasi",
	"Broadcasting",
	"Bisnis Digital",
}

var taskTitles = []string{
	"Membuat Landing Page Website",
	"Develop REST API Authentication",
	"Redesign UI Dashboard Admin",
	"Setup CI/CD Pipeline",
	"Database Migration & Optimization",
	"Implementasi Payment Gateway",
	"Unit Testing Module User",
	"Dokumentasi API Swagger",
	"Mobile App - Login Screen",
	"Integrasi Social Media Login",
	"Develop Chat Feature Real-time",
	"Setup Monitoring & Logging",
	"Optimasi Performance Website",
	"Membuat Report Generator PDF",
	"Implementasi Notifikasi Push",
	"Develop E-commerce Cart System",
	"Setup Email Template",
	"Membuat Data Visualization Dashboard",
	"Develop File Upload System",
	"Implementasi Role-based Access Control",
}

var taskDescriptions = []string{
	"Membuat landing page responsive dengan design modern menggunakan Tailwind CSS dan animasi smooth scroll.",
	"Mengembangkan REST API untuk autentikasi menggunakan JWT dengan fitur login, register, dan refresh token.",
	"Melakukan redesign pada halaman dashboard admin untuk meningkatkan user experience dan accessibility.",
	"Melakukan setup continuous integration dan continuous deployment menggunakan GitHub Actions.",
	"Melakukan migrasi database dan optimasi query untuk meningkatkan performa aplikasi.",
	"Mengintegrasikan payment gateway Midtrans untuk proses pembayaran online.",
	"Menulis unit test untuk module user dengan coverage minimal 80%.",
	"Membuat dokumentasi API lengkap menggunakan Swagger/OpenAPI specification.",
	"Develop tampilan login screen untuk aplikasi mobile dengan Flutter.",
	"Integrasi login menggunakan Google dan Facebook OAuth.",
	"Mengembangkan fitur chat real-time menggunakan WebSocket.",
	"Setup monitoring aplikasi menggunakan Prometheus dan Grafana.",
	"Melakukan optimasi performa website termasuk lazy loading dan caching.",
	"Membuat sistem generate report dalam format PDF yang bisa di-download.",
	"Implementasi push notification untuk web dan mobile application.",
	"Develop sistem keranjang belanja lengkap dengan kalkulasi harga dan diskon.",
	"Membuat email template responsive untuk berbagai keperluan notifikasi.",
	"Membuat dashboard visualisasi data menggunakan Chart.js atau D3.js.",
	"Develop sistem upload file dengan validasi tipe dan ukuran file.",
	"Implementasi sistem role dan permission untuk kontrol akses user.",
}

var supervisorSeeds = []supervisorSeed{
	{Name: "Bambang Agus Herlambang, M.Kom", Email: "bambang.herlambang@internapro.id", NIP: "198501234567891001"},
	{Name: "Noora Qotrun Nada, S.T., M.T.", Email: "noora@internapro.id", NIP: "198601234567891002"},
	{Name: "Mega Novita, Ph.D", Email: "mega.novita@internapro.id", NIP: "198701234567891003"},
}

func main() {
	godotenv.Load()
	rand.Seed(time.Now().UnixNano())

	db := openDB()
	defer db.Close()

	counts := seedCounts{
		Interns:                envInt("SEED_INTERNS", 60),
		TaskAssignments:        envInt("SEED_ASSIGNMENTS", 12),
		AttendanceDays:         envInt("SEED_ATTENDANCE_DAYS", 30),
		ReportsMinWeeks:        envInt("SEED_REPORT_WEEKS_MIN", 2),
		ReportsMaxWeeks:        envInt("SEED_REPORT_WEEKS_MAX", 4),
		NotificationsPerUser:   envInt("SEED_NOTIFICATIONS_PER_USER", 2),
		AssessmentsSampleRatio: envInt("SEED_ASSESSMENT_SAMPLE_RATIO", 3),
	}

	force := envBool("SEED_FORCE", false)
	reuseInterns := envBool("SEED_REUSE_INTERNS", false)

	fmt.Println("✅ Connected to database")
	fmt.Println("")

	adminID := seedAdmin(db)
	supervisorIDs := seedSupervisors(db)
	institutionIDs := seedInstitutions(db)
	interns := seedInterns(db, supervisorIDs, institutionIDs, counts.Interns, reuseInterns)

	shouldSeedHeavy := force || !hasRows(db, "tasks")
	if !shouldSeedHeavy {
		fmt.Println("ℹ️  Tasks already exist. Skipping heavy demo data (set SEED_FORCE=true to reseed).")
		printSummary()
		return
	}

	tasks := seedTaskAssignments(db, adminID, interns, counts.TaskAssignments)
	seedAttendances(db, interns, counts.AttendanceDays)
	seedAssessments(db, adminID, tasks, counts.AssessmentsSampleRatio)
	seedReports(db, interns, counts.ReportsMinWeeks, counts.ReportsMaxWeeks)
	seedNotifications(db, adminID, supervisorIDs, interns, counts.NotificationsPerUser)

	printSummary()
}

func openDB() *sql.DB {
	dbUser := env("DB_USER", "root")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := env("DB_HOST", "localhost")
	dbPort := env("DB_PORT", "3306")
	dbName := env("DB_NAME", "interna_db")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	return db
}

func seedAdmin(db *sql.DB) int64 {
	fmt.Println("Creating admin account...")
	adminID, err := createOrGetUser(db, "Administrator", "admin@internapro.id", "password", "admin")
	if err != nil {
		log.Fatal("Failed to create admin:", err)
	}
	fmt.Println("✅ Admin account ready")
	return adminID
}

func seedSupervisors(db *sql.DB) []int64 {
	fmt.Println("Creating supervisor accounts...")
	ids := make([]int64, 0, len(supervisorSeeds))
	for _, sup := range supervisorSeeds {
		userID, err := createOrGetUser(db, sup.Name, sup.Email, "password", "pembimbing")
		if err != nil {
			log.Println("⚠️  Supervisor create failed:", err)
			continue
		}
		ids = append(ids, userID)

		ensureSupervisorProfile(db, userID, sup.Name, sup.NIP)
	}
	fmt.Printf("✅ %d supervisors ready\n", len(ids))
	return ids
}

func seedInstitutions(db *sql.DB) []int64 {
	fmt.Println("Creating institutions...")
	var ids []int64
	for i, name := range schools {
		if i >= 6 {
			break
		}
		address := fmt.Sprintf("Jl. %s No. %d, Jakarta", lastNames[rand.Intn(len(lastNames))], rand.Intn(120)+1)
		phone := fmt.Sprintf("02%d-%07d", rand.Intn(10), rand.Intn(9000000)+1000000)
		email := strings.ToLower(strings.ReplaceAll(name, " ", "")) + "@example.edu"
		id, err := ensureInstitution(db, name, address, phone, email)
		if err == nil {
			ids = append(ids, id)
		}
	}
	fmt.Printf("✅ %d institutions ready\n", len(ids))
	return ids
}

func seedInterns(db *sql.DB, supervisorIDs []int64, institutionIDs []int64, count int, reuseExisting bool) []internSeed {
	interns := make([]internSeed, 0, count)
	usedEmails := map[string]bool{}

	if reuseExisting {
		existing, existingEmails := loadExistingInterns(db, count)
		for email := range existingEmails {
			usedEmails[email] = true
		}
		interns = append(interns, existing...)
		if len(interns) >= count {
			fmt.Printf("✅ %d interns ready (reused)\n", len(interns))
			return interns[:count]
		}
	}

	toCreate := count - len(interns)
	fmt.Printf("Creating %d interns...\n", toCreate)

	for i := 0; i < toCreate; i++ {
		first := firstNames[rand.Intn(len(firstNames))]
		last := lastNames[rand.Intn(len(lastNames))]
		fullName := first + " " + last

		email := uniqueEmail(first, last, usedEmails)
		userID, err := createOrGetUser(db, fullName, email, "password", "intern")
		if err != nil {
			log.Println("⚠️  Intern user create failed:", err)
			continue
		}

		startDate := time.Now().AddDate(0, -rand.Intn(3)-1, -rand.Intn(30))
		endDate := startDate.AddDate(0, rand.Intn(4)+3, 0)
		status := determineInternStatus(endDate)

		institutionID := pickID(institutionIDs)
		supervisorID := pickID(supervisorIDs)

		internID, err := ensureInternProfile(db, userID, institutionID, supervisorID, len(interns)+1, fullName, status, startDate, endDate)
		if err != nil {
			log.Println("⚠️  Intern profile create failed:", err)
			continue
		}

		interns = append(interns, internSeed{ID: internID, UserID: userID, Status: status})
	}

	fmt.Printf("✅ %d interns ready\n", len(interns))
	return interns
}

func seedTaskAssignments(db *sql.DB, adminID int64, interns []internSeed, assignmentCount int) []taskSeed {
	fmt.Printf("Creating %d task assignments...\n", assignmentCount)
	tasks := make([]taskSeed, 0, assignmentCount*5)

	activeInterns := filterActiveInterns(interns)
	allStatuses := []string{"pending", "in_progress", "submitted", "revision", "completed"}

	for i := 0; i < assignmentCount; i++ {
		title := taskTitles[i%len(taskTitles)]
		desc := taskDescriptions[i%len(taskDescriptions)]
		priority := []string{"low", "medium", "high"}[rand.Intn(3)]

		deadline := time.Now().AddDate(0, 0, rand.Intn(28)-7)
		startDate := time.Now().AddDate(0, 0, rand.Intn(10)-3)
		deadlineTime := fmt.Sprintf("%02d:00:00", rand.Intn(5)+14)
		assignToAll := i < 2

		assignmentID, err := insertTaskAssignment(db, adminID, title, desc, priority, startDate, deadline, deadlineTime, assignToAll)
		if err != nil {
			log.Println("⚠️  Task assignment create failed:", err)
			continue
		}

		assigned := pickInterns(activeInterns, assignToAll)
		for _, intern := range assigned {
			_, _ = db.Exec(`
				INSERT IGNORE INTO task_assignment_interns (task_assignment_id, intern_id)
				VALUES (?, ?)
			`, assignmentID, intern.ID)

			status := allStatuses[rand.Intn(len(allStatuses))]
			taskID, err := insertTask(db, assignmentID, intern.ID, adminID, title, desc, priority, status, startDate, deadline, deadlineTime)
			if err != nil {
				log.Println("⚠️  Task create failed:", err)
				continue
			}
			tasks = append(tasks, taskSeed{ID: taskID, InternID: intern.ID, Status: status})
		}
	}

	fmt.Printf("✅ %d task assignments ready\n", assignmentCount)
	return tasks
}

func seedAttendances(db *sql.DB, interns []internSeed, days int) {
	fmt.Printf("Creating attendance records (%d days)...\n", days)
	startDate := time.Now().AddDate(0, 0, -days)
	endDate := time.Now()

	for _, intern := range interns {
		if intern.Status == "cancelled" {
			continue
		}

		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			weekday := d.Weekday()
			if weekday == time.Saturday || weekday == time.Sunday {
				continue
			}

			status, checkIn, checkOut, lateReason, notes := generateAttendanceData(d)
			_, _ = db.Exec(`
				INSERT INTO attendances (intern_id, date, check_in_time, check_out_time, status, late_reason, notes, distance_meters)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE status = VALUES(status), check_in_time = VALUES(check_in_time), check_out_time = VALUES(check_out_time)
			`, intern.ID, d.Format("2006-01-02"), checkIn, checkOut, status, lateReason, notes, rand.Intn(900)+50)
		}
	}

	fmt.Println("✅ Attendance records created")
}

func seedAssessments(db *sql.DB, adminID int64, tasks []taskSeed, sampleRatio int) {
	fmt.Println("Creating assessments...")
	created := 0
	for _, task := range tasks {
		if task.Status != "completed" {
			continue
		}
		if sampleRatio > 1 && rand.Intn(sampleRatio) != 0 {
			continue
		}

		score := rand.Intn(31) + 70
		assessmentDate := time.Now().AddDate(0, 0, -rand.Intn(20))
		aspect := []string{"discipline", "work_quality", "attitude", "communication"}[rand.Intn(4)]

		_, err := db.Exec(`
			INSERT INTO assessments (
				intern_id, task_id, assessed_by, score, aspect,
				quality_score, speed_score, initiative_score, teamwork_score, communication_score,
				strengths, improvements, comments, assessment_date
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			task.InternID, task.ID, adminID, score, aspect,
			rand.Intn(31)+70, rand.Intn(31)+65, rand.Intn(41)+60, rand.Intn(31)+70, rand.Intn(36)+65,
			randomChoice([]string{"Problem solving yang baik", "Tekun dan teliti", "Komunikatif", "Cepat belajar", "Kreatif"}),
			randomChoice([]string{"Perlu lebih teliti", "Time management", "Dokumentasi bisa ditingkatkan", "Komunikasi lebih aktif"}),
			"Secara keseluruhan menunjukkan perkembangan yang "+randomChoice([]string{"baik", "cukup baik", "sangat baik"})+".",
			assessmentDate.Format("2006-01-02"),
		)
		if err == nil {
			created++
		}
	}
	fmt.Printf("✅ %d assessments created\n", created)
}

func seedReports(db *sql.DB, interns []internSeed, minWeeks, maxWeeks int) {
	fmt.Println("Creating reports...")
	created := 0
	for _, intern := range interns {
		if intern.Status == "cancelled" {
			continue
		}
		weeksCount := rand.Intn(maxWeeks-minWeeks+1) + minWeeks
		for w := 1; w <= weeksCount; w++ {
			periodStart := time.Now().AddDate(0, 0, -7*w)
			periodEnd := periodStart.AddDate(0, 0, 6)

			_, err := db.Exec(`
				INSERT INTO reports (intern_id, created_by, title, content, type, period_start, period_end, status, feedback)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`,
				intern.ID, intern.UserID,
				fmt.Sprintf("Laporan Mingguan - Minggu ke-%d", w),
				"Selama minggu ini saya telah mengerjakan beberapa tugas yang diberikan. Berikut ringkasan pekerjaan yang telah diselesaikan...",
				"weekly",
				periodStart.Format("2006-01-02"),
				periodEnd.Format("2006-01-02"),
				randomChoice([]string{"draft", "submitted", "reviewed"}),
				optionalString("Laporan sudah cukup lengkap. Teruskan!", rand.Intn(2) == 0),
			)
			if err == nil {
				created++
			}
		}
	}
	fmt.Printf("✅ %d reports created\n", created)
}

func seedNotifications(db *sql.DB, adminID int64, supervisorIDs []int64, interns []internSeed, perUser int) {
	fmt.Println("Creating notifications...")
	userIDs := []int64{adminID}
	userIDs = append(userIDs, supervisorIDs...)
	for _, intern := range interns {
		userIDs = append(userIDs, intern.UserID)
	}

	for _, uid := range userIDs {
		for i := 0; i < perUser; i++ {
			title := randomChoice([]string{"Tugas baru", "Presensi hari ini", "Pengumuman sistem"})
			message := randomChoice([]string{
				"Jangan lupa mengisi presensi hari ini.",
				"Anda mendapatkan tugas baru. Silakan cek halaman penugasan.",
				"Pengaturan sistem telah diperbarui.",
			})
			_, _ = db.Exec(`
				INSERT INTO notifications (user_id, type, title, message, icon, link)
				VALUES (?, ?, ?, ?, ?, ?)
			`, uid, "info", title, message, "bell", "/notifications")
		}
	}
	fmt.Println("✅ Notifications created")
}

func insertTaskAssignment(
	db *sql.DB,
	adminID int64,
	title, desc, priority string,
	startDate, deadline time.Time,
	deadlineTime string,
	assignToAll bool,
) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO task_assignments (title, description, assigned_by, priority, start_date, deadline, deadline_time, assign_to_all)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, title, desc, adminID, priority, startDate.Format("2006-01-02"), deadline.Format("2006-01-02"), deadlineTime, assignToAll)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func insertTask(
	db *sql.DB,
	assignmentID, internID, adminID int64,
	title, desc, priority, status string,
	startDate, deadline time.Time,
	deadlineTime string,
) (int64, error) {
	startedAt, submittedAt, completedAt, approvedAt, isLate, linksJSON, score, feedback := generateTaskData(status, deadline)

	var res sql.Result
	var err error
	if hasTaskTargetDate(db) {
		res, err = db.Exec(`
			INSERT INTO tasks (
				task_assignment_id, intern_id, assigned_by, title, description, priority, status,
				start_date, deadline, deadline_time, target_date,
				started_at, submitted_at, completed_at, approved_at,
				is_late, submission_links, score, admin_feedback
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			assignmentID, internID, adminID, title, desc, priority, status,
			startDate.Format("2006-01-02"), deadline.Format("2006-01-02"), deadlineTime, deadline.Format("2006-01-02"),
			startedAt, submittedAt, completedAt, approvedAt,
			isLate, linksJSON, score, feedback,
		)
	} else {
		res, err = db.Exec(`
			INSERT INTO tasks (
				task_assignment_id, intern_id, assigned_by, title, description, priority, status,
				start_date, deadline, deadline_time,
				started_at, submitted_at, completed_at, approved_at,
				is_late, submission_links, score, admin_feedback
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			assignmentID, internID, adminID, title, desc, priority, status,
			startDate.Format("2006-01-02"), deadline.Format("2006-01-02"), deadlineTime,
			startedAt, submittedAt, completedAt, approvedAt,
			isLate, linksJSON, score, feedback,
		)
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func generateTaskData(status string, deadline time.Time) (started, submitted, completed, approved *time.Time, isLate bool, linksJSON *string, score *int, feedback *string) {
	switch status {
	case "in_progress", "submitted", "revision", "completed":
		t := deadline.AddDate(0, 0, -rand.Intn(8)-3)
		started = &t
	}

	if status == "submitted" || status == "revision" || status == "completed" {
		t := time.Now().AddDate(0, 0, -rand.Intn(4))
		submitted = &t
		isLate = submitted.After(deadline)
		linksJSON = submissionLinksJSON()
	}

	if status == "completed" {
		t := time.Now().AddDate(0, 0, -rand.Intn(7))
		completed = &t
		sub := completed.Add(-time.Duration(rand.Intn(24)+1) * time.Hour)
		submitted = &sub
		app := completed.Add(time.Duration(rand.Intn(48)+1) * time.Hour)
		approved = &app
		isLate = submitted.After(deadline)
		s := rand.Intn(31) + 70
		score = &s
		if s >= 85 {
			msg := "Kerja bagus! Hasilnya sesuai ekspektasi."
			feedback = &msg
		} else {
			msg := "Sudah cukup baik, perlu sedikit improvement untuk kedepannya."
			feedback = &msg
		}
		linksJSON = submissionLinksJSON()
	}

	if status == "revision" {
		msg := "Perlu perbaikan pada bagian " + randomChoice([]string{"UI/UX", "validasi data", "error handling", "dokumentasi"}) + "."
		feedback = &msg
	}

	return started, submitted, completed, approved, isLate, linksJSON, score, feedback
}

func generateAttendanceData(date time.Time) (status string, checkIn *time.Time, checkOut *time.Time, lateReason *string, notes *string) {
	r := rand.Intn(100) + 1
	switch {
	case r <= 75:
		status = "present"
		ci := time.Date(date.Year(), date.Month(), date.Day(), 8, rand.Intn(16), 0, 0, time.Local)
		co := time.Date(date.Year(), date.Month(), date.Day(), rand.Intn(2)+16, rand.Intn(60), 0, 0, time.Local)
		return status, &ci, &co, nil, nil
	case r <= 88:
		status = "late"
		ci := time.Date(date.Year(), date.Month(), date.Day(), rand.Intn(2)+8, rand.Intn(44)+16, 0, 0, time.Local)
		co := time.Date(date.Year(), date.Month(), date.Day(), rand.Intn(2)+16, rand.Intn(60), 0, 0, time.Local)
		reason := randomChoice([]string{"Macet", "Hujan deras", "Kendaraan mogok", "Keperluan keluarga"})
		return status, &ci, &co, &reason, nil
	case r <= 94:
		status = "sick"
		note := "Sakit " + randomChoice([]string{"flu", "demam", "migrain"})
		return status, nil, nil, nil, &note
	case r <= 98:
		status = "permission"
		note := "Keperluan keluarga"
		return status, nil, nil, nil, &note
	default:
		status = "absent"
		return status, nil, nil, nil, nil
	}
}

func submissionLinksJSON() *string {
	data := map[string]string{
		"github": fmt.Sprintf("https://github.com/user/project-%d", rand.Intn(900)+100),
		"demo":   fmt.Sprintf("https://demo.example.com/project-%d", rand.Intn(900)+100),
	}
	b, _ := json.Marshal(data)
	s := string(b)
	return &s
}

func createOrGetUser(db *sql.DB, name, email, password, role string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	res, err := db.Exec(`
		INSERT INTO users (name, email, password_hash, role)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE name = VALUES(name), password_hash = VALUES(password_hash), role = VALUES(role), id = LAST_INSERT_ID(id)
	`, name, email, string(hash), role)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ensureSupervisorProfile(db *sql.DB, userID int64, fullName, nip string) {
	var id int64
	err := db.QueryRow("SELECT id FROM supervisors WHERE user_id = ?", userID).Scan(&id)
	if err == sql.ErrNoRows {
		_, _ = db.Exec(`
			INSERT INTO supervisors (user_id, full_name, nip, phone, institution, status)
			VALUES (?, ?, ?, ?, ?, ?)
		`, userID, fullName, nip, randomPhone(), "Direktorat Sistem Informasi", "active")
		return
	}
	if err == nil {
		_, _ = db.Exec(`
			UPDATE supervisors SET full_name = ?, nip = ?, phone = ?, institution = ?, status = ?
			WHERE user_id = ?
		`, fullName, nip, randomPhone(), "Direktorat Sistem Informasi", "active", userID)
	}
}

func ensureInstitution(db *sql.DB, name, address, phone, email string) (int64, error) {
	var id int64
	err := db.QueryRow("SELECT id FROM institutions WHERE name = ?", name).Scan(&id)
	if err == nil {
		return id, nil
	}
	res, err := db.Exec(`
		INSERT INTO institutions (name, address, phone, email)
		VALUES (?, ?, ?, ?)
	`, name, address, phone, email)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ensureInternProfile(
	db *sql.DB,
	userID, institutionID, supervisorID int64,
	index int,
	fullName, status string,
	startDate, endDate time.Time,
) (int64, error) {
	var id int64
	err := db.QueryRow("SELECT id FROM interns WHERE user_id = ?", userID).Scan(&id)
	if err == nil {
		_, _ = db.Exec(`
			UPDATE interns
			SET institution_id = ?, supervisor_id = ?, full_name = ?, nis = ?, student_id = ?,
			    school = ?, department = ?, date_of_birth = ?, gender = ?, phone = ?, address = ?, start_date = ?, end_date = ?, status = ?
			WHERE user_id = ?
		`,
			institutionID, supervisorID, fullName,
			fmt.Sprintf("NIS%04d", index), fmt.Sprintf("INT%04d", index),
			randomChoice(schools), randomChoice(departments),
			time.Date(2000, time.Month(rand.Intn(12)+1), rand.Intn(28)+1, 0, 0, 0, 0, time.Local),
			randomChoice([]string{"male", "female"}),
			randomPhone(),
			fmt.Sprintf("Jl. %s No. %d, Jakarta", randomChoice(lastNames), rand.Intn(100)+1),
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), status,
			userID,
		)
		return id, nil
	}

	res, err := db.Exec(`
		INSERT INTO interns (
			user_id, institution_id, supervisor_id, full_name, nis, student_id, school, department,
			date_of_birth, gender, phone, address, start_date, end_date, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		userID, institutionID, supervisorID, fullName,
		fmt.Sprintf("NIS%04d", index), fmt.Sprintf("INT%04d", index),
		randomChoice(schools), randomChoice(departments),
		time.Date(2000, time.Month(rand.Intn(12)+1), rand.Intn(28)+1, 0, 0, 0, 0, time.Local),
		randomChoice([]string{"male", "female"}),
		randomPhone(),
		fmt.Sprintf("Jl. %s No. %d, Jakarta", randomChoice(lastNames), rand.Intn(100)+1),
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), status,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func filterActiveInterns(interns []internSeed) []internSeed {
	var active []internSeed
	for _, i := range interns {
		if i.Status == "active" {
			active = append(active, i)
		}
	}
	if len(active) == 0 {
		return interns
	}
	return active
}

func pickInterns(interns []internSeed, assignToAll bool) []internSeed {
	if assignToAll || len(interns) == 0 {
		return interns
	}
	count := rand.Intn(8) + 5
	if count > len(interns) {
		count = len(interns)
	}
	shuffled := append([]internSeed{}, interns...)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	return shuffled[:count]
}

func determineInternStatus(endDate time.Time) string {
	if endDate.Before(time.Now()) {
		return "completed"
	}
	if rand.Intn(20) == 0 {
		return "cancelled"
	}
	return "active"
}

func uniqueEmail(first, last string, used map[string]bool) string {
	base := strings.ToLower(first + "." + last)
	email := base + "@student.id"
	counter := 1
	for used[email] {
		email = fmt.Sprintf("%s%d@student.id", base, counter)
		counter++
	}
	used[email] = true
	return email
}

func loadExistingInterns(db *sql.DB, limit int) ([]internSeed, map[string]bool) {
	interns := []internSeed{}
	usedEmails := map[string]bool{}
	rows, err := db.Query(`
		SELECT i.id, i.user_id, COALESCE(i.status, 'active') as status, u.email
		FROM interns i
		JOIN users u ON i.user_id = u.id
		ORDER BY i.id
		LIMIT ?
	`, limit)
	if err != nil {
		return interns, usedEmails
	}
	defer rows.Close()
	for rows.Next() {
		var id, userID int64
		var status, email string
		if err := rows.Scan(&id, &userID, &status, &email); err != nil {
			continue
		}
		usedEmails[email] = true
		interns = append(interns, internSeed{ID: id, UserID: userID, Status: status})
	}
	return interns, usedEmails
}

func hasTaskTargetDate(db *sql.DB) bool {
	if tasksHasTargetDate != nil {
		return *tasksHasTargetDate
	}
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'tasks'
		  AND COLUMN_NAME = 'target_date'
	`).Scan(&count)
	has := err == nil && count > 0
	tasksHasTargetDate = &has
	return has
}

func randomPhone() string {
	return fmt.Sprintf("08%d%07d", rand.Intn(90)+10, rand.Intn(9000000)+1000000)
}

func randomChoice(list []string) string {
	return list[rand.Intn(len(list))]
}

func optionalString(value string, enable bool) interface{} {
	if !enable {
		return nil
	}
	return value
}

func pickID(ids []int64) int64 {
	if len(ids) == 0 {
		return 0
	}
	return ids[rand.Intn(len(ids))]
}

func hasRows(db *sql.DB, table string) bool {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func env(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func envInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	val := strings.ToLower(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val == "1" || val == "true" || val == "yes"
}

func printSummary() {
	fmt.Println("")
	fmt.Println("🎉 Seed data created successfully!")
	fmt.Println("")
	fmt.Println("📝 Demo Account Credentials:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("")
	fmt.Println("👤 Admin Account:")
	fmt.Println("   Email:    admin@internapro.id")
	fmt.Println("   Password: password")
	fmt.Println("")
	fmt.Println("👨‍💼 Supervisor Accounts:")
	fmt.Println("   - bambang.herlambang@internapro.id")
	fmt.Println("   - khamizar@internapro.id")
	fmt.Println("   - mega.novita@internapro.id")
	fmt.Println("   Password: password")
	fmt.Println("")
	fmt.Println("👩‍💻 Intern Accounts:")
	fmt.Println("   Email:    <generated>@student.id")
	fmt.Println("   Password: password")
	fmt.Println("")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Tip: set SEED_FORCE=true to reseed demo data.")
}
