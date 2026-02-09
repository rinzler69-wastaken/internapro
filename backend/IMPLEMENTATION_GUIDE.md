# INTERNA Backend Implementation Guide

## 📦 What We've Built So Far

### ✅ Completed Components

1. **Database Schema** (`database/schema.sql`)
   - Complete MySQL schema with all tables
   - Proper indexes and foreign keys
   - Auto-calculated assessment categories
   - Activity logging support

2. **Configuration System** (`internal/config/`)
   - Environment variable loading
   - Database configuration
   - JWT settings
   - Upload file settings
   - Office location settings

3. **Database Connection** (`internal/database/`)
   - MySQL connection with pooling
   - Connection health check

4. **Models** (`internal/models/`)
   - User, Supervisor, Institution
   - Intern with details
   - Task with attachments
   - Attendance with geolocation
   - Leave requests
   - Assessments with auto-categories

5. **Utilities** (`internal/utils/`)
   - Geolocation (Haversine distance calculation)
   - File upload/download handling
   - HTTP response helpers
   - Pagination support

6. **Middleware** (`internal/middleware/`)
   - JWT authentication
   - Role-based access control

7. **Routes** (`internal/routes/`)
   - Complete API endpoint structure
   - Protected and public routes
   - Static file serving

8. **Main Server** (`cmd/server/main.go`)
   - Server initialization
   - Graceful shutdown
   - CORS configuration

## 🚧 What Needs to Be Implemented

### Priority 1: Core Handlers (Critical)

You need to create handler files for each domain. Here's the structure:

#### 1. Auth Handler (`internal/handlers/auth.go`)
```go
type AuthHandler struct {
    db *sql.DB
}

Required methods:
- Register(w, r)           // Create new user
- Login(w, r)              // Login with email/password
- Setup2FA(w, r)           // Generate QR code for Google Authenticator
- Verify2FA(w, r)          // Verify TOTP token
- Disable2FA(w, r)         // Disable 2FA
- GetCurrentUser(w, r)     // Get logged-in user info
- Logout(w, r)             // Logout (optional, mostly client-side)
```

#### 2. Intern Handler (`internal/handlers/intern.go`)
```go
Required methods:
- GetAll(w, r)       // List all interns (with pagination)
- GetByID(w, r)      // Get single intern
- Create(w, r)       // Create new intern
- Update(w, r)       // Update intern
- Delete(w, r)       // Soft/hard delete intern
```

#### 3. Task Handler (`internal/handlers/task.go`)
```go
Required methods:
- GetAll(w, r)            // List all tasks
- GetByID(w, r)           // Get single task
- GetByInternID(w, r)     // Get tasks for specific intern
- Create(w, r)            // Create new task
- Update(w, r)            // Update task
- Delete(w, r)            // Delete task
- UploadAttachment(w, r)  // Upload file to task
- MarkComplete(w, r)      // Mark task as completed
```

#### 4. Attendance Handler (`internal/handlers/attendance.go`)
```go
Required methods:
- CheckIn(w, r)          // Check in with geolocation
- CheckOut(w, r)         // Check out
- GetAll(w, r)           // List all attendance
- GetByID(w, r)          // Get single attendance
- GetByInternID(w, r)    // Get intern's attendance history
- GetToday(w, r)         // Get today's attendance for user
```

#### 5. Leave Handler (`internal/handlers/leave.go`)
```go
Required methods:
- GetAll(w, r)           // List all leave requests
- GetByID(w, r)          // Get single leave request
- GetByInternID(w, r)    // Get intern's leave history
- Create(w, r)           // Create leave request
- Update(w, r)           // Update leave request
- Approve(w, r)          // Approve leave (supervisor only)
- Reject(w, r)           // Reject leave (supervisor only)
- UploadAttachment(w, r) // Upload leave letter
```

