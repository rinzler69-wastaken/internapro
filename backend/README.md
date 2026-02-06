# INTERNA - Internship Management System

Sistem Manajemen Magang berbasis web dengan teknologi Go (Backend) dan HTML/CSS/Tailwind (Frontend).

## 🚀 Features

1. **Manajemen Siswa Magang** - Data siswa, pembimbing, periode, institusi
2. **Penugasan** - Insert tugas dengan upload file (jpg, jpeg, png, pdf)
3. **Sistem Presensi** - Berbasis geolokasi (<1km dari kantor) dengan jam hadir/pulang
4. **Perizinan** - Form perizinan dengan upload surat izin
5. **Penilaian** - Scoring 0-100 dengan 4 indikator
6. **Laporan** - Laporan siswa, absensi, izin, penilaian, sertifikat
7. **Autentikasi** - Login + Google Authenticator (2FA)

## 📁 Project Structure

```
dsi_interna_sys/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration
│   ├── database/
│   │   └── mysql.go             # Database connection
│   ├── models/
│   │   ├── user.go
│   │   ├── intern.go
│   │   ├── task.go
│   │   ├── attendance.go
│   │   ├── leave.go
│   │   └── assessment.go
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── intern.go
│   │   ├── task.go
│   │   ├── attendance.go
│   │   ├── leave.go
│   │   ├── assessment.go
│   │   └── report.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── logger.go
│   │   └── cors.go
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── intern_service.go
│   │   ├── task_service.go
│   │   ├── attendance_service.go
│   │   ├── leave_service.go
│   │   ├── assessment_service.go
│   │   └── report_service.go
│   └── utils/
│       ├── geo.go               # Geolocation utilities
│       ├── file.go              # File upload utilities
│       ├── validator.go         # Validation
│       └── response.go          # HTTP response helpers
├── database/
│   └── schema.sql               # MySQL schema
├── uploads/                     # File uploads directory
├── web/                         # Frontend files
│   ├── static/
│   │   ├── css/
│   │   └── js/
│   └── templates/
│       └── *.html
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

## 🛠️ Tech Stack

- **Backend**: Go 1.21+
- **Database**: MySQL 8.0+
- **Frontend**: HTML, CSS, Tailwind CSS
- **Authentication**: JWT + Google Authenticator (TOTP)

## 📦 Dependencies

- `gorilla/mux` - HTTP router
- `go-sql-driver/mysql` - MySQL driver
- `golang-jwt/jwt` - JWT authentication
- `pquerna/otp` - Google Authenticator
- `golang.org/x/crypto` - Password hashing
- `joho/godotenv` - Environment variables
- `rs/cors` - CORS middleware

## ⚙️ Setup

### 1. Clone & Install Dependencies

```bash
go mod download
```

### 2. Database Setup

```bash
# Create database
mysql -u root -p < database/schema.sql

# Or manually
mysql -u root -p
CREATE DATABASE interna_db;
USE interna_db;
source database/schema.sql;
```

### 3. Environment Configuration

Create `.env` file:

```env
# Server
SERVER_PORT=8080
SERVER_HOST=localhost

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=interna_db

# JWT
JWT_SECRET=your-super-secret-key-change-this
JWT_EXPIRY=24h

# Upload
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE=5242880  # 5MB

# Office Location (Semarang example)
OFFICE_LATITUDE=-7.035549620262833
OFFICE_LONGITUDE=110.47464898482643
OFFICE_RADIUS=1000  # meters

# Time Settings
CHECK_IN_TIME=08:00:00
CHECK_OUT_TIME=17:00:00
LATE_TOLERANCE_MINUTES=15
```

### 4. Create Upload Directory

```bash
mkdir -p uploads/tasks uploads/leaves
```

### 5. Run Application

```bash
go run cmd/server/main.go
```

Server will start at `http://localhost:8080`

## 🔐 API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/2fa/setup` - Setup Google Authenticator
- `POST /api/auth/2fa/verify` - Verify 2FA token
- `POST /api/auth/logout` - Logout

### Interns
- `GET /api/interns` - List all interns
- `GET /api/interns/:id` - Get intern details
- `POST /api/interns` - Create intern
- `PUT /api/interns/:id` - Update intern
- `DELETE /api/interns/:id` - Delete intern

### Tasks
- `GET /api/tasks` - List tasks
- `GET /api/tasks/:id` - Get task details
- `POST /api/tasks` - Create task
- `PUT /api/tasks/:id` - Update task
- `DELETE /api/tasks/:id` - Delete task
- `POST /api/tasks/:id/attachments` - Upload attachment

### Attendance
- `POST /api/attendance/checkin` - Check in
- `POST /api/attendance/checkout` - Check out
- `GET /api/attendance` - Get attendance records
- `GET /api/attendance/:id` - Get specific attendance

### Leave Requests
- `GET /api/leaves` - List leave requests
- `POST /api/leaves` - Create leave request
- `PUT /api/leaves/:id` - Update leave request
- `POST /api/leaves/:id/approve` - Approve leave
- `POST /api/leaves/:id/reject` - Reject leave

### Assessments
- `GET /api/assessments` - List assessments
- `POST /api/assessments` - Create assessment
- `PUT /api/assessments/:id` - Update assessment
- `GET /api/assessments/intern/:id` - Get intern assessments

### Reports
- `GET /api/reports/intern/:id` - Intern report
- `GET /api/reports/attendance/:id` - Attendance report
- `GET /api/reports/certificate/:id` - Certificate
- `POST /api/reports/certificate/:id/generate` - Generate certificate

## 🗺️ Database Schema Highlights

### Key Tables
- `users` - Authentication & roles
- `interns` - Student data & biography
- `supervisors` - Supervisor information
- `tasks` - Task assignments
- `task_attachments` - File uploads
- `attendances` - Check-in/out records
- `leave_requests` - Leave permissions
- `assessments` - Scoring & evaluation
- `certificates` - Final reports

### Assessment Categories (Auto-calculated)
- Score 85-100: Very Good (Sangat Baik)
- Score 70-84: Good (Baik)
- Score 50-69: Not Good (Tidak Baik)
- Score 0-49: Very Bad (Sangat Tidak Baik)

## 📍 Geolocation Logic

Attendance check-in/out requires:
1. Distance < 1km from office location (configurable)
2. Within allowed time window
3. If late, reason must be provided

Distance calculation uses Haversine formula.

## 📝 File Upload Rules

- Allowed formats: JPG, JPEG, PNG, PDF
- Maximum size: 5MB (configurable)
- Stored in `/uploads` directory
- File paths saved in database

## 🔒 Security Features

- Password hashing (bcrypt)
- JWT authentication
- Google Authenticator 2FA
- CORS protection
- Input validation
- SQL injection prevention (prepared statements)
- Activity logging

## 📊 Reporting Features

Generate reports for:
1. Intern activity summary
2. Attendance records
3. Leave history
4. Assessment scores
5. Certificate/report card (rapor)

## 🎯 Next Steps

1. Complete backend implementation
2. Create frontend templates
3. Integrate Tailwind CSS
4. Testing & debugging
5. Deploy to production

## 📄 License

Internal use only - Internship project
