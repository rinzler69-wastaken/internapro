# Dashboard Implementation Plan

## Overview
Reverse engineer the Laravel intern dashboard with toggle between Presensi (Map) and Tugas Aktif views. Keep white theme with Vercel-inspired styling, CSS-based visualizations.

## Tasks

### Phase 1: Backend API Updates
- [ ] 1.1 Update `GetInternDashboard` to include task_breakdown (pending, in_progress, submitted, completed, revision)
- [ ] 1.2 Add weekly_attendance_counts formatted for CSS charts
- [ ] 1.3 Add recent_attendance_history to response
- [ ] 1.4 Ensure all existing data is properly structured

### Phase 2: Frontend API Integration
- [ ] 2.1 Add API endpoints for check-in/check-out in api.js
- [ ] 2.2 Add permission submission endpoint
- [ ] 2.3 Create dashboard store for state management

### Phase 3: Dashboard Structure (Dashboard.svelte)
- [ ] 3.1 Welcome header with user name + formatted date
- [ ] 3.2 Toggle switch (Presensi / Tugas Aktif) at top
- [ ] 3.3 Main content area with conditional rendering

### Phase 4: Presensi View (Left Column)
- [ ] 4.1 Leaflet map container for GPS attendance
- [ ] 4.2 Office location marker and radius circle
- [ ] 4.3 User location marker with real-time updates
- [ ] 4.4 Distance display from office
- [ ] 4.5 GPS status indicator
- [ ] 4.6 Check-in button (disabled until valid GPS)
- [ ] 4.7 Check-out button (when already checked in)
- [ ] 4.8 Completed attendance state

### Phase 5: Permission/Sick Modal
- [ ] 5.1 Modal overlay with backdrop blur
- [ ] 5.2 Permission/Sick radio buttons
- [ ] 5.3 File upload for proof
- [ ] 5.4 Notes textarea
- [ ] 5.5 Submit button with GPS coordinates

### Phase 6: Tugas Aktif View (Left Column)
- [ ] 6.1 Task progress stats grid (total, pending, in_progress, completed)
- [ ] 6.2 CSS-based progress bar
- [ ] 6.3 Task list with priority badges (high/medium/low)
- [ ] 6.4 Status badges with color coding
- [ ] 6.5 Deadline display with urgency indicators
- [ ] 6.6 Empty state when all tasks completed

### Phase 7: Right Column - Quick Stats & Charts
- [ ] 7.1 Attendance percentage card
- [ ] 7.2 Tasks done card
- [ ] 7.3 Recent attendance history list
- [ ] 7.4 CSS-based pie chart for task status
- [ ] 7.5 CSS-based bar chart for weekly attendance

### Phase 8: Late Reason Modal
- [ ] 8.1 Modal for late check-in reason
- [ ] 8.2 Reason textarea
- [ ] 8.3 Submit with check-in

### Phase 9: Design & Polish
- [ ] 9.1 Apply consistent color tints (no crazy gradients)
- [ ] 9.2 Ensure mobile responsiveness
- [ ] 9.3 Add loading states
- [ ] 9.4 Add transitions/animations
- [ ] 9.5 Integrate with Layout/Topbar

### Dependencies
- Leaflet (for maps)
- Chart.js NOT needed (CSS-based visualizations)

## Color Scheme (White Theme + Vercel-inspired)
- Background: #f9fafb (light gray)
- Cards: #ffffff
- Accents:
  - Blue/Indigo: #6366f1 (primary actions)
  - Green/Emerald: #10b981 (success, completed)
  - Amber/Yellow: #f59e0b (pending, warning)
  - Violet: #8b5cf6 (submitted)
  - Rose/Red: #f43f5e (urgent, late)
- Text: #111827 (primary), #6b7280 (secondary)
- Border: #e5e7eb

## API Response Structure
```json
{
  "today_attendance": {
    "checked_in": boolean,
    "checked_out": boolean,
    "status": string | null,
    "check_in_time": string | null,
    "distance": number | null
  },
  "task_stats": {
    "total": number,
    "pending": number,
    "in_progress": number,
    "completed": number,
    "percentage": number
  },
  "task_breakdown": {
    "pending": number,
    "in_progress": number,
    "submitted": number,
    "completed": number,
    "revision": number
  },
  "recent_tasks": [
    {
      "id": number,
      "title": string,
      "status": string,
      "priority": string,
      "deadline": string,
      "is_late": boolean
    }
  ],
  "weekly_attendance": [
    { "date": string, "day": string, "status": string }
  ],
  "weekly_attendance_counts": {
    "labels": ["Mon", "Tue", ...],
    "data": [present_count, absent_count, ...],
    "colors": ["#10b981", "#f43f5e", ...]
  },
  "attendance_percentage": number,
  "attendance_history": [
    { "date": string, "status": string, "check_in": string }
  ],
  "office": {
    "latitude": number,
    "longitude": number,
    "radius": number,
    "name": string
  },
  "user": {
    "name": string,
    "role": string
  }
}
```

## File Structure
- frontend/src/pages/Dashboard.svelte (main page)
- frontend/src/components/dashboard/ (optional components)
- backend/internal/handlers/dashboard.go (API updates)
- frontend/src/lib/api.js (API integration)