#### 6. Assessment Handler (`internal/handlers/assessment.go`)
```go
Required methods:
- GetAll(w, r)           // List all assessments
- GetByID(w, r)          // Get single assessment
- GetByInternID(w, r)    // Get intern's assessments
- Create(w, r)           // Create assessment (supervisor only)
- Update(w, r)           // Update assessment
- Delete(w, r)           // Delete assessment
```

#### 7. Report Handler (`internal/handlers/report.go`)
```go
Required methods:
- GetInternReport(w, r)      // Comprehensive intern report
- GetAttendanceReport(w, r)  // Attendance summary
- GetAssessmentReport(w, r)  // Assessment summary
- GetCertificate(w, r)       // Get certificate
- GenerateCertificate(w, r)  // Generate new certificate
```

### Priority 2: Services Layer (Optional but Recommended)

Create business logic layer to keep handlers clean:

```
internal/services/
├── auth_service.go
├── intern_service.go
├── task_service.go
├── attendance_service.go
├── leave_service.go
├── assessment_service.go
└── report_service.go
```

### Priority 3: Additional Utilities

1. **Validator** (`internal/utils/validator.go`)
   - Email validation
   - Phone validation
   - Date range validation
   - Custom validators

2. **Password Hasher** (`internal/utils/password.go`)
   - bcrypt hashing
   - Password verification

3. **JWT Helper** (`internal/utils/jwt.go`)
   - Token generation
   - Token validation
   - Refresh token logic

4. **Logger Middleware** (`internal/middleware/logger.go`)
   - Request logging
   - Response time tracking

## 🔧 Implementation Steps

### Step 1: Setup Development Environment

```bash
# 1. Clone/create project
cd /path/to/project

# 2. Install dependencies
go mod download

# 3. Create .env file
cp .env.example .env
# Edit .env with your database credentials

# 4. Create database
mysql -u root -p
CREATE DATABASE interna_db;
USE interna_db;
source database/schema.sql;

# 5. Create upload directories
mkdir -p uploads/tasks uploads/leaves
```

### Step 2: Implement Auth Handler (Start Here!)

This is the most critical handler. Example structure:

```go
package handlers

import (
    "database/sql"
    "net/http"
    "time"
    
    "dsi_interna_sys/internal/models"
    "dsi_interna_sys/internal/utils"
    "dsi_interna_sys/internal/config"
    "dsi_interna_sys/internal/middleware"
    
    "github.com/golang-jwt/jwt/v5"
    "github.com/pquerna/otp/totp"
    "golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
    db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
    return &AuthHandler{db: db}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body
    // 2. Validate input
    // 3. Hash password
    // 4. Insert user into database
    // 5. Create associated supervisor/intern record
    // 6. Return success response
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // 1. Parse credentials
    // 2. Find user in database
    // 3. Verify password
    // 4. Check if 2FA is enabled
    // 5. Generate JWT token
    // 6. Return token
}

func (h *AuthHandler) Setup2FA(w http.ResponseWriter, r *http.Request) {
    // 1. Get user from context
    // 2. Generate TOTP secret
    // 3. Generate QR code
    // 4. Save secret to database
    // 5. Return QR code URL
}

// ... implement other methods
```

### Step 3: Implement Attendance Handler (Core Feature)

This is critical for the geolocation-based check-in:

```go
func (h *AttendanceHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
    // 1. Get user from context
    // 2. Parse latitude/longitude from request
    // 3. Validate coordinates
    // 4. Get office settings from config
    // 5. Calculate distance using Haversine
    // 6. Check if within radius
    // 7. Check current time vs allowed check-in time
    // 8. If late, require reason
    // 9. Create attendance record
    // 10. Return success
}
```

### Step 4: Implement Task Handler

```go
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Get supervisor from context
    // 2. Parse task data
    // 3. Validate dates
    // 4. Insert task
    // 5. Handle file uploads if present
    // 6. Return created task
}

func (h *TaskHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
    // 1. Get task ID from URL
    // 2. Verify task exists
    // 3. Parse multipart form
    // 4. Validate file (jpg, jpeg, png, pdf only)
    // 5. Save file using utils.UploadFile
    // 6. Create attachment record
    // 7. Return file info
}
```

