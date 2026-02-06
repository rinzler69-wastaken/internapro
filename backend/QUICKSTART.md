# 🚀 INTERNA - Quick Start Guide

## What You Have

A complete Go backend structure for an internship management system with:

✅ **Database Schema** - MySQL with all tables
✅ **Configuration System** - Environment-based config
✅ **Models** - Complete data structures
✅ **Utilities** - Geolocation, file upload, HTTP responses
✅ **Middleware** - JWT authentication & role-based access
✅ **Routes** - All API endpoints mapped
✅ **2 Complete Handlers** - Auth & Attendance (fully implemented)
✅ **6 Stub Handlers** - Ready for you to implement

## Project Structure

```
dsi-interna.sys/
├── cmd/server/main.go          ✅ Server entry point
├── database/schema.sql          ✅ MySQL schema
├── internal/
│   ├── config/config.go         ✅ Configuration loader
│   ├── database/mysql.go        ✅ Database connection
│   ├── models/                  ✅ Data models
│   │   ├── user.go
│   │   ├── intern.go
│   │   ├── task.go
│   │   ├── attendance.go
│   │   ├── leave.go
│   │   └── assessment.go
│   ├── handlers/                ✅/🚧 Request handlers
│   │   ├── auth.go              ✅ COMPLETE
│   │   ├── attendance.go        ✅ COMPLETE
│   │   ├── intern.go            🚧 STUB (needs implementation)
│   │   ├── task.go              🚧 STUB (needs implementation)
│   │   ├── leave.go             🚧 STUB (needs implementation)
│   │   ├── assessment.go        🚧 STUB (needs implementation)
│   │   └── report.go            🚧 STUB (needs implementation)
│   ├── middleware/
│   │   └── auth.go              ✅ JWT middleware
│   ├── routes/routes.go         ✅ API routing
│   └── utils/                   ✅ Utilities
│       ├── geo.go               ✅ Geolocation (Haversine)
│       ├── file.go              ✅ File upload
│       └── response.go          ✅ HTTP responses
├── .env.example                 ✅ Environment template
├── go.mod                       ✅ Dependencies
├── README.md                    ✅ Full documentation
├── IMPLEMENTATION_GUIDE.md      ✅ Implementation guide
└── setup.sh                     ✅ Setup script
```

## 🏃‍♂️ Quick Setup (5 minutes)

### Option 1: Automated Setup

```bash
# Run the setup script
chmod +x setup.sh
./setup.sh

# Follow the prompts
# Then edit .env with your database credentials
nano .env

# Start the server
go run cmd/server/main.go
```

### Option 2: Manual Setup

```bash
# 1. Install dependencies
go mod download

# 2. Setup environment
cp .env.example .env
# Edit .env with your database credentials

# 3. Create database
mysql -u root -p
CREATE DATABASE interna_db;
USE interna_db;
source database/schema.sql;
exit;

# 4. Create upload directories
mkdir -p uploads/tasks uploads/leaves

# 5. Start server
go run cmd/server/main.go
```

## ✅ Test Your Setup

```bash
# Health check
curl http://localhost:8080/api/health

# Expected response:
# {"status":"ok","message":"INTERNA API is running"}
```

## 🔑 Test Authentication

### 1. Register a New User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "supervisor@dsi_interna_sys.com",
    "password": "password123",
    "role": "supervisor",
    "full_name": "John Supervisor",
    "nip": "123456",
    "position": "Senior Developer"
  }'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "supervisor@dsi_interna_sys.com",
    "password": "password123"
  }'
```

Save the token from the response!

### 3. Get Current User

```bash
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 📱 Test Attendance (Geolocation Feature)

This is the coolest feature - geolocation-based check-in!

### Check In

```bash
curl -X POST http://localhost:8080/api/attendance/checkin \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": -7.035549620262833,
    "longitude": 110.47464898482643
  }'
```

**Important**: The default office location is set to Semarang coordinates. If you're testing from a different location, either:
1. Update office coordinates in `.env` to match your test location
2. Use coordinates within 1km of the office location

### Check Today's Attendance

