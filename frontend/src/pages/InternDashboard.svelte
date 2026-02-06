<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';
  import Chart from 'chart.js/auto';
  import L from 'leaflet';
  
  // Office Configuration - This should match backend config
  const OFFICE = $state({
    lat: -7.0355,
    lng: 110.4746,
    name: 'PT. DUTA SOLUSI INFORMATIKA',
    maxDistance: 1000 // meters
  });

  // ==================== STATE ====================
  let loading = $state(true);
  let dashboardLoading = $state(false);
  
  // Attendance State
  let todayAttendance = $state(null);
  let gpsStatus = $state('loading'); // loading, ready, error, denied
  let gps = $state({ lat: null, lng: null, error: null });
  let checkingIn = $state(false);
  let checkingOut = $state(false);
  
  // Map State
  let map = $state(null);
  let userMarker = $state(null);
  let officeMarker = $state(null);
  let radiusCircle = $state(null);
  
  // Dashboard Data
  let dashboardData = $state({
    totalTasks: 0,
    pendingTasks: 0,
    completedTasks: 0,
    inProgressTasks: 0,
    taskStats: { pending: 0, in_progress: 0, submitted: 0, completed: 0, revision: 0 },
    taskBreakdown: { pending: 0, in_progress: 0, submitted: 0, completed: 0, revision: 0 },
    recentTasks: [],
    weeklyAttendance: { days: [], counts: [], colors: [] },
    attendancePercentage: 0,
    attendanceHistory: [],
    office: null
  });

  // Modals
  let showLateModal = $state(false);
  let showPermissionModal = $state(false);
  let lateReason = $state('');
  let permissionForm = $state({
    type: 'permission',
    reason: '',
    document: null
  });

  // Charts
  let taskProgressChart = null;
  let weeklyAttendanceChart = null;
  let watchId = null;

  // ==================== DERIVED STATE ====================
  
  // Calculate distance from office
  const distance = $derived(
    gps.lat && gps.lng 
      ? Math.round(getDistance(gps.lat, gps.lng, OFFICE.lat, OFFICE.lng))
      : null
  );

  // Check if user can check in