### Step 5: Implement Other Handlers

Follow similar patterns for:
- Leave requests
- Assessments
- Reports

### Step 6: Testing

```bash
# Run the server
go run cmd/server/main.go

# Test health endpoint
curl http://localhost:8080/api/health

# Test registration
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "intern@example.com",
    "password": "password123",
    "role": "intern",
    "full_name": "John Doe"
  }'

# Test login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "intern@example.com",
    "password": "password123"
  }'
```

## 📝 Handler Implementation Template

Here's a template for any handler:

```go
package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    
    "dsi_interna_sys/internal/models"
    "dsi_interna_sys/internal/utils"
    "dsi_interna_sys/internal/middleware"
    
    "github.com/gorilla/mux"
)

type YourHandler struct {
    db *sql.DB
}

func NewYourHandler(db *sql.DB) *YourHandler {
    return &YourHandler{db: db}
}

func (h *YourHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    // Get user from context
    claims, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        utils.RespondUnauthorized(w, "Unauthorized")
        return
    }
    
    // Parse query parameters (pagination, filters)
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if page < 1 { page = 1 }
    if limit < 1 { limit = 10 }
    
    offset := (page - 1) * limit
    
    // Query database
    query := `SELECT * FROM your_table LIMIT ? OFFSET ?`
    rows, err := h.db.Query(query, limit, offset)
    if err != nil {
        utils.RespondInternalError(w, "Database error")
        return
    }
    defer rows.Close()
    
    // Scan results
    var items []models.YourModel
    for rows.Next() {
        var item models.YourModel
        if err := rows.Scan(&item.ID, &item.Field1 /* ... */); err != nil {
            continue
        }
        items = append(items, item)
    }
    
    // Get total count
    var total int64
    h.db.QueryRow("SELECT COUNT(*) FROM your_table").Scan(&total)
    
    // Return paginated response
    pagination := utils.CalculatePagination(page, limit, total)
    utils.RespondPaginated(w, items, pagination)
}

func (h *YourHandler) Create(w http.ResponseWriter, r *http.Request) {
    // Similar pattern: parse, validate, insert, respond
}

// ... other methods
```

## 🎯 Quick Start Checklist

- [ ] Setup .env file with database credentials
- [ ] Run database schema
- [ ] Create upload directories
- [ ] Implement AuthHandler
- [ ] Implement AttendanceHandler (geolocation feature)
- [ ] Implement TaskHandler (with file uploads)
- [ ] Implement LeaveHandler
- [ ] Implement AssessmentHandler
- [ ] Implement ReportHandler
- [ ] Test all endpoints
- [ ] Create basic HTML frontend
- [ ] Deploy

## 💡 Pro Tips

1. **Start with Auth**: Get login/register working first
2. **Test as you go**: Use curl or Postman after each handler
3. **Use transactions**: For operations that modify multiple tables
4. **Log everything**: Add logging to debug issues
5. **Handle errors properly**: Always check errors and return appropriate responses
6. **Validate input**: Never trust user input
7. **Use prepared statements**: Prevent SQL injection

## 🐛 Common Issues & Solutions

1. **Database connection fails**
   - Check MySQL is running: `sudo service mysql status`
   - Verify credentials in .env
   - Ensure database exists

2. **File upload fails**
   - Check upload directory exists and is writable
   - Verify file size limits
   - Check file extension validation

3. **Geolocation not working**
   - Ensure coordinates are valid (-90 to 90 for lat, -180 to 180 for lon)
   - Check office settings in database
   - Verify Haversine calculation

4. **JWT token invalid**
   - Check JWT_SECRET in .env
   - Ensure token not expired
   - Verify Authorization header format

Good luck with your implementation! 🚀
