# Interna Frontend

Modern, clean frontend for the Interna internship management system built with Svelte and Vite.

## Design System

- **Fonts**: Geist (headings), Inter (content)
- **Styling**: Tailwind CSS with Vercel-inspired design
- **Colors**: Clean blacks, grays, subtle shadows
- **No gradients**: Simple, professional aesthetic

## Features

- 🔐 **Authentication** - JWT-based login with 2FA support
- 📍 **Geolocation Attendance** - Check-in/out with location verification
- 📊 **Analytics Dashboard** - Performance trends and insights
- 📋 **Task Management** - View and manage internship tasks (coming soon)
- 👤 **Profile Management** - Account settings and security

## Tech Stack

- **Framework**: Svelte 4
- **Build Tool**: Vite 5
- **Styling**: Tailwind CSS 3
- **Routing**: svelte-routing
- **Charts**: Chart.js 4
- **Icons**: Heroicons (via SVG)

## Setup

### Install Dependencies

```bash
npm install
```

### Environment

The frontend proxies API requests to `http://localhost:8080` by default (configured in `vite.config.js`).

Make sure your backend is running on port 8080, or update the proxy configuration.

### Development

```bash
npm run dev
```

Open [http://localhost:5173](http://localhost:5173) in your browser.

### Build for Production

```bash
npm run build
```

The built files will be in the `dist/` directory.

### Preview Production Build

```bash
npm run preview
```

## Project Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── api.js           # API client
│   │   └── stores.js        # Svelte stores
│   ├── components/
│   │   ├── Navbar.svelte    # Navigation bar
│   │   ├── Toast.svelte     # Toast notifications
│   │   └── LoadingSpinner.svelte
│   ├── pages/
│   │   ├── Login.svelte     # Login page
│   │   ├── Dashboard.svelte # Main dashboard
│   │   ├── Attendance.svelte # Check-in/out
│   │   ├── Tasks.svelte     # Task management
│   │   ├── Analytics.svelte # Performance analytics
│   │   └── Profile.svelte   # User profile
│   ├── App.svelte           # Main app component
│   ├── main.js              # Entry point
│   └── app.css              # Global styles
├── index.html
├── vite.config.js
├── tailwind.config.js
└── package.json
```

## Key Components

### API Client (`src/lib/api.js`)

Centralized API communication with automatic token management:

```javascript
import { api } from './lib/api';

// Login
await api.login(email, password);

// Check in
await api.checkIn(latitude, longitude);

// Get analytics
const trends = await api.getWeeklyTrends(internId);
```

### Stores (`src/lib/stores.js`)

Global state management:

- `auth` - Authentication state
- `toast` - Toast notifications
- `location` - Geolocation data
- `todayAttendance` - Today's attendance status

### Pages

#### Login
- Clean, minimal design
- Email/password authentication
- Error handling

#### Dashboard
- Quick stats overview
- Today's attendance status
- Quick action links

#### Attendance
- Geolocation-based check-in/out
- Live location display
- Attendance status

#### Analytics
- Weekly trends chart
- Check-in patterns
- Performance insights
- Overall score

## Styling Guide

### Typography

```svelte
<!-- Headings -->
<h1 class="font-geist">Large Heading</h1>

<!-- Body text -->
<p class="font-inter">Regular content text</p>
```

### Buttons

```svelte
<!-- Primary button -->
<button class="btn-primary">Click Me</button>

<!-- Secondary button -->
<button class="btn-secondary">Cancel</button>
```

### Cards

```svelte
<div class="card p-6">
  Card content
</div>
```

### Inputs

```svelte
<label class="label">Field Name</label>
<input class="input" type="text" />
```

### Badges

```svelte
<span class="badge badge-success">Success</span>
<span class="badge badge-warning">Warning</span>
<span class="badge badge-error">Error</span>
```

## Features by Page

### ✅ Implemented

- Login with email/password
- Dashboard with attendance overview
- Geolocation-based attendance check-in/out
- Performance analytics with charts
- User profile display

### 🚧 Coming Soon (Backend Required)

- Task management
- Leave request management
- Assessment viewing
- 2FA setup
- Full profile editing

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)

Geolocation requires HTTPS in production.

## Development Tips

### Adding a New Page

1. Create component in `src/pages/`
2. Import in `App.svelte`
3. Add route in Router
4. Add navigation link in `Navbar.svelte`

### Making API Calls

Always use the `api` client:

```javascript
import { api } from '../lib/api';

try {
  const response = await api.getTasks();
  // Handle success
} catch (error) {
  // Handle error
  toast.add(error.message, 'error');
}
```

### Showing Notifications

```javascript
import { toast } from '../lib/stores';

toast.add('Success message', 'success');
toast.add('Error message', 'error');
toast.add('Info message', 'info');
```

## Deployment

### Build

```bash
npm run build
```

### Serve

You can serve the `dist/` folder with any static server:

```bash
# Using Python
python -m http.server -d dist 8000

# Using Node.js serve
npx serve dist

# Using nginx (production)
# Point nginx root to dist/
```

### Environment Variables

For production, update the API proxy in `vite.config.js` or use environment variables:

```javascript
proxy: {
  '/api': {
    target: import.meta.env.VITE_API_URL || 'http://localhost:8080',
    changeOrigin: true
  }
}
```

## Troubleshooting

**Geolocation not working**
- Enable location permissions in browser
- Use HTTPS (required for geolocation)
- Check browser console for errors

**API calls failing**
- Ensure backend is running on port 8080
- Check Network tab in DevTools
- Verify JWT token is being sent

**Styles not loading**
- Run `npm install` to ensure Tailwind is installed
- Check `tailwind.config.js` paths
- Clear browser cache

## License

Internal use - Internship project
