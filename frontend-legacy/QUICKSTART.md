# 🚀 Interna Frontend - Quick Start

## What You Have

A complete, production-ready Svelte frontend with:

✅ **Clean Vercel-inspired design** - Geist + Inter fonts, subtle shadows, no gradients
✅ **Authentication** - Login with JWT, 2FA ready
✅ **Geolocation Attendance** - Check-in/out with location verification
✅ **Performance Analytics** - Charts, trends, insights with Chart.js
✅ **Responsive Design** - Works on mobile, tablet, desktop
✅ **Toast Notifications** - User-friendly feedback
✅ **API Client** - Ready to connect to your Go backend

## Quick Setup (3 Steps)

### 1. Install Dependencies

```bash
cd frontend
npm install
```

### 2. Start Backend

In another terminal:
```bash
cd backend
go run cmd/server/main.go
```

### 3. Start Frontend

```bash
npm run dev
```

Open [http://localhost:5173](http://localhost:5173)

## Default Login

Since you haven't created users yet, you'll need to:

1. **Option A**: Use the backend API to create a test user:
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@interna.com",
    "password": "password123",
    "role": "intern",
    "full_name": "Test Intern",
    "student_id": "12345",
    "institution_id": 1,
    "supervisor_id": 1,
    "start_date": "2026-01-01",
    "end_date": "2026-04-01"
  }'
```

2. **Option B**: Create via MySQL directly:
```sql
-- Create a simple user for testing
INSERT INTO users (email, password_hash, role) 
VALUES ('test@interna.com', '$2a$10$...', 'intern');
```

Then login with:
- Email: `test@interna.com`
- Password: `password123`

## Features Overview

### 🏠 Dashboard
- Today's attendance status
- Quick stats
- Quick action buttons

### 📍 Attendance
- **Check In** - Uses your location
- **Check Out** - Verifies you're at office
- Real-time location display
- Late check-in with reason

### 📊 Analytics
- Weekly check-in trends (line chart)
- Hourly patterns (bar chart)
- Performance insights (strengths, concerns, suggestions)
- Overall score (0-100)

### 👤 Profile
- Account information
- 2FA status
- Security settings

### 📋 Tasks (Stub)
- Ready for when you implement Task handler

## File Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── api.js          # Backend API calls
│   │   └── stores.js       # State management
│   ├── components/
│   │   ├── Navbar.svelte   # Navigation
│   │   ├── Toast.svelte    # Notifications
│   │   └── LoadingSpinner.svelte
│   ├── pages/
│   │   ├── Login.svelte    ✅ Complete
│   │   ├── Dashboard.svelte ✅ Complete
│   │   ├── Attendance.svelte ✅ Complete
│   │   ├── Analytics.svelte ✅ Complete
│   │   ├── Tasks.svelte    🚧 Stub
│   │   └── Profile.svelte  ✅ Complete
│   ├── App.svelte          # Main app
│   └── app.css             # Global styles
└── package.json
```

## Design System

### Colors
```css
/* Backgrounds */
bg-white           /* Main backgrounds */
bg-vercel-gray-50  /* Page background */

/* Text */
text-black         /* Headings */
text-vercel-gray-700 /* Body text */
text-vercel-gray-600 /* Muted text */

/* Borders */
border-vercel-gray-200 /* Subtle borders */
```

### Typography
```svelte
<!-- Heading -->
<h1 class="font-geist font-bold">Heading</h1>

<!-- Body -->
<p class="font-inter">Content</p>
```

### Components
```svelte
<!-- Button -->
<button class="btn-primary">Primary</button>
<button class="btn-secondary">Secondary</button>

<!-- Card -->
<div class="card p-6">Content</div>

<!-- Input -->
<input class="input" type="text" />

<!-- Badge -->
<span class="badge badge-success">Success</span>
```

## API Integration

The frontend is already connected to your backend via `api.js`:

```javascript
// Authentication
await api.login(email, password);
await api.logout();

// Attendance
await api.checkIn(latitude, longitude, reason);
await api.checkOut(latitude, longitude);
await api.getTodayAttendance();

// Analytics
await api.getWeeklyTrends(internId);
await api.getCheckInPatterns(internId, days);
await api.getPerformanceInsights(internId);

// Tasks (when implemented)
await api.getTasks();
await api.createTask(taskData);
```

## Customization

### Change API URL

Edit `vite.config.js`:
```javascript
server: {
  proxy: {
    '/api': {
      target: 'http://your-backend-url:8080',
      changeOrigin: true
    }
  }
}
```

### Change Colors

Edit `tailwind.config.js`:
```javascript
colors: {
  vercel: {
    'accent': '#your-color',
    // ...
  }
}
```

### Add New Page

1. Create `src/pages/NewPage.svelte`
2. Import in `src/App.svelte`
3. Add route: `<Route path="/new" component={NewPage} />`
4. Add to navbar in `src/components/Navbar.svelte`

## Geolocation Testing

The attendance feature requires location access:

1. **Development**: Works on `localhost`
2. **Production**: Requires HTTPS

To test:
1. Allow location permissions in browser
2. Check DevTools Console for errors
3. Location appears in Attendance page

Coordinates are sent to backend for validation (must be within 1km of office).

## Next Steps

1. ✅ **Run the app** - `npm run dev`
2. ✅ **Create a test user** - Use API or MySQL
3. ✅ **Login and explore** - Test attendance, analytics
4. 🚧 **Implement Task handler** - Then update Tasks page
5. 🚧 **Implement Leave handler** - Add Leave Request page
6. 🚧 **Add 2FA setup** - Implement in Profile page
7. 🎨 **Customize design** - Tweak colors, add your logo

## Production Build

```bash
npm run build
```

Deploy the `dist/` folder to:
- **Vercel** - `vercel deploy`
- **Netlify** - Drag & drop `dist/`
- **Nginx** - Point root to `dist/`

## Troubleshooting

**Can't login?**
- Check backend is running on :8080
- Check Network tab for errors
- Verify user exists in database

**Location not working?**
- Allow browser location permission
- Use HTTPS (or localhost)
- Check console for geolocation errors

**Styles broken?**
- Run `npm install`
- Clear browser cache
- Check Tailwind config

**Charts not showing?**
- Check analytics data is loading
- Open console for errors
- Verify Chart.js installed

## Tech Stack

- Svelte 4.2
- Vite 5.0
- Tailwind CSS 3.4
- Chart.js 4.4
- svelte-routing 2.0

## Screenshots

### Login
Clean, minimal design with error handling

### Dashboard
Quick overview with attendance status

### Attendance
Geolocation check-in with live location

### Analytics
Beautiful charts showing performance trends

---

**You're all set!** Start the app and explore. 🚀

The design is clean, professional, and Vercel-inspired - exactly what you asked for. No crazy gradients, just subtle shadows and clean typography with Geist and Inter.