const canCheckIn = $derived(
  !loading &&
  !checkingIn &&
  gpsStatus === 'ready' &&
  distance !== null &&
  distance <= OFFICE.maxDistance &&
  todayAttendance === null
);

  
  // Check if user can check out
  const canCheckOut = $derived(
    !loading &&
    !checkingOut &&
    todayAttendance !== null &&
    (todayAttendance.checked_in === true || !!todayAttendance.check_in_time) &&
    (todayAttendance.checked_out === false || !todayAttendance.check_out_time) &&
    !['permission', 'sick'].includes(todayAttendance.status)
  );

  // Check if user is late
  const isLate = $derived(() => {
    const now = new Date();
    const hour = now.getHours();
    const minute = now.getMinutes();
    return (hour > 8) || (hour === 8 && minute > 0);
  });

  // Attendance completion status
  const attendanceComplete = $derived(
    todayAttendance !== null && 
    ((todayAttendance.checked_out === true || !!todayAttendance.check_out_time) || ['permission', 'sick'].includes(todayAttendance.status))
  );

  // ==================== GEOLOCATION ====================
  
  // Haversine formula for distance calculation
  function getDistance(lat1, lon1, lat2, lon2) {
    const R = 6371e3; // metres
    const φ1 = lat1 * Math.PI / 180;
    const φ2 = lat2 * Math.PI / 180;
    const Δφ = (lat2 - lat1) * Math.PI / 180;
    const Δλ = (lon2 - lon1) * Math.PI / 180;
    const a = Math.sin(Δφ / 2) * Math.sin(Δφ / 2) +
      Math.cos(φ1) * Math.cos(φ2) *
      Math.sin(Δλ / 2) * Math.sin(Δλ / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c;
  }

  // Initialize geolocation tracking
  function initGeolocation() {
    if (!navigator.geolocation) {
      gpsStatus = 'error';
      gps.error = 'Geolocation not supported';
      return;
    }

    // Get initial position
    navigator.geolocation.getCurrentPosition(
      (position) => {
        gps = {
          lat: position.coords.latitude,
          lng: position.coords.longitude,
          error: null
        };
        gpsStatus = 'ready';
        updateMapPosition();
      },
      (error) => {
        console.error('Geolocation error:', error);
        gpsStatus = error.code === 1 ? 'denied' : 'error';
        gps.error = error.message;
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 }
    );

    // Watch position for continuous updates
    watchId = navigator.geolocation.watchPosition(
      (position) => {
        gps = {
          lat: position.coords.latitude,
          lng: position.coords.longitude,
          error: null
        };
        if (gpsStatus !== 'ready') gpsStatus = 'ready';
        updateMapPosition();
      },
      (error) => {
        console.error('Watch position error:', error);
      },
      { enableHighAccuracy: true, maximumAge: 30000 }
    );
  }

  // Update user marker position on map
  function updateMapPosition() {
    if (!map || !gps.lat || !gps.lng) return;

    /** @type {[number, number]} */
    const userPos = [gps.lat, gps.lng];

    if (userMarker) {
      userMarker.setLatLng(userPos);
    } else {
      userMarker = L.marker(userPos, {
        icon: L.divIcon({
          className: 'user-marker',
          html: '<div style="background:#10b981; width:24px; height:24px; border-radius:50%; border:3px solid white; box-shadow:0 2px 8px rgba(0,0,0,0.3); animation: pulse 2s infinite;"></div>',
          iconSize: [24, 24]
        })
      }).addTo(map);
    }

    // Center map on user if far from view
    const bounds = map.getBounds();
    if (!bounds.contains(userPos)) {
      map.setView(userPos, 15);
    }
  }

  // ==================== MAP INITIALIZATION ====================
  
  function initMap() {
    if (typeof window === 'undefined') return;
    
    if (map) {
      map.remove();
    }

    const mapElement = document.getElementById('map');
    if (!mapElement) return;

    // Initialize map centered on office
    map = L.map('map').setView([OFFICE.lat, OFFICE.lng], 15);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap contributors'
    }).addTo(map);

    // Office marker
    officeMarker = L.marker([OFFICE.lat, OFFICE.lng], {
      icon: L.divIcon({
        className: 'office-marker',
        html: '<div style="background:#6366f1; width:32px; height:32px; border-radius:50%; border:4px solid white; box-shadow:0 2px 8px rgba(0,0,0,0.3);"></div>',
        iconSize: [32, 32]
      })
    }).addTo(map).bindPopup(OFFICE.name);

    // Office radius circle
    radiusCircle = L.circle([OFFICE.lat, OFFICE.lng], {
      color: '#6366f1',
      fillColor: '#6366f1',
      fillOpacity: 0.1,
      radius: OFFICE.maxDistance
    }).addTo(map);

    setTimeout(() => map.invalidateSize(), 100);
  }

  // ==================== DATA FETCHING ====================
  
  async function fetchDashboardData() {
    dashboardLoading = true;
    try {
      const [tasksRes, attendanceRes, historyRes] = await Promise.all([
        api.getTasks({ limit: 100 }),
        api.getTodayAttendance(),
        api.getAttendance({ limit: 30 })
      ]);
      
      const tasks = tasksRes.data || [];
    //   const todayData = attendanceRes.data;
      const history = historyRes.data || [];

      // Parse today's attendance (handle array or object, or nested attendance property like in Attendance.svelte)
todayAttendance = attendanceRes.data?.attendance || null;


      // Calculate task stats
      const taskStats = { pending: 0, in_progress: 0, submitted: 0, completed: 0, revision: 0 };
      tasks.forEach(t => {
        if (taskStats[t.status] !== undefined) taskStats[t.status]++;
      });

      // Process weekly attendance
      const weeklyStats = processWeeklyAttendance(history);

      // Calculate attendance percentage
      const presentCount = history.filter(h => ['present', 'late'].includes(h.status)).length;
      const totalDays = history.length || 1;
      const attendancePercentage = Math.round((presentCount / totalDays) * 100);

      dashboardData = {
        totalTasks: tasks.length,
        pendingTasks: taskStats.pending,
        completedTasks: taskStats.completed,
        inProgressTasks: taskStats.in_progress,
        taskStats: taskStats,
        taskBreakdown: taskStats,
        recentTasks: tasks.slice(0, 5),
        weeklyAttendance: weeklyStats,
        attendancePercentage: attendancePercentage,
        attendanceHistory: history,
        office: null
      };

      // Initialize charts after data is loaded
      setTimeout(initCharts, 100);
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
    } finally {
      dashboardLoading = false;
    }
  }

  function processWeeklyAttendance(history) {
    const days = [];
    const counts = [];
    const colors = [];
    const today = new Date();

    for (let i = 6; i >= 0; i--) {
      const d = new Date();
      d.setDate(today.getDate() - i);
      const dateStr = d.toISOString().split('T')[0];
      const dayName = d.toLocaleDateString('id-ID', { weekday: 'short' });
      
      days.push(dayName);
      
      const record = history.find(h => {
        const hDate = h.date || (h.check_in_time ? h.check_in_time.split('T')[0] : '');
        return hDate === dateStr;
      });

      if (record) {
        counts.push(1);
        if (record.status === 'present') colors.push('#10b981');
        else if (record.status === 'late') colors.push('#f59e0b');
        else if (record.status === 'sick') colors.push('#6366f1');
        else if (record.status === 'permission') colors.push('#8b5cf6');
        else colors.push('#ef4444');
      } else {
        counts.push(0);
        colors.push('#e2e8f0');
      }
    }
    return { days, counts, colors };
  }

  // ==================== CHART INITIALIZATION ====================
  
  function initCharts() {
    // Destroy existing charts
    if (taskProgressChart) taskProgressChart.destroy();
    if (weeklyAttendanceChart) weeklyAttendanceChart.destroy();

    // Task Progress Pie Chart
    const taskCanvas = document.getElementById('taskProgressChart');
    if (taskCanvas instanceof HTMLCanvasElement && dashboardData.taskBreakdown) {
      const ctx = taskCanvas.getContext('2d');
      const breakdown = dashboardData.taskBreakdown;
      
      taskProgressChart = new Chart(ctx, {
        type: 'doughnut',
        data: {
          labels: ['Pending', 'In Progress', 'Submitted', 'Completed', 'Revision'],
          datasets: [{
            data: [
              breakdown.pending || 0,
              breakdown.in_progress || 0,
              breakdown.submitted || 0,
              breakdown.completed || 0,
              breakdown.revision || 0
            ],
            backgroundColor: [
              '#f59e0b', // amber
              '#6366f1', // indigo
              '#8b5cf6', // violet
              '#10b981', // emerald
              '#ef4444'  // red
            ],
            borderWidth: 0
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: {
              position: 'bottom',
              labels: {
                padding: 12,
                font: { size: 11, weight: 600 },
                usePointStyle: true,
                pointStyle: 'circle'
              }
            }
          }
        }
      });
    }

    // Weekly Attendance Bar Chart
    const weeklyCanvas = document.getElementById('weeklyAttendanceChart');
    if (weeklyCanvas instanceof HTMLCanvasElement && dashboardData.weeklyAttendance?.days?.length) {
      const ctx = weeklyCanvas.getContext('2d');
      const weekly = dashboardData.weeklyAttendance;
      
      weeklyAttendanceChart = new Chart(ctx, {
        type: 'bar',
        data: {
          labels: weekly.days,
          datasets: [{
            label: 'Attendance',
            data: weekly.counts,
            backgroundColor: weekly.colors,
            borderRadius: 6,
            barThickness: 24
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          scales: {
            y: {
              beginAtZero: true,
              ticks: { stepSize: 1, font: { size: 10 } },
              grid: { display: false }
            },
            x: {
              grid: { display: false },
              ticks: { font: { size: 11, weight: 600 } }
            }
          },
          plugins: {
            legend: { display: false }
          }
        }
      });
    }
  }

  // ==================== ATTENDANCE ACTIONS ====================
  
  async function handleCheckIn() {
    if (!canCheckIn) return;

    // Check if late and show modal
    if (isLate() && !showLateModal) {
      showLateModal = true;
      return;
    }

    checkingIn = true;
    try {
      const payload = {
        latitude: gps.lat,
        longitude: gps.lng
      };
      
      if (lateReason && lateReason.trim()) {
        payload.reason = lateReason.trim();
      }

      const res = await api.checkIn(gps.lat, gps.lng, lateReason || null);
      
      if (res.success !== false) {
        // Refresh dashboard data to get updated attendance
        await fetchDashboardData();
        showLateModal = false;
        lateReason = '';
      }
    } catch (err) {
      console.error('Check-in failed:', err);
      alert(err.message || 'Check-in failed');
    } finally {
      checkingIn = false;
    }
  }

  async function handleCheckOut() {
    if (!canCheckOut) return;
    
    if (!confirm('Konfirmasi check-out?')) return;

    checkingOut = true;
    try {
      await api.checkOut(gps.lat || 0, gps.lng || 0);
      await fetchDashboardData();
    } catch (err) {
      console.error('Check-out failed:', err);
      alert(err.message || 'Check-out failed');
    } finally {
      checkingOut = false;
    }
  }

  async function handlePermissionSubmit() {
    if (!permissionForm.reason?.trim()) {
      alert('Please provide a reason');
      return;
    }

    try {
      await api.submitPermission(permissionForm);
      showPermissionModal = false;
      permissionForm = { type: 'permission', reason: '', document: null };
      await fetchDashboardData();
    } catch (err) {
      console.error('Permission submission failed:', err);
      alert(err.message || 'Submission failed');
    }
  }

  function handleFileChange(e) {
    const file = e.target.files?.[0];
    if (file) {
      permissionForm.document = file;
    }
  }

  // ==================== LIFECYCLE ====================
  
  onMount(async () => {
    loading = true;
    
    // Start geolocation
    initGeolocation();
    
    // Fetch dashboard data
    await fetchDashboardData();
    
    loading = false;
    
    // Initialize map after DOM is ready
    setTimeout(() => {
      initMap();
      updateMapPosition();
    }, 100);
  });

  onDestroy(() => {
    // Cleanup
    if (watchId) {
      navigator.geolocation.clearWatch(watchId);
    }
    if (map) {
      map.remove();
    }
    if (taskProgressChart) taskProgressChart.destroy();
    if (weeklyAttendanceChart) weeklyAttendanceChart.destroy();
  });

  // ==================== UTILITIES ====================
  
  function getStatusBadgeClass(status) {
    const classes = {
      pending: 'bg-amber-50 text-amber-700 border-amber-200',
      in_progress: 'bg-indigo-50 text-indigo-700 border-indigo-200',
      submitted: 'bg-violet-50 text-violet-700 border-violet-200',
      completed: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      revision: 'bg-red-50 text-red-700 border-red-200'
    };
    return classes[status] || 'bg-slate-50 text-slate-700 border-slate-200';
  }

  function getStatusLabel(status) {
    const labels = {
      pending: 'Pending',
      in_progress: 'In Progress',
      submitted: 'Submitted',
      completed: 'Completed',
      revision: 'Revision'
    };
    return labels[status] || status;
  }

  function formatDate(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  function formatTime(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
  }
</script>

<svelte:head>
  <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"
        integrity="sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=" crossorigin="" />
  <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20,400,0,0" />
</svelte:head>

<style>


:root {
  --bg: #f9fafb;
  --card: #ffffff;
  --text: #111827;
  --muted: #6b7280;
  --border: #e5e7eb;
  --ring: #d1d5db;
}

  .material-symbols-outlined {
    font-variation-settings:
      'FILL' 0,
      'wght' 200,
      'GRAD' 0,
      'opsz' 20;
  }

  :global(.user-marker), :global(.office-marker) {
    z-index: 1000 !important;
  }
  
  @keyframes pulse {
    0%, 100% { transform: scale(1); opacity: 1; }
    50% { transform: scale(1.1); opacity: 0.8; }
  }
  
  .stat-card, .card {
    background: rgba(255, 255, 255, 0.85) !important;
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }




.action-pill {
  padding: 0.5rem 0.9rem;   /* slimmer */
  border-radius: 9999px;
  border: 1px solid transparent;
  min-height: 2.5rem;       /* reduced thickness */
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.15s ease;
  cursor: pointer;
}

/* Primary action (check-in/out) */
.pill-primary {
  background: #111;
  color: white;
  border-color: #111;
}

.pill-primary:hover:not(:disabled) {
  background: #000;
}

/* Secondary action (izin/sakit) */
.pill-secondary {
  background: #f3f4f6;
  color: #111;
  border-color: var(--border);
  border: 1px solid;
}

.pill-secondary:hover:not(:disabled) {
  background: #e5e7eb;
}

/* Success pill (clock-in info) */
.pill-success {
  background: #ecfdf5;
  color: #065f46;
  border-color: #a7f3d0;
    cursor: not-allowed;

}

.action-pill:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

  .map-reset {
    position: absolute;
    bottom: 1rem;
    right: 1rem;
    z-index: 400;
    background: white;
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 0.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
    transition: all 0.2s;
    cursor: pointer;
  }
  .map-reset:hover {
    background: #f3f4f6;
    transform: scale(1.05);
  }

</style>

<div class="slide-up max-w-[1600px] mx-auto space-y-6">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
    <div>
<h2 class="text-xl font-semibold text-slate-800 tracking-tight">
  Welcome back, {auth.user?.name}
</h2>

    </div>
    <div class="text-right hidden sm:block">
      <div class="text-sm font-semibold text-slate-600">
        {new Date().toLocaleDateString('id-ID', { weekday: 'long', day: '2-digit', month: 'long', year: 'numeric' })}
      </div>
    </div>
  </div>

  {#if loading}
    <div class="text-center py-12">
      <div class="inline-block w-12 h-12 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
      <p class="mt-4 text-slate-600">Loading dashboard...</p>
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- LEFT COLUMN (2/3) -->
      <div class="lg:col-span-2 space-y-6">
        
        <!-- ATTENDANCE CARD -->
        <div class="card p-0 overflow-hidden">
<div class="px-3 py-2 sm:px-6 sm:py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
            <h3 class="font-bold text-lg text-slate-800 flex items-center gap-2">
              <span class="material-symbols-outlined text-indigo-500">map</span> Presensi Harian
            </h3>
            <div class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold"
                 class:bg-emerald-100={gpsStatus === 'ready'}
                 class:text-emerald-700={gpsStatus === 'ready'}
                 class:bg-slate-100={gpsStatus === 'loading'}
                 class:text-slate-500={gpsStatus === 'loading'}
                 class:bg-red-100={gpsStatus === 'error' || gpsStatus === 'denied'}
                 class:text-red-700={gpsStatus === 'error' || gpsStatus === 'denied'}>
              <div class="w-1.5 h-1.5 rounded-full mr-1.5"
                   class:bg-emerald-500={gpsStatus === 'ready'}
                   class:bg-slate-400={gpsStatus === 'loading'}
                   class:bg-red-500={gpsStatus === 'error' || gpsStatus === 'denied'}
                   class:animate-pulse={gpsStatus === 'loading'}></div>
              {#if gpsStatus === 'ready'}GPS Ready
              {:else if gpsStatus === 'loading'}Mencari Lokasi...
              {:else if gpsStatus === 'denied'}GPS Denied
              {:else}GPS Error{/if}
            </div>
          </div>

<div class="p-4 sm:p-5 space-y-4">
            <!-- Map -->
            <div class="relative h-[250px] sm:h-[300px] w-full rounded-xl overflow-hidden border-slate-200 shadow-inner">
              <div id="map" class="h-full w-full z-0"></div>
              <button class="map-reset" onclick={() => map?.setView([OFFICE.lat, OFFICE.lng], 15)} aria-label="Target" title="Target">
                <span class="material-symbols-outlined text-slate-600">gps_fixed</span>
              </button>
            </div>

            <!-- Office Info & Distance -->
<div class="flex flex-col sm:flex-row justify-between items-center bg-slate-50 p-2 rounded-xl border border-slate-100">
  <span class="text-sm font-medium text-slate-600 flex items-center gap-2 whitespace-nowrap">
    <i class="material-symbols-outlined text-slate-400">domain</i>
    Kantor:
    <span class="font-bold text-slate-800">
      {OFFICE.name}
    </span>
  </span>
              <span class="text-sm font-medium text-slate-600">
                Jarak: <span class="font-mono font-bold" 
                            class:text-emerald-600={distance !== null && distance <= OFFICE.maxDistance}
                            class:text-red-600={distance !== null && distance > OFFICE.maxDistance}
                            class:text-slate-800={distance === null}>
                  {distance !== null ? `${distance} m` : '-- m'}
                </span>
              </span>
            </div>

            <!-- Action Buttons -->
            <div>
              {#if attendanceComplete}
                <!-- Attendance Completed -->
                <div class="text-center p-6 bg-emerald-50/50 border border-emerald-100 rounded-xl">
                  <div class="w-16 h-16 bg-emerald-100 text-emerald-600 rounded-full flex items-center justify-center text-3xl mx-auto mb-3 shadow-sm">
                    <span class="material-symbols-outlined text-3xl">check</span>
                  </div>
                  <h4 class="font-bold text-emerald-800 text-lg">Selesai Hari Ini</h4>
                  <p class="text-emerald-600 font-medium text-sm mt-1">
                    Status: <span class="capitalize">{todayAttendance?.status || 'Completed'}</span>
                  </p>
                </div>
              {:else if canCheckOut}
                <!-- Show Check-In Time & Check-Out Button -->
                <div class="flex flex-col sm:flex-row gap-3">

<div class="flex-1 action-pill pill-success text-emerald-700 border-emerald-100 shadow-sm text-sm">
                    <span class="material-symbols-outlined mr-2">schedule</span> 
                    <span>Masuk Pada: <strong class="font-mono text-base">{formatTime(todayAttendance?.check_in_time)}</strong></span>
                  </div>
                  <button
                    onclick={handleCheckOut}
                    disabled={checkingOut}
  class="flex-1 action-pill pill-primary text-sm sm:text-base disabled:opacity-50 disabled:cursor-not-allowed shadow-sm">
                    {#if checkingOut}
                      <span class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
                      Processing...
                    {:else}
                      <span class="material-symbols-outlined mr-2">logout</span> PRESENSI KELUAR
                    {/if}
                  </button>
                </div>
              {:else}
                <div class="flex flex-col sm:flex-row gap-3">

                  <!-- Check-In Button -->
                  <button
                    onclick={handleCheckIn}
                    disabled={!canCheckIn || checkingIn}
  class="flex-1 action-pill pill-primary text-sm sm:text-base disabled:cursor-not-allowed shadow-sm">
                    {#if checkingIn}
                      <span class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
                      Processing...
                    {:else if gpsStatus === 'loading'}
                      <span class="material-symbols-outlined animate-spin mr-2">progress_activity</span> Menunggu GPS...
                    {:else if gpsStatus === 'denied'}
                      <span class="material-symbols-outlined mr-2">warning</span> GPS Ditolak
                    {:else if gpsStatus === 'error'}
                      <span class="material-symbols-outlined mr-2">error</span> GPS Error
                    {:else if distance === null}
                      <span class="material-symbols-outlined mr-2">near_me</span> Menunggu Lokasi...
                    {:else if distance > OFFICE.maxDistance}
                      <span class="material-symbols-outlined mr-2">location_on</span> Terlalu Jauh ({distance}m)
                    {:else}
                      <span class="material-symbols-outlined mr-2">login</span> PRESENSI MASUK
                    {/if}
                  </button>

                  <!-- Permission Button -->
                  <button
                    onclick={() => showPermissionModal = true}
  class="flex-1 action-pill pill-secondary text-sm sm:text-base shadow-sm">
                    <span class="material-symbols-outlined mr-2">sick</span> Izin / Sakit
                  </button>
                </div>
              {/if}
            </div>
          </div>
        </div>

        <!-- TASKS CARD -->
        <div class="card p-0 overflow-hidden">
<div class="px-5 py-5 sm:px-8 sm:py-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
            <h3 class="font-bold text-lg text-slate-800 flex items-center gap-2">
              <span class="material-symbols-outlined text-teal-500">assignment</span> Tugas Saya
            </h3>
            <a href="/tasks" class="text-xs font-bold text-teal-600 hover:text-teal-700 uppercase tracking-wider hover:underline">
              Lihat Semua
            </a>
          </div>

          <div class="p-4 sm:p-6 space-y-6">
            <!-- Task Stats -->
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
              <div class="bg-slate-50 p-3 sm:p-4 rounded-xl border border-slate-100 text-center hover:bg-white hover:shadow-md transition-all">
                <div class="text-2xl font-black text-slate-700">{dashboardData.totalTasks}</div>
                <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider mt-1">Total</div>
              </div>
              <div class="bg-amber-50 p-3 sm:p-4 rounded-xl border border-amber-100 text-center hover:bg-white hover:shadow-md transition-all">
                <div class="text-2xl font-black text-amber-600">{dashboardData.pendingTasks}</div>
                <div class="text-[10px] font-bold text-amber-600/70 uppercase tracking-wider mt-1">Pending</div>
              </div>
              <div class="bg-indigo-50 p-3 sm:p-4 rounded-xl border border-indigo-100 text-center hover:bg-white hover:shadow-md transition-all">
                <div class="text-2xl font-black text-indigo-600">{dashboardData.inProgressTasks}</div>
                <div class="text-[10px] font-bold text-indigo-600/70 uppercase tracking-wider mt-1">Proses</div>
              </div>
              <div class="bg-emerald-50 p-3 sm:p-4 rounded-xl border border-emerald-100 text-center hover:bg-white hover:shadow-md transition-all">
                <div class="text-2xl font-black text-emerald-600">{dashboardData.completedTasks}</div>
                <div class="text-[10px] font-bold text-emerald-600/70 uppercase tracking-wider mt-1">Selesai</div>
              </div>
            </div>

            <!-- Progress Bar -->
            {#if dashboardData.totalTasks > 0}
              <div>
                <div class="flex justify-between items-center mb-2">
                  <span class="text-xs font-bold text-slate-600 uppercase tracking-wider">Progress</span>
                  <span class="text-sm font-black text-slate-800">
                    {Math.round((dashboardData.completedTasks / dashboardData.totalTasks) * 100)}%
                  </span>
                </div>
                <div class="h-3 bg-slate-100 rounded-full overflow-hidden">
                  <div class="h-full bg-gradient-to-r from-emerald-500 to-teal-500 rounded-full transition-all duration-500"
                       style="width: {(dashboardData.completedTasks / dashboardData.totalTasks) * 100}%"></div>
                </div>
              </div>
            {/if}

            <!-- Recent Tasks -->
            {#if dashboardData.recentTasks?.length > 0}
              <div class="space-y-3">
                <h4 class="text-xs font-bold text-slate-500 uppercase tracking-wider">Tugas Terbaru</h4>
                {#each dashboardData.recentTasks as task}
                  <a href="/tasks/{task.id}" class="block p-3 bg-slate-50 rounded-lg border border-slate-200 hover:shadow-sm hover:bg-white transition-all">
                    <div class="flex justify-between items-start mb-2">
                      <h5 class="font-semibold text-slate-800 text-sm">{task.title}</h5>
                      <span class="px-2 py-0.5 rounded text-[10px] font-bold border {getStatusBadgeClass(task.status)}">
                        {getStatusLabel(task.status)}
                      </span>
                    </div>
                    <div class="flex items-center gap-3 text-xs text-slate-500">
                      {#if task.deadline}
                        <span class="flex items-center"><span class="material-symbols-outlined text-[16px] mr-1">calendar_today</span>{formatDate(task.deadline)}</span>
                      {/if}
                      {#if task.priority}
                        <span class="capitalize flex items-center">
                          <span class="material-symbols-outlined text-[16px] mr-1" 
                             class:text-red-500={task.priority === 'high'}
                             class:text-amber-500={task.priority === 'medium'}
                             class:text-slate-400={task.priority === 'low'}>flag</span>
                          {task.priority}
                        </span>
                      {/if}
                    </div>
                  </a>
                {/each}
              </div>
            {:else}
              <div class="text-center py-8 text-slate-400">
                <span class="material-symbols-outlined text-4xl mb-2">inbox</span>
                <p class="text-sm">Tidak ada tugas</p>
              </div>
            {/if}
          </div>
        </div>
      </div>

      <!-- RIGHT COLUMN (1/3) -->
      <div class="space-y-6">
        
        <!-- Task Breakdown Chart -->
        <div class="card p-4 sm:p-6">
          <h3 class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2">
            <span class="material-symbols-outlined text-violet-500">pie_chart</span> Status Tugas
          </h3>
          <div class="h-[200px]">
            <canvas id="taskProgressChart"></canvas>
          </div>
        </div>

        <!-- Weekly Attendance -->
        <div class="card p-4 sm:p-6">
          <h3 class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2">
            <span class="material-symbols-outlined text-indigo-500">date_range</span> Presensi Minggu Ini
          </h3>
          <div class="h-[200px]">
            <canvas id="weeklyAttendanceChart"></canvas>
          </div>
        </div>

        <!-- Attendance Percentage -->
        <div class="card p-4 sm:p-6">
          <h3 class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2">
            <span class="material-symbols-outlined text-emerald-500">percent</span> Kehadiran (30 Hari)
          </h3>
          <div class="text-center">
            <div class="text-5xl font-black text-slate-800 mb-2">
              {dashboardData.attendancePercentage}%
            </div>
            <div class="h-2 bg-slate-100 rounded-full overflow-hidden">
              <div class="h-full bg-gradient-to-r from-emerald-500 to-teal-500 rounded-full transition-all duration-500"
                   style="width: {dashboardData.attendancePercentage}%"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>

<!-- Late Reason Modal -->
{#if showLateModal}
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onclick={(e) => e.target === e.currentTarget && (showLateModal = false)}>
    <div class="bg-white rounded-2xl p-6 max-w-md w-full shadow-2xl">
      <h3 class="text-xl font-bold text-slate-800 mb-4">Alasan Terlambat</h3>
      <p class="text-sm text-slate-600 mb-4">Anda terlambat. Mohon berikan alasan:</p>
      <textarea
        bind:value={lateReason}
        placeholder="Tulis alasan..."
        class="w-full px-4 py-3 border border-slate-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
        rows="3"></textarea>
      <div class="flex gap-3 mt-4">
        <button onclick={() => showLateModal = false} class="btn flex-1 bg-slate-100 text-slate-700 hover:bg-slate-200">
          Batal
        </button>
        <button onclick={handleCheckIn} class="btn flex-1 bg-indigo-600 text-white hover:bg-indigo-700">
          Lanjutkan Check-In
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Permission Modal -->
{#if showPermissionModal}
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onclick={(e) => e.target === e.currentTarget && (showPermissionModal = false)}>
    <div class="bg-white rounded-2xl p-6 max-w-md w-full shadow-2xl">
      <h3 class="text-xl font-bold text-slate-800 mb-4">Ajukan Izin/Sakit</h3>
      
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-semibold text-slate-700 mb-2">Jenis</label>
          <select bind:value={permissionForm.type} class="w-full px-4 py-2.5 border border-slate-200 rounded-lg focus:ring-2 focus:ring-indigo-500">
            <option value="permission">Izin</option>
            <option value="sick">Sakit</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-semibold text-slate-700 mb-2">Alasan</label>
          <textarea
            bind:value={permissionForm.reason}
            placeholder="Jelaskan alasan..."
            class="w-full px-4 py-3 border border-slate-200 rounded-lg focus:ring-2 focus:ring-indigo-500 resize-none"
            rows="3"></textarea>
        </div>

        <div>
          <label class="block text-sm font-semibold text-slate-700 mb-2">Dokumen Pendukung (Opsional)</label>
          <input
            type="file"
            accept="image/*,.pdf"
            onchange={handleFileChange}
            class="w-full px-4 py-2.5 border border-slate-200 rounded-lg focus:ring-2 focus:ring-indigo-500" />
        </div>
      </div>

      <div class="flex gap-3 mt-6">
        <button onclick={() => { showPermissionModal = false; permissionForm = { type: 'permission', reason: '', document: null }; }} 
                class="btn flex-1 bg-slate-100 text-slate-700 hover:bg-slate-200">
          Batal
        </button>
        <button onclick={handlePermissionSubmit} class="btn flex-1 bg-indigo-600 text-white hover:bg-indigo-700">
          Kirim
        </button>
      </div>
    </div>
  </div>
{/if}