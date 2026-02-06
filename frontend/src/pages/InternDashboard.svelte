<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';
  import Chart from 'chart.js/auto';
  import L from 'leaflet';



  
  
  // Office Configuration
  const OFFICE = {
    lat: -7.0355,
    lng: 110.4746,
    name: 'PT. DUTA SOLUSI INFORMATIKA',
    maxDistance: 1000 // meters
  };

  // State
  let loading = $state(true);
  let todayAttendance = $state(null);
  let gpsStatus = $state('loading');
  let gps = $state({ lat: null, lng: null, error: null });
  let map = $state(null);
  let userMarker = $state(null);
  let officeMarker = $state(null);
  
  // Dashboard Data
  let dashboardData = $state({
    totalTasks: 0,
    pendingTasks: 0,
    completedTasks: 0,
    taskStats: { pending: 0, in_progress: 0, submitted: 0, completed: 0, revision: 0 },
    taskBreakdown: { pending: 0, in_progress: 0, submitted: 0, completed: 0, revision: 0 },
    recentTasks: [],
    weeklyAttendance: { days: [], counts: [], colors: [] },
    attendancePercentage: 0
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

  // Check-in state
  let checkingIn = $state(false);

  // Charts
  let taskProgressChart = null;
  let weeklyAttendanceChart = null;

  // Calculated distance
  const distance = $derived(
    gps.lat && gps.lng 
      ? Math.round(getDistance(gps.lat, gps.lng, OFFICE.lat, OFFICE.lng))
      : null
  );

// 2. Update canCheckIn to be strictly dependent on data existence
const canCheckIn = $derived(
  !loading &&
  gpsStatus === 'ready' &&
  distance !== null &&
  distance <= OFFICE.maxDistance &&
  todayAttendance === null // Strictly null
);

  const isLate = $derived(() => {
    const now = new Date();
    const hour = now.getHours();
    const minute = now.getMinutes();
    return (hour > 8) || (hour === 8 && minute > 0); // Late after 08:00
  });

  
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

  // Fetch dashboard data
  async function fetchDashboardData() {
    loading = true;
    try {
      const res = await api.getInternDashboard();
      const data = res.data;
      
      todayAttendance = data.today_attendance || null;

      dashboardData = {
        totalTasks: data.task_stats.total,
        pendingTasks: data.task_stats.pending,
        completedTasks: data.task_stats.completed,
        taskStats: data.task_breakdown,
        taskBreakdown: data.task_breakdown,
        recentTasks: data.recent_tasks || [],
        weeklyAttendance: {
          days: data.weekly_attendance_counts.labels,
          counts: data.weekly_attendance_counts.data,
          colors: data.weekly_attendance_counts.colors
        },
        attendancePercentage: data.attendance_percentage
      };

      // Initialize charts
      setTimeout(initCharts, 100);
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
    } finally {
      loading = false;
    }
  }

  // Initialize Leaflet map
  function initMap() {
    if (typeof window === 'undefined' || !window.L) return;
    
    if (map) {
      map.remove();
    }

    const mapElement = document.getElementById('map');
    if (!mapElement) return;

    map = window.L.map('map').setView([OFFICE.lat, OFFICE.lng], 15);

    window.L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap contributors'
    }).addTo(map);

    // Office marker
    officeMarker = window.L.marker([OFFICE.lat, OFFICE.lng], {
      icon: window.L.divIcon({
        className: 'office-marker',
        html: '<div style="background:#6366f1; width:32px; height:32px; border-radius:50%; border:4px solid white; box-shadow:0 2px 8px rgba(0,0,0,0.3);"></div>',
        iconSize: [32, 32]
      })
    }).addTo(map).bindPopup(OFFICE.name);

    // Circle showing check-in range
    window.L.circle([OFFICE.lat, OFFICE.lng], {
      color: '#6366f1',
      fillColor: '#6366f1',
      fillOpacity: 0.1,
      radius: OFFICE.maxDistance
    }).addTo(map);
  }

  // Update user location on map
  function updateUserLocation() {
    if (!map || !gps.lat || !gps.lng) return;

    if (userMarker) {
      userMarker.setLatLng([gps.lat, gps.lng]);
    } else {
      userMarker = window.L.marker([gps.lat, gps.lng], {
        icon: window.L.divIcon({
          className: 'user-marker',
          html: '<div style="background:#10b981; width:20px; height:20px; border-radius:50%; border:3px solid white; box-shadow:0 2px 8px rgba(0,0,0,0.3);"></div>',
          iconSize: [20, 20]
        })
      }).addTo(map).bindPopup('Lokasi Anda');
    }

    map.setView([gps.lat, gps.lng], 16);
  }

  // Initialize GPS tracking
  function initGPS() {
    if (!navigator.geolocation) {
      gpsStatus = 'error';
      gps.error = 'GPS tidak didukung di browser ini';
      return;
    }

    gpsStatus = 'loading';
    
    navigator.geolocation.watchPosition(
      (position) => {
        gps.lat = position.coords.latitude;
        gps.lng = position.coords.longitude;
        gps.error = null;
        gpsStatus = 'ready';
        updateUserLocation();
      },
      (error) => {
        gps.error = error.message;
        gpsStatus = 'error';
      },
      {
        enableHighAccuracy: true,
        maximumAge: 10000
      }
    );
  }

  // Handle check-in
  async function handleCheckIn() {
    if (!canCheckIn) return;
    
    // Check if late
    if (isLate()) {
      showLateModal = true;
      return;
    }

    await performCheckIn();
  }

  async function performCheckIn(reason = null) {
    checkingIn = true;
    try {
      const result = await api.checkIn(gps.lat, gps.lng, reason);
      todayAttendance = result.data;
      showLateModal = false;
      lateReason = '';
    } catch (err) {
      console.error('Check-in failed:', err);
      alert('Gagal check-in: ' + (err.message || 'Unknown error'));
    } finally {
      checkingIn = false;
    }
  }

  async function handleCheckOut() {
    if (!todayAttendance || todayAttendance.check_out_time) return;

    if (!gps.lat || !gps.lng) {
      alert('Lokasi GPS belum siap');
      return;
    }

    checkingIn = true;
    try {
      const result = await api.checkOut(gps.lat, gps.lng);
      todayAttendance = result.data;
    } catch (err) {
      console.error('Check-out failed:', err);
      alert('Gagal check-out: ' + (err.message || 'Unknown error'));
    } finally {
      checkingIn = false;
    }
  }

  function submitLateReason() {
    if (!lateReason.trim()) {
      alert('Mohon isi alasan keterlambatan!');
      return;
    }
    performCheckIn(lateReason);
  }

  async function submitPermission() {
    try {
      const payload = {
        type: permissionForm.type,
        reason: permissionForm.reason,
        document: permissionForm.document,
        latitude: gps.lat,
        longitude: gps.lng
      };
      
      await api.submitPermission(payload);
      showPermissionModal = false;
      permissionForm = { type: 'permission', reason: '', document: null };
      await fetchDashboardData();
    } catch (err) {
      console.error('Permission submission failed:', err);
      alert('Gagal mengajukan izin: ' + (err.message || 'Unknown error'));
    }
  }

  function handleFileChange(e) {
    const file = e.target.files?.[0];
    if (file) {
      permissionForm.document = file;
    }
  }

  // Initialize charts
  function initCharts() {
    // Task Progress Donut Chart
    const taskProgressCtx = document.getElementById('taskProgressChart');
    if (taskProgressCtx && taskProgressCtx instanceof HTMLCanvasElement) {
      if (taskProgressChart) taskProgressChart.destroy();
      
      taskProgressChart = new Chart(taskProgressCtx.getContext('2d'), {
        type: 'doughnut',
        data: {
          labels: ['Pending', 'Proses', 'Submitted', 'Completed', 'Revisi'],
          datasets: [{
            data: [
              dashboardData.taskBreakdown.pending,
              dashboardData.taskBreakdown.in_progress,
              dashboardData.taskBreakdown.submitted,
              dashboardData.taskBreakdown.completed,
              dashboardData.taskBreakdown.revision
            ],
            backgroundColor: ['#fbbf24', '#6366f1', '#a855f7', '#10b981', '#f43f5e'],
            borderWidth: 0,
            hoverOffset: 4
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: {
              position: 'right',
              labels: {
                usePointStyle: true,
                pointStyle: 'circle',
                boxWidth: 6,
                font: { size: 10, family: 'Inter' }
              }
            }
          },
          cutout: '70%'
        }
      });
    }

    // Weekly Attendance Bar Chart
    const weeklyAttendanceCtx = document.getElementById('weeklyAttendanceChart');
    if (weeklyAttendanceCtx && weeklyAttendanceCtx instanceof HTMLCanvasElement) {
      if (weeklyAttendanceChart) weeklyAttendanceChart.destroy();
      
      weeklyAttendanceChart = new Chart(weeklyAttendanceCtx.getContext('2d'), {
        type: 'bar',
        data: {
          labels: dashboardData.weeklyAttendance.days,
          datasets: [{
            label: 'Status',
            data: dashboardData.weeklyAttendance.counts,
            backgroundColor: dashboardData.weeklyAttendance.colors,
            borderRadius: 4,
            barThickness: 16
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          scales: {
            y: { beginAtZero: true, display: false },
            x: {
              grid: { display: false },
              ticks: { font: { size: 10 } }
            }
          },
          plugins: {
            legend: { display: false }
          }
        }
      });
    }
  }

  function formatTime(timeStr) {
    if (!timeStr) return '-';
    if (timeStr.includes('T')) {
      return new Date(timeStr).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
    }
    return timeStr;
  }

  function getChartColor(status) {
    switch (status) {
      case 'present': return '#10b981';
      case 'late': return '#f59e0b';
      case 'permission': return '#3b82f6';
      case 'sick': return '#a855f7';
      case 'absent': return '#ef4444';
      default: return '#cbd5e1';
    }
  }

  function getPriorityColor(priority) {
    switch (priority) {
      case 'high': return 'bg-rose-100 text-rose-800 border-rose-200';
      case 'medium': return 'bg-amber-100 text-amber-800 border-amber-200';
      default: return 'bg-emerald-100 text-emerald-800 border-emerald-200';
    }
  }

  function getStatusColor(status) {
    switch (status) {
      case 'completed': return 'bg-emerald-100 text-emerald-700';
      case 'in_progress': return 'bg-indigo-100 text-indigo-700';
      case 'submitted': return 'bg-purple-100 text-purple-700';
      case 'revision': return 'bg-rose-100 text-rose-700';
      default: return 'bg-amber-100 text-amber-700';
    }
  }

  function getStatusLabel(status) {
    switch (status) {
      case 'pending': return 'Pending';
      case 'in_progress': return 'Dalam Proses';
      case 'submitted': return 'Submitted';
      case 'completed': return 'Selesai';
      case 'revision': return 'Revisi';
      default: return status;
    }
  }

  onMount(() => {
    // Load Leaflet CSS and JS
    const link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css';
    link.integrity = 'sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=';
    link.crossOrigin = '';
    document.head.appendChild(link);

    const script = document.createElement('script');
    script.src = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.js';
    script.integrity = 'sha256-20nQCchB9co0qIjJZRGuk2/Z9VM+kNiyxNV1lvTlZBo=';
    script.crossOrigin = '';
    script.onload = () => {
      initMap();
      initGPS();
    };
    document.head.appendChild(script);

    fetchDashboardData();

    return () => {
      if (map) map.remove();
      if (taskProgressChart) taskProgressChart.destroy();
      if (weeklyAttendanceChart) weeklyAttendanceChart.destroy();
    };
  });
</script>

<div class="slide-up max-w-[1600px] mx-auto space-y-6">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
    <div>
      <h2 class="text-2xl sm:text-3xl font-bold text-slate-800 tracking-tight mb-1">
        Selamat datang kembali! {auth.user?.name}
      </h2>
      <p class="text-slate-500 font-medium text-sm sm:text-base">
        Berikut ringkasan aktivitas magang Anda.
      </p>
    </div>
    <div class="text-right hidden sm:block">
      <div class="text-sm font-semibold text-slate-600">
        {new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
      </div>
    </div>
  </div>

  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- LEFT COLUMN (2/3) -->
    <div class="lg:col-span-2 space-y-6">
      <!-- ATTENDANCE CARD -->
      <div class="card p-0 overflow-hidden">
        <div class="p-4 sm:p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
          <h3 class="font-bold text-lg text-slate-800 flex items-center gap-2">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-indigo-500">
              <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/>
              <circle cx="12" cy="10" r="3"/>
            </svg>
            Presensi Harian
          </h3>
          <div class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold {gpsStatus === 'ready' ? 'bg-emerald-100 text-emerald-700' : gpsStatus === 'error' ? 'bg-rose-100 text-rose-700' : 'bg-slate-100 text-slate-500'}">
            <div class="w-1.5 h-1.5 rounded-full {gpsStatus === 'ready' ? 'bg-emerald-500' : gpsStatus === 'error' ? 'bg-rose-500' : 'bg-slate-400'} mr-1.5 {gpsStatus === 'loading' ? 'animate-pulse' : ''}"></div>
            {gpsStatus === 'ready' ? 'GPS Aktif' : gpsStatus === 'error' ? 'GPS Error' : 'Mencari Lokasi...'}
          </div>
        </div>

        <div class="p-4 sm:p-6 space-y-4">
          <!-- Map -->
          <div class="relative h-[220px] sm:h-[300px] w-full rounded-xl overflow-hidden border border-slate-200 shadow-inner">
            <div id="map" class="h-full w-full z-0"></div>
          </div>

          <!-- Office Info -->
          <div class="flex flex-col sm:flex-row justify-between items-center gap-4 bg-slate-50 p-4 rounded-xl border border-slate-100">
            <span class="text-sm font-medium text-slate-600 flex items-center gap-2">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-slate-400">
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                <polyline points="9 22 9 12 15 12 15 22"/>
              </svg>
              Kantor: <span class="font-bold text-slate-800">{OFFICE.name}</span>
            </span>
            <span class="text-sm font-medium text-slate-600">
              Jarak: <span class="font-mono font-bold text-slate-800">{distance !== null ? distance + ' m' : '-- m'}</span>
            </span>
          </div>

          <!-- Action Buttons -->
          <div>
            {#if loading}
              <div class="text-center py-4">
                <div class="inline-block w-6 h-6 border-3 border-indigo-600 border-t-transparent rounded-full animate-spin"></div>
              </div>
            {:else if !todayAttendance}
              <!-- Check-in Button -->
              <button
                onclick={handleCheckIn}
                disabled={!canCheckIn || checkingIn}
                class="btn w-full py-3.5 text-base font-bold shadow-lg shadow-indigo-500/20 disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-slate-300 disabled:text-slate-500 bg-indigo-600 text-white hover:bg-indigo-700 transition-all"
              >
                {#if checkingIn}
                  <div class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                {/if}
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="inline mr-2">
                  <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/>
                  <circle cx="12" cy="10" r="3"/>
                </svg>
                {canCheckIn ? 'PRESENSI MASUK SEKARANG' : (distance !== null ? (distance > OFFICE.maxDistance ? `TERLALU JAUH (${distance}m)` : 'Menunggu GPS...') : 'Menunggu GPS...')}
              </button>
            {:else if !todayAttendance.check_out_time && todayAttendance.status !== 'permission' && todayAttendance.status !== 'sick'}
              <!-- Check-out Button -->
              <div class="text-center mb-4">
                <div class="inline-flex items-center px-4 py-2 rounded-lg bg-emerald-50 text-emerald-700 font-medium text-sm border border-emerald-100">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="mr-2">
                    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                    <polyline points="22 4 12 14.01 9 11.01"/>
                  </svg>
                  Anda masuk pukul <strong class="ml-1">{formatTime(todayAttendance.check_in_time)}</strong>
                </div>
              </div>
              <button
                onclick={handleCheckOut}
                disabled={checkingIn}
                class="btn w-full py-3.5 text-base font-bold bg-amber-500 text-white hover:bg-amber-600 shadow-lg shadow-amber-500/20 transition-all disabled:opacity-50"
              >
                {#if checkingIn}
                  <div class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                {/if}
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="inline mr-2">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
                  <polyline points="16 17 21 12 16 7"/>
                  <line x1="21" y1="12" x2="9" y2="12"/>
                </svg>
                PRESENSI KELUAR SEKARANG
              </button>
            {:else}
              <!-- Attendance Completed -->
              <div class="text-center p-6 bg-emerald-50/50 border border-emerald-100 rounded-xl">
                <div class="w-16 h-16 bg-emerald-100 text-emerald-600 rounded-full flex items-center justify-center text-3xl mx-auto mb-3 shadow-sm">
                  <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"/>
                  </svg>
                </div>
                <h4 class="font-bold text-emerald-800 text-lg">Selesai Hari Ini</h4>
                <p class="text-emerald-600 font-medium text-sm mt-1">
                  Status: <span class="capitalize">{todayAttendance.status}</span>
                </p>
              </div>
            {/if}
          </div>

          <!-- Permission Link -->
          {#if !todayAttendance}
            <div class="text-center pt-2">
              <button
                onclick={() => showPermissionModal = true}
                class="text-sm text-indigo-600 hover:text-indigo-700 font-semibold hover:underline"
              >
                Tidak bisa datang? Ajukan Izin/Sakit
              </button>
            </div>
          {/if}
        </div>
      </div>

      <!-- TASKS CARD -->
      <div class="card p-0 overflow-hidden">
        <div class="p-4 sm:p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
          <h3 class="font-bold text-lg text-slate-800 flex items-center gap-2">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-teal-500">
              <path d="M9 11l3 3L22 4"/>
              <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>
            </svg>
            Tugas Saya
          </h3>
          <a href="/tasks" class="text-xs font-bold text-teal-600 hover:text-teal-700 uppercase tracking-wider hover:underline">
            Lihat Semua
          </a>
        </div>

        <div class="p-4 sm:p-6 space-y-6">
          <!-- Task Progress Stats -->
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
              <div class="text-2xl font-black text-indigo-600">{dashboardData.taskStats.in_progress}</div>
              <div class="text-[10px] font-bold text-indigo-600/70 uppercase tracking-wider mt-1">Proses</div>
            </div>
            <div class="bg-emerald-50 p-3 sm:p-4 rounded-xl border border-emerald-100 text-center hover:bg-white hover:shadow-md transition-all">
              <div class="text-2xl font-black text-emerald-600">{dashboardData.completedTasks}</div>
              <div class="text-[10px] font-bold text-emerald-600/70 uppercase tracking-wider mt-1">Selesai</div>
            </div>
          </div>

          <!-- Recent Tasks -->
          <div class="space-y-3">
            <h4 class="text-sm font-bold text-slate-700 uppercase tracking-wider">Tugas Terbaru</h4>
            {#if dashboardData.recentTasks.length > 0}
              {#each dashboardData.recentTasks as task}
                <div class="bg-white p-4 rounded-xl border border-slate-100 hover:border-slate-300 hover:shadow-md transition-all">
                  <div class="flex items-start justify-between gap-3 mb-2">
                    <h5 class="font-bold text-slate-800 text-sm flex-1">{task.title}</h5>
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold {getPriorityColor(task.priority)}">
                      {task.priority}
                    </span>
                  </div>
                  <p class="text-xs text-slate-500 mb-3 line-clamp-2">{task.description || 'Tidak ada deskripsi'}</p>
                  <div class="flex items-center justify-between text-xs">
                    <span class="inline-flex items-center px-2 py-0.5 rounded font-bold {getStatusColor(task.status)}">
                      {getStatusLabel(task.status)}
                    </span>
                    <span class="text-slate-400 flex items-center gap-1">
                      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                        <line x1="16" y1="2" x2="16" y2="6"/>
                        <line x1="8" y1="2" x2="8" y2="6"/>
                        <line x1="3" y1="10" x2="21" y2="10"/>
                      </svg>
                      {task.deadline || 'No deadline'}
                    </span>
                  </div>
                </div>
              {/each}
            {:else}
              <div class="text-center py-8 text-slate-400">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="mx-auto mb-2 opacity-50">
                  <path d="M9 11l3 3L22 4"/>
                  <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>
                </svg>
                <p class="text-sm">Belum ada tugas</p>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </div>

    <!-- RIGHT COLUMN (1/3) -->
    <div class="space-y-6">
      <!-- Task Progress Chart -->
      <div class="card p-0 overflow-hidden">
        <div class="p-4 border-b border-slate-100 bg-slate-50/50">
          <h3 class="font-bold text-base text-slate-800">Progress Tugas</h3>
        </div>
        <div class="p-4">
          <div class="h-48">
            <canvas id="taskProgressChart"></canvas>
          </div>
        </div>
      </div>

      <!-- Weekly Attendance Chart -->
      <div class="card p-0 overflow-hidden">
        <div class="p-4 border-b border-slate-100 bg-slate-50/50">
          <h3 class="font-bold text-base text-slate-800">Kehadiran Mingguan</h3>
        </div>
        <div class="p-4">
          <div class="h-32">
            <canvas id="weeklyAttendanceChart"></canvas>
          </div>
        </div>
      </div>

      <!-- Attendance Percentage -->
      <div class="card">
        <h4 class="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">Persentase Kehadiran</h4>
        <div class="text-4xl font-black text-slate-800 mb-2">{dashboardData.attendancePercentage}%</div>
        <div class="w-full bg-slate-200 rounded-full h-3 overflow-hidden">
          <div class="bg-gradient-to-r from-emerald-500 to-teal-500 h-full rounded-full transition-all" style="width: {dashboardData.attendancePercentage}%"></div>
        </div>
      </div>
    </div>
  </div>
</div>

<!-- Late Reason Modal -->
{#if showLateModal}
  <div 
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/40 backdrop-blur-sm" 
    role="button" 
    tabindex="0" 
    onclick={(e) => { if (e.target === e.currentTarget) showLateModal = false; }} 
    onkeydown={(e) => { if (e.key === 'Escape') showLateModal = false; }}
  >
    <div class="bg-white rounded-2xl shadow-xl w-full max-w-md p-6" role="document">
      <h3 class="text-lg font-bold text-slate-800 mb-4">Alasan Terlambat</h3>
      <p class="text-sm text-slate-600 mb-4">Anda terlambat hari ini. Mohon isi alasan keterlambatan Anda.</p>
      <textarea
        bind:value={lateReason}
        class="w-full px-4 py-3 border border-slate-200 rounded-lg resize-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
        rows="4"
        placeholder="Contoh: Terlambat karena macet..."
      ></textarea>
      <div class="flex gap-3 mt-4">
        <button onclick={() => showLateModal = false} class="btn btn-outline flex-1">Batal</button>
        <button onclick={submitLateReason} class="btn btn-primary flex-1" disabled={checkingIn}>
          {checkingIn ? 'Mengirim...' : 'Kirim & Check In'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Permission Modal -->
{#if showPermissionModal}
  <div 
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/40 backdrop-blur-sm" 
    role="button" 
    tabindex="0" 
    onclick={(e) => { if (e.target === e.currentTarget) showPermissionModal = false; }} 
    onkeydown={(e) => { if (e.key === 'Escape') showPermissionModal = false; }}
  >
    <div class="bg-white rounded-2xl shadow-xl w-full max-w-md p-6" role="document">
      <h3 class="text-lg font-bold text-slate-800 mb-4">Ajukan Izin/Sakit</h3>
      
      <div class="space-y-4">
        <div>
          <label for="permission-type" class="block text-sm font-semibold text-slate-700 mb-2">Tipe</label>
          <select id="permission-type" bind:value={permissionForm.type} class="w-full px-4 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500">
            <option value="permission">Izin</option>
            <option value="sick">Sakit</option>
          </select>
        </div>

        <div>
          <label for="permission-reason" class="block text-sm font-semibold text-slate-700 mb-2">Alasan</label>
          <textarea
            id="permission-reason"
            bind:value={permissionForm.reason}
            class="w-full px-4 py-3 border border-slate-200 rounded-lg resize-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            rows="4"
            placeholder="Jelaskan alasan Anda..."
          ></textarea>
        </div>

        <div>
          <label for="permission-doc" class="block text-sm font-semibold text-slate-700 mb-2">Dokumen Pendukung (Opsional)</label>
          <input
            id="permission-doc"
            type="file"
            accept="image/*,.pdf"
            onchange={handleFileChange}
            class="w-full px-4 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm"
          />
        </div>
      </div>

      <div class="flex gap-3 mt-6">
        <button onclick={() => showPermissionModal = false} class="btn btn-outline flex-1">Batal</button>
        <button onclick={submitPermission} class="btn btn-primary flex-1">Kirim</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .slide-up {
    animation: slideUp 0.4s ease-out;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>