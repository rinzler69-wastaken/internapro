import { render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import InternDashboard from '../InternDashboard.svelte';

// Mock Leaflet and Chart.js that rely on DOM APIs not available in jsdom
vi.mock('leaflet', () => ({
  map: vi.fn(() => ({
    setView: vi.fn().mockReturnThis(),
    remove: vi.fn(),
    invalidateSize: vi.fn(),
    _container: {}
  })),
  tileLayer: vi.fn(() => ({ addTo: vi.fn() })),
  marker: vi.fn(() => ({ addTo: vi.fn(), setLatLng: vi.fn(), bindPopup: vi.fn() })),
  circle: vi.fn(() => ({ addTo: vi.fn() })),
  divIcon: vi.fn(() => ({})),
}));

vi.mock('chart.js/auto', () => ({ default: class Chart { destroy() {} } }));

// Mock API responses
vi.mock('../lib/api.js', () => ({
  api: {
    getTasks: vi.fn(async () => ({ data: [] })),
    getTodayAttendance: vi.fn(async () => ({ data: null })),
    getAttendance: vi.fn(async () => ({ data: [] })),
  },
}));

// Mock auth store
vi.mock('../lib/auth.svelte.js', () => ({
  auth: {
    user: { name: 'Test Intern' },
    token: 'fake-token',
  },
}));

// Helper: mock geolocation
beforeAll(() => {
  // Provide a controllable geolocation implementation
  global.navigator.geolocation = {
    getCurrentPosition: (success) => {
      success({ coords: { latitude: -7.03, longitude: 110.47 } });
    },
    watchPosition: (success) => {
      success({ coords: { latitude: -7.03, longitude: 110.47 } });
      return 1; // fake watch id
    },
    clearWatch: vi.fn(),
  };
});

describe('InternDashboard', () => {
  test('renders welcome header with user name and tabs', async () => {
    render(InternDashboard);

    expect(await screen.findByText(/Selamat Datang,/i)).toBeInTheDocument();
    expect(screen.getByText('Test Intern')).toBeInTheDocument();

    // Buttons for mode switching exist
    expect(screen.getByRole('button', { name: /Presensi/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Tugas/i })).toBeInTheDocument();
  });

  test('shows attendance card with action buttons', async () => {
    render(InternDashboard);

    // Wait until loading spinner is gone
    await waitFor(() => expect(screen.queryByText(/Loading dashboard/i)).not.toBeInTheDocument());

    expect(screen.getByText(/Presensi Harian/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /PRESENSI MASUK/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Izin \/ Sakit/i })).toBeInTheDocument();
  });

  test('switches to tugas view when tab clicked', async () => {
    render(InternDashboard);

    const tugasTab = screen.getByRole('button', { name: /Tugas/i });
    await userEvent.click(tugasTab);

    // Wait for re-render
    await waitFor(() => {
      expect(screen.getByText(/Tugas Aktif/i)).toBeInTheDocument();
    });
  });
});
