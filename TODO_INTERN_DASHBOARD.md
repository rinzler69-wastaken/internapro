# Intern Dashboard Reverse Engineering Plan

## Overview
Reverse engineer the Laravel intern dashboard to match Vercel-inspired design with CSS-based visualizations and GPS attendance.

## Tasks

### Phase 1: Data Layer & API Updates
- [ ] 1.1 Update API to include weekly attendance data
- [ ] 1.2 Update API to include task breakdown stats
- [ ] 1.3 Update API to include attendance percentage
- [ ] 1.4 Add office location config endpoint

### Phase 2: Dashboard UI Components
- [ ] 2.1 Welcome header with date display
- [ ] 2.2 Attendance card with GPS map (Leaflet)
- [ ] 2.3 Check-in/out with geolocation logic
- [ ] 2.4 Permission/sick leave modal
- [ ] 2.5 Late reason modal

### Phase 3: Task Section
- [ ] 3.1 Task progress stats grid
- [ ] 3.2 Progress bar component
- [ ] 3.3 Task list with priority/status/deadline badges
- [ ] 3.4 CSS-based pie chart for task status

### Phase 4: Charts & Stats
- [ ] 4.1 Weekly attendance bar chart (CSS-based)
- [ ] 4.2 Quick stats summary
- [ ] 4.3 Attendance history list

### Phase 5: Design & Polish
- [ ] 5.1 Vercel-style color accents
- [ ] 5.2 Mobile responsiveness
- [ ] 5.3 Loading states & transitions

## Design System
- Background: `#000000` / `#111111` (dark mode)
- Cards: `#111111` / `#1A1A1A`
- Accents:
  - Blue: `#60A5FA` (attendance)
  - Green: `#34D399` (tasks done)
  - Amber: `#FBBF24` (pending)
  - Violet: `#A78BFA` (submitted)
- Border: `#333333`
- Text: `#EDEDED` (primary), `#A1A1AA` (secondary)

## CSS-Based Visualizations
- Pie Chart: conic-gradient CSS
- Bar Chart: flexbox with percentage heights
- Donut Chart: SVG or conic-gradient

## Dependencies
- Leaflet (for maps)
- Already installed: Chart.js (but prefer CSS approach)

