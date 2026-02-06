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
    todayAttendance.checked_in === true &&
    todayAttendance.checked_out === false &&
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
    (todayAttendance.checked_out === true || ['permission', 'sick'].includes(todayAttendance.status))
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
    if (typeof window === 'undefined' || !window.L) return;
    
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
      const res = await api.getInternDashboardData();
      const data = res.data;
      
      // Parse today's attendance
      todayAttendance = data.today_attendance?.checked_in ? data.today_attendance : null;

      // Parse dashboard data
      dashboardData = {
        totalTasks: data.task_stats?.total || 0,
        pendingTasks: data.task_stats?.pending || 0,
        completedTasks: data.task_stats?.completed || 0,
        inProgressTasks: data.task_stats?.in_progress || 0,
        taskStats: data.task_stats || {},
        taskBreakdown: data.task_breakdown || {},
        recentTasks: data.recent_tasks || [],
        weeklyAttendance: {
          days: data.weekly_attendance_counts?.labels || [],
          counts: data.weekly_attendance_counts?.data || [],
          colors: data.weekly_attendance_counts?.colors || []
        },
        attendancePercentage: data.attendance_percentage || 0,
        attendanceHistory: data.attendance_history || [],
        office: data.office || null
      };

      // Override OFFICE config if backend provides it
      if (data.office?.latitude && data.office?.longitude) {
        OFFICE.lat = data.office.latitude;
        OFFICE.lng = data.office.longitude;
        OFFICE.maxDistance = data.office.radius || 1000;
        OFFICE.name = data.office.name || OFFICE.name;
      }

      // Initialize charts after data is loaded
      setTimeout(initCharts, 100);
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
    } finally {
      dashboardLoading = false;
    }
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
    
    // Initialize map first
    initMap();
    
    // Start geolocation
    initGeolocation();
    
    // Fetch dashboard data
    await fetchDashboardData();
    
    loading = false;
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
</script>

<svelte:head>
  <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"
        integrity="sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=" crossorigin="" />
</svelte:head>

<style>
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
</style>