```bash
curl http://localhost:8080/api/attendance/today \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 🎯 What To Implement Next

You have 6 handlers to complete. Here's the recommended order:

### Priority 1 (Core Features)
1. **Intern Handler** (`internal/handlers/intern.go`)
   - CRUD operations for interns
   - This is fundamental

2. **Task Handler** (`internal/handlers/task.go`)
   - CRUD operations for tasks
   - File upload for attachments

### Priority 2 (Important Features)
3. **Leave Handler** (`internal/handlers/leave.go`)
   - Leave request management
   - Approval/rejection workflow

4. **Assessment Handler** (`internal/handlers/assessment.go`)
   - Scoring system (0-100)
   - Auto-categorization

### Priority 3 (Reporting)
5. **Report Handler** (`internal/handlers/report.go`)
   - Generate various reports
   - Certificate generation

## 📝 Implementation Template

Here's how to implement any handler:

```go
func (h *YourHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Get user from context
    claims, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        utils.RespondUnauthorized(w, "Unauthorized")
        return
    }
    
    // 2. Parse request body
    var req YourRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.RespondBadRequest(w, "Invalid request body")
        return
    }
    
    // 3. Validate input
    if req.Field == "" {
        utils.RespondBadRequest(w, "Field is required")
        return
    }
    
    // 4. Execute database operation
    result, err := h.db.Exec(
        "INSERT INTO table (field) VALUES (?)",
        req.Field,
    )
    if err != nil {
        utils.RespondInternalError(w, "Database error")
        return
    }
    
    // 5. Get inserted ID
    id, _ := result.LastInsertId()
    
    // 6. Return success
    utils.RespondCreated(w, "Created successfully", map[string]interface{}{
        "id": id,
    })
}
```

## 🔧 Common Tasks

### Update Office Location

Edit `.env`:
```env
OFFICE_LATITUDE=-7.035549620262833    # Your office latitude
OFFICE_LONGITUDE=110.47464898482643  # Your office longitude
OFFICE_RADIUS=1000            # Radius in meters
```

### Change Check-in/Check-out Times

Edit `.env`:
```env
CHECK_IN_TIME=08:00:00        # Official check-in time
CHECK_OUT_TIME=17:00:00       # Official check-out time
LATE_TOLERANCE_MINUTES=15     # Grace period
```

### Enable 2FA for a User

```bash
# 1. Setup 2FA (get QR code)
curl -X POST http://localhost:8080/api/auth/2fa/setup \
  -H "Authorization: Bearer YOUR_TOKEN"

# 2. Scan QR code with Google Authenticator app

# 3. Verify with code from app
curl -X POST http://localhost:8080/api/auth/2fa/verify \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "123456"}'
```

## 📚 API Documentation

All endpoints are defined in `/internal/routes/routes.go`:

**Public Endpoints:**
- `POST /api/auth/register` - Register
- `POST /api/auth/login` - Login

**Protected Endpoints (require JWT):**
- Auth: `/api/auth/*`
- Interns: `/api/interns/*`
- Tasks: `/api/tasks/*`
- Attendance: `/api/attendance/*`
- Leaves: `/api/leaves/*`
- Assessments: `/api/assessments/*`
- Reports: `/api/reports/*`

## 🐛 Troubleshooting

**Problem:** Database connection failed
```bash
# Check MySQL is running
sudo service mysql status

# Verify credentials in .env
cat .env
```

**Problem:** Port 8080 already in use
```bash
# Change port in .env
SERVER_PORT=8081
```

**Problem:** File upload fails
```bash
# Check upload directory exists and is writable
ls -la uploads/
chmod -R 755 uploads/
```

**Problem:** Geolocation check-in fails (not within radius)
```bash
# Update office coordinates in .env to your test location
# OR use test coordinates within 1km of office location
```

## 📖 Learn More

- **Full Documentation**: `README.md`
- **Implementation Guide**: `IMPLEMENTATION_GUIDE.md`
- **Database Schema**: `database/schema.sql`

## 🎓 Week Plan

**Day 1-2:** Complete Intern & Task handlers
**Day 3-4:** Complete Leave & Assessment handlers
**Day 5:** Complete Report handler
**Day 6:** Build basic frontend (HTML/Tailwind)
**Day 7:** Testing, bug fixes, polish

Good luck with your internship project! 🚀

Need help? Check the IMPLEMENTATION_GUIDE.md for detailed examples.
