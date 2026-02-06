<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';
  import Chart from 'chart.js/auto';

  // ==================== STATE ====================
  let loading = $state(true);
  let stats = $state({
    total_interns: 0,
    total_tasks: 0,
    completed_on_time: 0,
    completed_late: 0,
    pending_tasks: 0,
    present_today: 0,
    total_today: 0
  });

  let recentTasks = $state([]);
  let todayAttendance = $state([]);
  let weeklyTrend = $state([]);
  let pendingRegistrations = $state(0);

  // Charts
  let taskCompletionChart = null;
  let weeklyAttendanceChart = null;

  // ==================== DERIVED STATE ====================
  const attendanceRate = $derived(
    stats.total_today > 0 
      ? Math.round((stats.present_today / stats.total_today) * 100) 
      : 0
  );

  const taskCompletionRate = $derived(
    (stats.completed_on_time + stats.completed_late) > 0
      ? Math.round((stats.completed_on_time / (stats.completed_on_time + stats.completed_late)) * 100)
      : 0
  );

  // ==================== DATA FETCHING ====================
  
  async function fetchDashboardData() {
    loading = true;
    try {
      // Fetch admin dashboard data
      const res = await api.request('/analytics/dashboard/admin');
      const data = res.data;

      stats = data.stats || stats;
      recentTasks = data.recent_tasks || [];
      todayAttendance = data.today_attendance || [];
      weeklyTrend = data.weekly_trend || [];

      // Fetch pending registrations count (interns with pending status)
      try {
        const internsRes = await api.getInterns({ status: 'pending' });
        pendingRegistrations = internsRes.data?.pagination?.total || 0;
      } catch (err) {
        console.error('Failed to fetch pending registrations:', err);
      }

      // Initialize charts
      setTimeout(initCharts, 100);
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
    } finally {
      loading = false;
    }
  }

  // ==================== CHART INITIALIZATION ====================
  
  function initCharts() {
    // Destroy existing charts
    if (taskCompletionChart) taskCompletionChart.destroy();
    if (weeklyAttendanceChart) weeklyAttendanceChart.destroy();

    // Task Completion Doughnut Chart
    const taskCanvas = document.getElementById('taskCompletionChart');
    if (taskCanvas instanceof HTMLCanvasElement && (stats.completed_on_time > 0 || stats.completed_late > 0 || stats.pending_tasks > 0)) {
      const ctx = taskCanvas.getContext('2d');
      
      taskCompletionChart = new Chart(ctx, {
        type: 'doughnut',
        data: {
          labels: ['Tepat Waktu', 'Terlambat', 'Pending'],
          datasets: [{
            data: [
              stats.completed_on_time,
              stats.completed_late,
              stats.pending_tasks
            ],
            backgroundColor: [
              '#10b981', // emerald - on time
              '#f59e0b', // amber - late
              '#6366f1'  // indigo - pending
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

    // Weekly Attendance Trend Line Chart
    const weeklyCanvas = document.getElementById('weeklyAttendanceChart');
    if (weeklyCanvas instanceof HTMLCanvasElement && weeklyTrend?.length > 0) {
      const ctx = weeklyCanvas.getContext('2d');
      
      weeklyAttendanceChart = new Chart(ctx, {
        type: 'line',
        data: {
          labels: weeklyTrend.map(d => d.day),
          datasets: [
            {
              label: 'Hadir',
              data: weeklyTrend.map(d => d.present),
              borderColor: '#10b981',
              backgroundColor: 'rgba(16, 185, 129, 0.1)',
              borderWidth: 2,
              fill: true,
              tension: 0.4
            },
            {
              label: 'Tidak Hadir',
              data: weeklyTrend.map(d => d.absent),
              borderColor: '#ef4444',
              backgroundColor: 'rgba(239, 68, 68, 0.1)',
              borderWidth: 2,
              fill: true,
              tension: 0.4
            }
          ]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          scales: {
            y: {
              beginAtZero: true,
              ticks: { 
                stepSize: 1,
                font: { size: 10 }
              },
              grid: { color: 'rgba(0, 0, 0, 0.05)' }
            },
            x: {
              grid: { display: false },
              ticks: { font: { size: 11, weight: 600 } }
            }
          },
          plugins: {
            legend: {
              position: 'top',
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
  }

  // ==================== LIFECYCLE ====================
  
  onMount(async () => {
    await fetchDashboardData();
  });

  onDestroy(() => {
    if (taskCompletionChart) taskCompletionChart.destroy();
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
      revision: 'Revision',
      present: 'Hadir',
      late: 'Terlambat',
      absent: 'Tidak Hadir',
      sick: 'Sakit',
      permission: 'Izin'
    };
    return labels[status] || status;
  }

  function getAttendanceStatusBadgeClass(status) {
    const classes = {
      present: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      late: 'bg-amber-50 text-amber-700 border-amber-200',
      absent: 'bg-red-50 text-red-700 border-red-200',
      sick: 'bg-blue-50 text-blue-700 border-blue-200',
      permission: 'bg-indigo-50 text-indigo-700 border-indigo-200'
    };
    return classes[status] || 'bg-slate-50 text-slate-700 border-slate-200';
  }
</script>

<style>
  .stat-card, .card {
    background: rgba(255, 255, 255, 0.85) !important;
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }
</style>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
  <!-- Header -->
  <div class="mb-6 md:mb-8">
    <h2 class="text-2xl sm:text-3xl font-bold text-slate-800 mb-2">
      Selamat Datang, {auth.user?.name || 'Admin'}! 👋
    </h2>
    <p class="text-slate-600 text-sm sm:text-base">
      Pantau aktivitas magang Interns dalam sekilas.
    </p>
  </div>

  {#if loading}
    <div class="text-center py-12">
      <div class="inline-block w-12 h-12 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
      <p class="mt-4 text-slate-600">Loading dashboard...</p>
    </div>
  {:else}
    <!-- Pending Registration Alert -->
    {#if pendingRegistrations > 0}
      <div class="card p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6"
           style="background: linear-gradient(135deg, rgba(251,191,36,0.15) 0%, rgba(245,158,11,0.15) 100%); border: 2px solid rgba(251,191,36,0.4);">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-2xl bg-amber-500 text-white flex items-center justify-center shadow-lg shadow-amber-500/30">
            <i class="fas fa-user-clock text-xl"></i>
          </div>
          <div>
            <p class="font-bold text-amber-800 text-base">{pendingRegistrations} Pendaftaran Menunggu Approval</p>
            <p class="text-sm text-amber-600">Ada calon magang yang mendaftar dan membutuhkan persetujuan Anda.</p>
          </div>
        </div>
        <a href="/interns?status=pending" 
           class="btn bg-amber-500 hover:bg-amber-600 text-white shadow-lg shadow-amber-500/30 whitespace-nowrap">
          <i class="fas fa-eye"></i> Lihat & Approve
        </a>
      </div>
    {/if}

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
      <!-- Total Siswa -->
      <div class="stat-card bg-white/85 backdrop-blur-xl rounded-2xl p-6 border border-slate-200/80 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-200">
        <div class="flex items-center justify-center w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-violet-100 text-violet-700 mb-4">
          <i class="fas fa-users text-2xl"></i>
        </div>
        <div class="text-3xl sm:text-4xl font-extrabold text-slate-800 mb-1">
          {stats.total_interns}
        </div>
        <div class="text-sm font-medium text-slate-500 uppercase tracking-wide">
          Total Siswa
        </div>
      </div>

      <!-- Tepat Waktu -->
      <div class="stat-card bg-white/85 backdrop-blur-xl rounded-2xl p-6 border border-slate-200/80 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-200">
        <div class="flex items-center justify-center w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-green-100 text-green-700 mb-4">
          <i class="fas fa-check-circle text-2xl"></i>
        </div>
        <div class="text-3xl sm:text-4xl font-extrabold text-slate-800 mb-1">
          {stats.completed_on_time}
        </div>
        <div class="text-sm font-medium text-slate-500 uppercase tracking-wide">
          Tepat Waktu
        </div>
      </div>

      <!-- Terlambat -->
      <div class="stat-card bg-white/85 backdrop-blur-xl rounded-2xl p-6 border border-slate-200/80 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-200">
        <div class="flex items-center justify-center w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-orange-100 text-orange-700 mb-4">
          <i class="fas fa-clock text-2xl"></i>
        </div>
        <div class="text-3xl sm:text-4xl font-extrabold text-slate-800 mb-1">
          {stats.completed_late}
        </div>
        <div class="text-sm font-medium text-slate-500 uppercase tracking-wide">
          Terlambat
        </div>
      </div>

      <!-- Presensi Hari Ini -->
      <div class="stat-card bg-white/85 backdrop-blur-xl rounded-2xl p-6 border border-slate-200/80 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-200">
        <div class="flex items-center justify-center w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-blue-100 text-blue-700 mb-4">
          <i class="fas fa-calendar-check text-2xl"></i>
        </div>
        <div class="text-3xl sm:text-4xl font-extrabold text-slate-800 mb-1">
          {stats.present_today}/{stats.total_today}
        </div>
        <div class="text-sm font-medium text-slate-500 uppercase tracking-wide">
          Hadir Hari Ini
        </div>
        <div class="mt-2 h-1.5 bg-slate-100 rounded-full overflow-hidden">
          <div class="h-full bg-gradient-to-r from-blue-500 to-indigo-500 rounded-full transition-all duration-500"
               style="width: {attendanceRate}%"></div>
        </div>
      </div>
    </div>

    <!-- Main Content Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      
      <!-- Left Column (2/3) -->
      <div class="lg:col-span-2 space-y-6">
        
        <!-- Recent Tasks -->
        <div class="card p-0 overflow-hidden">
          <div class="p-4 sm:p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
            <h3 class="font-bold text-lg text-slate-800 flex items-center gap-2">
              <i class="fas fa-tasks text-teal-500"></i> Tugas Terbaru
            </h3>
            <a href="/tasks" class="text-xs font-bold text-teal-600 hover:text-teal-700 uppercase tracking-wider hover:underline">
              Lihat Semua
            </a>
          </div>

          <div class="p-4 sm:p-6">
            {#if recentTasks?.length > 0}
              <div class="space-y-3">
                {#each recentTasks as task}
                  <div class="p-4 bg-slate-50 rounded-lg border border-slate-100 hover:shadow-sm transition-all">
                    <div class="flex justify-between items-start mb-2">
                      <div class="flex-1">
                        <h4 class="font-semibold text-slate-800 mb-1">{task.title}</h4>
                        <p class="text-xs text-slate-500">
                          <i class="fas fa-user mr-1"></i>{task.intern_name}
                        </p>
                      </div>
                      <div class="flex flex-col items-end gap-2">
                        <span class="px-2.5 py-1 rounded text-[10px] font-bold border {getStatusBadgeClass(task.status)}">
                          {getStatusLabel(task.status)}
                        </span>
                        {#if task.is_late}
                          <span class="px-2 py-0.5 rounded text-[9px] font-bold bg-red-50 text-red-600 border border-red-200">
                            <i class="fas fa-exclamation-circle mr-1"></i>LATE
                          </span>
                        {/if}
                      </div>
                    </div>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="text-center py-8 text-slate-400">
                <i class="fas fa-inbox text-3xl mb-2"></i>
                <p class="text-sm">Tidak ada tugas terbaru</p>
              </div>
            {/if}
          </div>
        </div>

        <!-- Today's Attendance -->
        <div class="card p-0 overflow-hidden">
          <div class="p-4 sm:p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
            <h3 class="font-bold text-lg text-slate-800 flex items-center gap-2">
              <i class="fas fa-clipboard-check text-indigo-500"></i> Presensi Hari Ini
            </h3>
            <a href="/attendance" class="text-xs font-bold text-indigo-600 hover:text-indigo-700 uppercase tracking-wider hover:underline">
              Lihat Semua
            </a>
          </div>

          <div class="p-4 sm:p-6">
            {#if todayAttendance?.length > 0}
              <div class="overflow-x-auto">
                <table class="w-full">
                  <thead>
                    <tr class="border-b border-slate-200">
                      <th class="text-left py-3 px-2 text-xs font-bold text-slate-600 uppercase tracking-wider">Nama</th>
                      <th class="text-left py-3 px-2 text-xs font-bold text-slate-600 uppercase tracking-wider">Status</th>
                      <th class="text-left py-3 px-2 text-xs font-bold text-slate-600 uppercase tracking-wider">Jam Masuk</th>
                      <th class="text-right py-3 px-2 text-xs font-bold text-slate-600 uppercase tracking-wider">Jarak</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each todayAttendance as attendance}
                      <tr class="border-b border-slate-100 hover:bg-slate-50 transition-colors">
                        <td class="py-3 px-2 text-sm font-medium text-slate-800">{attendance.intern_name}</td>
                        <td class="py-3 px-2">
                          <span class="px-2 py-0.5 rounded text-[10px] font-bold border {getAttendanceStatusBadgeClass(attendance.status)}">
                            {getStatusLabel(attendance.status)}
                          </span>
                        </td>
                        <td class="py-3 px-2 text-sm text-slate-600 font-mono">
                          {attendance.check_in_time || '-'}
                        </td>
                        <td class="py-3 px-2 text-right text-sm text-slate-600 font-mono">
                          {#if attendance.distance !== null && attendance.distance !== undefined}
                            {attendance.distance}m
                          {:else}
                            -
                          {/if}
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {:else}
              <div class="text-center py-8 text-slate-400">
                <i class="fas fa-calendar-times text-3xl mb-2"></i>
                <p class="text-sm">Belum ada presensi hari ini</p>
              </div>
            {/if}
          </div>
        </div>
      </div>

      <!-- Right Column (1/3) -->
      <div class="space-y-6">
        
        <!-- Task Completion Chart -->
        <div class="card p-4 sm:p-6">
          <h3 class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2">
            <i class="fas fa-chart-pie text-emerald-500"></i> Penyelesaian Tugas
          </h3>
          <div class="h-[240px] mb-4">
            <canvas id="taskCompletionChart"></canvas>
          </div>
          <div class="text-center">
            <div class="text-sm text-slate-600 mb-1">Tingkat Ketepatan Waktu</div>
            <div class="text-3xl font-black text-slate-800">{taskCompletionRate}%</div>
          </div>
        </div>

        <!-- Weekly Attendance Trend -->
        <div class="card p-4 sm:p-6">
          <h3 class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2">
            <i class="fas fa-chart-line text-blue-500"></i> Tren Mingguan
          </h3>
          <div class="h-[240px]">
            <canvas id="weeklyAttendanceChart"></canvas>
          </div>
        </div>

        <!-- Quick Stats -->
        <div class="card p-4 sm:p-6">
          <h3 class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2">
            <i class="fas fa-info-circle text-violet-500"></i> Info Cepat
          </h3>
          <div class="space-y-3">
            <div class="flex justify-between items-center p-3 bg-slate-50 rounded-lg">
              <span class="text-sm font-medium text-slate-600">Total Tugas</span>
              <span class="text-lg font-black text-slate-800">{stats.total_tasks}</span>
            </div>
            <div class="flex justify-between items-center p-3 bg-slate-50 rounded-lg">
              <span class="text-sm font-medium text-slate-600">Tugas Pending</span>
              <span class="text-lg font-black text-amber-600">{stats.pending_tasks}</span>
            </div>
            <div class="flex justify-between items-center p-3 bg-slate-50 rounded-lg">
              <span class="text-sm font-medium text-slate-600">Tingkat Kehadiran</span>
              <span class="text-lg font-black text-emerald-600">{attendanceRate}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>