<div class="slide-up max-w-[1600px] mx-auto space-y-6">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
    <div>
      <h2 class="text-2xl sm:text-3xl font-bold text-slate-800 tracking-tight mb-1">
        Selamat datang kembali, {auth.user?.name || 'User'}!
      </h2>
      <p class="text-slate-500 font-medium text-sm sm:text-base">
        Berikut ringkasan aktivitas magang Anda.
      </p>
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
          <div class="p-4 sm:p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
            <h3 class="font-bold text-lg text-slate-800 flex items-center gap-2">
              <i class="fas fa-map-marked-alt text-indigo-500"></i> Presensi Harian
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

          <div class="p-4 sm:p-6 space-y-4">
            <!-- Map -->
            <div class="relative h-[220px] sm:h-[300px] w-full rounded-xl overflow-hidden border border-slate-200 shadow-inner">
              <div id="map" class="h-full w-full z-0"></div>
            </div>

            <!-- Office Info & Distance -->
            <div class="flex flex-col sm:flex-row justify-between items-center gap-4 bg-slate-50 p-4 rounded-xl border border-slate-100">
              <span class="text-sm font-medium text-slate-600 flex items-center gap-2">
                <i class="fas fa-building text-slate-400"></i> Kantor: 
                <span class="font-bold text-slate-800">{OFFICE.name}</span>
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
                    <i class="fas fa-check"></i>
                  </div>
                  <h4 class="font-bold text-emerald-800 text-lg">Selesai Hari Ini</h4>
                  <p class="text-emerald-600 font-medium text-sm mt-1">
                    Status: <span class="capitalize">{todayAttendance?.status || 'Completed'}</span>
                  </p>
                </div>
              {:else if canCheckOut}
                <!-- Show Check-In Time & Check-Out Button -->
                <div class="text-center mb-4">
                  <div class="inline-flex items-center px-4 py-2 rounded-lg bg-emerald-50 text-emerald-700 font-medium text-sm border border-emerald-100">
                    <i class="fas fa-check-circle mr-2"></i> Anda masuk pukul 
                    <strong class="ml-1">{todayAttendance?.check_in_time || 'N/A'}</strong>
                  </div>
                </div>
                <button
                  onclick={handleCheckOut}
                  disabled={checkingOut}
                  class="btn w-full py-3.5 text-base font-bold bg-amber-500 text-white hover:bg-amber-600 shadow-lg shadow-amber-500/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed">
                  {#if checkingOut}
                    <span class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
                    Processing...
                  {:else}
                    <i class="fas fa-sign-out-alt mr-2"></i> PRESENSI KELUAR SEKARANG
                  {/if}
                </button>
              {:else}
                <!-- Check-In Button -->
                <button
                  onclick={handleCheckIn}
                  disabled={!canCheckIn || checkingIn}
                  class="btn w-full py-3.5 text-base font-bold shadow-lg shadow-indigo-500/20 transition-all"
                  class:bg-indigo-600={canCheckIn}
                  class:text-white={canCheckIn}
                  class:hover:bg-indigo-700={canCheckIn}
                  class:bg-slate-300={!canCheckIn}
                  class:text-slate-500={!canCheckIn}
                  class:cursor-not-allowed={!canCheckIn}
                  class:opacity-50={!canCheckIn || checkingIn}>
                  {#if checkingIn}
                    <span class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
                    Processing...
                  {:else if gpsStatus === 'loading'}
                    <i class="fas fa-spinner fa-spin mr-2"></i> Menunggu GPS...
                  {:else if gpsStatus === 'denied'}
                    <i class="fas fa-exclamation-triangle mr-2"></i> GPS Ditolak
                  {:else if gpsStatus === 'error'}
                    <i class="fas fa-exclamation-circle mr-2"></i> GPS Error
                  {:else if distance === null}
                    <i class="fas fa-location-arrow mr-2"></i> Menunggu Lokasi...
                  {:else if distance > OFFICE.maxDistance}
                    <i class="fas fa-map-marker-alt mr-2"></i> Terlalu Jauh ({distance}m)
                  {:else}
                    <i class="fas fa-sign-in-alt mr-2"></i> PRESENSI MASUK SEKARANG
                  {/if}
                </button>

                <!-- Permission Button -->
                <button
                  onclick={() => showPermissionModal = true}
                  class="btn w-full mt-3 py-2.5 text-sm font-semibold bg-slate-100 text-slate-700 hover:bg-slate-200 border border-slate-200">
                  <i class="fas fa-file-medical mr-2"></i> Ajukan Izin/Sakit
                </button>
              {/if}
            </div>
          </div>
        </div>

        <!-- TASKS CARD -->
        <div class="card p-0 overflow-hidden">
          <div class="p-4 sm:p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
            <h3 class="font-bold text-lg text-slate-800 flex items-center gap-2">
              <i class="fas fa-clipboard-list text-teal-500"></i> Tugas Saya
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
                  <div class="p-3 bg-slate-50 rounded-lg border border-slate-100 hover:shadow-sm transition-all">
                    <div class="flex justify-between items-start mb-2">
                      <h5 class="font-semibold text-slate-800 text-sm">{task.title}</h5>
                      <span class="px-2 py-0.5 rounded text-[10px] font-bold border {getStatusBadgeClass(task.status)}">
                        {getStatusLabel(task.status)}
                      </span>
                    </div>
                    <div class="flex items-center gap-3 text-xs text-slate-500">
                      {#if task.deadline}
                        <span><i class="fas fa-calendar-alt mr-1"></i>{formatDate(task.deadline)}</span>
                      {/if}
                      {#if task.priority}
                        <span class="capitalize">
                          <i class="fas fa-flag mr-1" 
                             class:text-red-500={task.priority === 'high'}
                             class:text-amber-500={task.priority === 'medium'}
                             class:text-slate-400={task.priority === 'low'}></i>
                          {task.priority}
                        </span>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="text-center py-8 text-slate-400">
                <i class="fas fa-inbox text-3xl mb-2"></i>
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
            <i class="fas fa-chart-pie text-violet-500"></i> Status Tugas
          </h3>
          <div class="h-[200px]">
            <canvas id="taskProgressChart"></canvas>
          </div>
        </div>

        <!-- Weekly Attendance -->
        <div class="card p-4 sm:p-6">
          <h3 class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2">
            <i class="fas fa-calendar-week text-indigo-500"></i> Presensi Minggu Ini
          </h3>
          <div class="h-[200px]">
            <canvas id="weeklyAttendanceChart"></canvas>
          </div>
        </div>

        <!-- Attendance Percentage -->
        <div class="card p-4 sm:p-6">
          <h3 class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2">
            <i class="fas fa-percent text-emerald-500"></i> Kehadiran (30 Hari)
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