<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';
  import Chart from 'chart.js/auto';

  // State
  let loading = $state(true);
  let dashboardData = $state({
    totalInterns: 0,
    completedOnTime: 0,
    completedLate: 0,
    presentToday: 0,
    pendingRegistrations: 0,
    pendingTasks: 0,
    taskBreakdown: { completed_on_time: 0, completed_late: 0, pending: 0 },
    attendanceToday: { present: 0, late: 0, permission: 0, sick: 0, absent: 0 },
    attendanceTrend: [],
    submittedTasks: [],
    recentActivities: []
  });

  // Chart instances
  let taskPieChart = null;
  let attendanceTodayChart = null;
  let attendanceTrendChart = null;

  // Fetch dashboard data
  async function fetchDashboardData() {
    loading = true;
    try {
      const [
        dashboardRes,
        pendingInternsRes,
        submittedTasksRes
      ] = await Promise.all([
        api.getAdminDashboard(),
        api.getInterns({ status: 'pending', limit: 1 }),
        api.getTasks({ status: 'submitted', limit: 5 })
      ]);

      const data = dashboardRes.data;
      const todayAtt = data.today_attendance || [];
      
      dashboardData = {
        totalInterns: data.stats.total_interns,
        completedOnTime: data.stats.completed_on_time,
        completedLate: data.stats.completed_late,
        presentToday: data.stats.present_today,
        pendingRegistrations: pendingInternsRes.meta?.total || 0,
        pendingTasks: data.stats.pending_tasks,
        taskBreakdown: {
          completed_on_time: data.stats.completed_on_time,
          completed_late: data.stats.completed_late,
          pending: data.stats.pending_tasks
        },
        attendanceToday: {
          present: todayAtt.filter(a => a.status === 'present').length,
          late: todayAtt.filter(a => a.status === 'late').length,
          permission: todayAtt.filter(a => a.status === 'permission').length,
          sick: todayAtt.filter(a => a.status === 'sick').length,
          absent: data.stats.total_interns - todayAtt.length
        },
        attendanceTrend: data.weekly_trend,
        submittedTasks: submittedTasksRes.data || [],
        recentActivities: []
      };

      // Initialize charts after data is loaded
      $effect(() => {
        if (!loading) {
          setTimeout(initCharts, 100);
        }
      });
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
    } finally {
      loading = false;
    }
  }

  function initCharts() {
    // Task Pie Chart
    const taskPieCtx = document.getElementById('taskPieChart');
    if (taskPieCtx && taskPieCtx instanceof HTMLCanvasElement) {
      if (taskPieChart) taskPieChart.destroy();
      taskPieChart = new Chart(taskPieCtx.getContext('2d'), {
        type: 'pie',
        data: {
          labels: ['Tepat Waktu', 'Terlambat', 'Dalam Proses'],
          datasets: [{
            data: [
              dashboardData.taskBreakdown.completed_on_time,
              dashboardData.taskBreakdown.completed_late,
              dashboardData.taskBreakdown.pending
            ],
            backgroundColor: ['#10b981', '#f59e0b', '#8b5cf6'],
            borderWidth: 0,
            hoverOffset: 8
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: {
              position: 'right',
              align: 'end',
              labels: {
                color: '#64748b',
                usePointStyle: true,
                pointStyle: 'circle',
                padding: 20,
                font: { family: 'Inter', size: 13, weight: 500 }
              }
            },
            tooltip: {
              backgroundColor: 'rgba(30, 41, 59, 0.95)',
              titleColor: '#fff',
              bodyColor: '#e2e8f0',
              borderColor: 'rgba(255, 255, 255, 0.1)',
              borderWidth: 1,
              cornerRadius: 12,
              padding: 14
            }
          }
        }
      });
    }

    // Attendance Today Donut Chart
    const attendanceTodayCtx = document.getElementById('attendanceTodayChart');
    if (attendanceTodayCtx && attendanceTodayCtx instanceof HTMLCanvasElement) {
      if (attendanceTodayChart) attendanceTodayChart.destroy();
      attendanceTodayChart = new Chart(attendanceTodayCtx.getContext('2d'), {
        type: 'doughnut',
        data: {
          labels: ['Hadir', 'Terlambat', 'Izin', 'Sakit', 'Belum Absen'],
          datasets: [{
            data: [
              dashboardData.attendanceToday.present,
              dashboardData.attendanceToday.late,
              dashboardData.attendanceToday.permission,
              dashboardData.attendanceToday.sick,
              dashboardData.attendanceToday.absent
            ],
            backgroundColor: ['#10b981', '#f59e0b', '#3b82f6', '#a855f7', '#ef4444'],
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
                color: '#64748b',
                font: { family: 'Inter', size: 11 },
                padding: 12,
                usePointStyle: true,
                pointStyle: 'circle'
              }
            }
          },
          cutout: '65%'
        }
      });
    }

    // Attendance Trend Bar Chart
    const attendanceTrendCtx = document.getElementById('attendanceTrendChart');
    if (attendanceTrendCtx && attendanceTrendCtx instanceof HTMLCanvasElement) {
      if (attendanceTrendChart) attendanceTrendChart.destroy();
      attendanceTrendChart = new Chart(attendanceTrendCtx.getContext('2d'), {
        type: 'bar',
        data: {
          labels: dashboardData.attendanceTrend.map(d => d.day),
          datasets: [
            {
              label: 'Hadir',
              data: dashboardData.attendanceTrend.map(d => d.present),
              backgroundColor: '#10b981',
              borderRadius: 6,
              barThickness: 20
            },
            {
              label: 'Tidak Hadir',
              data: dashboardData.attendanceTrend.map(d => d.absent),
              backgroundColor: '#ef4444',
              borderRadius: 6,
              barThickness: 20
            }
          ]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          scales: {
            x: {
              stacked: true,
              grid: { display: false },
              ticks: { color: '#64748b', font: { size: 11 } }
            },
            y: {
              stacked: true,
              beginAtZero: true,
              grid: { color: 'rgba(226, 232, 240, 0.6)' },
              ticks: { color: '#64748b', stepSize: 1, font: { size: 11 } }
            }
          },
          plugins: {
            legend: {
              position: 'top',
              align: 'end',
              labels: {
                color: '#64748b',
                usePointStyle: true,
                padding: 16,
                font: { size: 12 }
              }
            }
          }
        }
      });
    }
  }

  onMount(() => {
    fetchDashboardData();
  });
</script>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
  <!-- Header -->
  <div class="mb-6 md:mb-8">
    <h2 class="text-2xl sm:text-3xl font-bold text-slate-800 mb-2">
      Selamat Datang, {auth.user?.name}! 👋
    </h2>
    <p class="text-slate-600 text-sm sm:text-base">
      Dashboard ringkas dan bersih untuk memantau aktivitas magang.
    </p>
  </div>

  {#if loading}
    <div class="text-center py-20">
      <div class="inline-block w-8 h-8 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin"></div>
      <p class="text-slate-500 mt-4">Memuat dashboard...</p>
    </div>
  {:else}
    <!-- Pending Registration Alert -->
    {#if dashboardData.pendingRegistrations > 0}
      <div class="card p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6 bg-gradient-to-br from-amber-50 to-amber-100/30 border-2 border-amber-200">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-2xl bg-amber-500 text-white flex items-center justify-center shadow-lg shadow-amber-500/30">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
              <circle cx="8.5" cy="7" r="4"/>
              <polyline points="17 11 19 13 23 9"/>
            </svg>
          </div>
          <div>
            <p class="font-bold text-amber-800 text-base">{dashboardData.pendingRegistrations} Pendaftaran Menunggu Approval</p>
            <p class="text-sm text-amber-600">Ada calon magang yang mendaftar dan membutuhkan persetujuan Anda.</p>
          </div>
        </div>
        <a href="/interns?status=pending" class="btn bg-amber-500 hover:bg-amber-600 text-white shadow-lg shadow-amber-500/30 whitespace-nowrap px-5 py-2.5">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="inline mr-2">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
            <circle cx="12" cy="12" r="3"/>
          </svg>
          Lihat & Approve
        </a>
      </div>
    {/if}

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
      <!-- Total Siswa -->
      <div class="stat-card bg-white/85 backdrop-blur-xl rounded-2xl p-6 border border-slate-200/80 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-200">
        <div class="flex items-center justify-center w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-violet-100 text-violet-700 mb-4">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
            <circle cx="9" cy="7" r="4"/>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
          </svg>
        </div>
        <div class="text-3xl sm:text-4xl font-extrabold text-slate-800 mb-1">
          {dashboardData.totalInterns}
        </div>
        <div class="text-sm font-medium text-slate-500 uppercase tracking-wide">
          Total Siswa
        </div>
      </div>

      <!-- Tepat Waktu -->
      <div class="stat-card bg-white/85 backdrop-blur-xl rounded-2xl p-6 border border-slate-200/80 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-200">
        <div class="flex items-center justify-center w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-green-100 text-green-700 mb-4">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
        </div>
        <div class="text-3xl sm:text-4xl font-extrabold text-slate-800 mb-1">
          {dashboardData.completedOnTime}
        </div>
        <div class="text-sm font-medium text-slate-500 uppercase tracking-wide">
          Tepat Waktu
        </div>
      </div>

      <!-- Terlambat -->
      <div class="stat-card bg-white/85 backdrop-blur-xl rounded-2xl p-6 border border-slate-200/80 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-200">
        <div class="flex items-center justify-center w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-orange-100 text-orange-700 mb-4">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </div>
        <div class="text-3xl sm:text-4xl font-extrabold text-slate-800 mb-1">
          {dashboardData.completedLate}
        </div>
        <div class="text-sm font-medium text-slate-500 uppercase tracking-wide">
          Terlambat
        </div>
      </div>

      <!-- Kehadiran -->
      <div class="stat-card bg-white/85 backdrop-blur-xl rounded-2xl p-6 border border-slate-200/80 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-200">
        <div class="flex items-center justify-center w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-sky-100 text-sky-700 mb-4">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
            <line x1="16" y1="2" x2="16" y2="6"/>
            <line x1="8" y1="2" x2="8" y2="6"/>
            <line x1="3" y1="10" x2="21" y2="10"/>
            <path d="M9 16l2 2 4-4"/>
          </svg>
        </div>
        <div class="text-2xl sm:text-3xl font-extrabold text-slate-800 mb-1">
          {dashboardData.presentToday} / {dashboardData.totalInterns}
        </div>
        <div class="text-sm font-medium text-slate-500 uppercase tracking-wide">
          Kehadiran
        </div>
      </div>
    </div>

    <!-- Task Overview Card -->
    <div class="card bg-white/85 backdrop-blur-xl rounded-2xl border border-slate-200/80 shadow-sm overflow-hidden">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-6 border-b border-slate-100">
        <h3 class="text-lg font-semibold text-slate-700 flex items-center gap-2">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-white-500">
            <path d="M21.21 15.89A10 10 0 1 1 8 2.83"/>
            <path d="M22 12A10 10 0 0 0 12 2v10z"/>
          </svg>
          Statistik Tugas
        </h3>
        <a href="/tasks/create" class="inline-flex items-center justify-center px-4 py-2 bg-black hover:bg-gray-800 text-white text-sm font-semibold rounded-xl transition-colors duration-200">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="mr-2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          Buat Tugas
        </a>
      </div>
      <div class="p-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6 items-center">
          <div class="chart-container h-64 sm:h-72">
            <canvas id="taskPieChart"></canvas>
          </div>
          <div class="space-y-3">
            <div class="flex items-center gap-4 p-4 bg-slate-50 rounded-xl border border-slate-100">
              <div class="w-3 h-3 rounded-full bg-green-400 flex-shrink-0"></div>
              <div class="flex-1 text-sm font-medium text-slate-600">Tepat Waktu</div>
              <strong class="text-lg font-bold text-slate-800">{dashboardData.completedOnTime}</strong>
            </div>
            <div class="flex items-center gap-4 p-4 bg-slate-50 rounded-xl border border-slate-100">
              <div class="w-3 h-3 rounded-full bg-yellow-400 flex-shrink-0"></div>
              <div class="flex-1 text-sm font-medium text-slate-600">Terlambat</div>
              <strong class="text-lg font-bold text-slate-800">{dashboardData.completedLate}</strong>
            </div>
            <div class="flex items-center gap-4 p-4 bg-slate-50 rounded-xl border border-slate-100">
              <div class="w-3 h-3 rounded-full bg-violet-400 flex-shrink-0"></div>
              <div class="flex-1 text-sm font-medium text-slate-600">Dalam Proses</div>
              <strong class="text-lg font-bold text-slate-800">{dashboardData.pendingTasks}</strong>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Submitted Tasks (Pending Review) -->
    {#if dashboardData.submittedTasks.length > 0}
      <div class="card bg-gradient-to-br from-sky-50 to-blue-50 backdrop-blur-xl rounded-2xl border-2 border-sky-200 shadow-sm overflow-hidden">
        <div class="p-6 border-b border-sky-100">
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
            <div>
              <h3 class="text-lg font-bold text-sky-700 flex items-center gap-2 mb-1">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M9 11l3 3L22 4"/>
                  <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>
                </svg>
                Tugas Menunggu Review
                <span class="inline-flex items-center justify-center w-7 h-7 bg-sky-600 text-white text-xs font-bold rounded-full">
                  {dashboardData.submittedTasks.length}
                </span>
              </h3>
              <p class="text-sm text-sky-600">Siswa telah mengumpulkan tugas berikut dan menunggu review Anda.</p>
            </div>
            <a href="/tasks?status=submitted" class="btn bg-sky-600 hover:bg-sky-700 text-white shadow-lg shadow-sky-600/30 text-sm px-4 py-2">
              Lihat Semua
            </a>
          </div>
        </div>
        <div class="p-6">
          <div class="space-y-3">
            {#each dashboardData.submittedTasks as task}
  <div class="bg-white p-4 rounded-xl border border-sky-100 hover:border-sky-300 hover:shadow-md transition-all cursor-pointer flex items-center justify-between gap-4">

    <div class="flex-1">
      <h4 class="font-bold text-slate-800 mb-1">{task.title}</h4>

      <p class="text-sm text-slate-500 mb-2">
        {task.description?.substring(0, 100) || 'Tidak ada deskripsi'}
      </p>

      <div class="flex items-center gap-4 text-xs text-slate-500">
        <span class="flex items-center gap-1">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
            <circle cx="12" cy="7" r="4"/>
          </svg>
          {task.intern_name || 'Unknown'}
        </span>

        <span class="flex items-center gap-1">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
            <line x1="16" y1="2" x2="16" y2="6"/>
            <line x1="8" y1="2" x2="8" y2="6"/>
            <line x1="3" y1="10" x2="21" y2="10"/>
          </svg>
          {task.deadline || 'Tidak ada deadline'}
        </span>
      </div>
    </div>

    <a 
      href="/task-assignments/{task.id}" 
      class="btn btn-sm bg-sky-600 hover:bg-sky-700 text-white px-3 py-1.5 rounded transition-colors duration-200 whitespace-nowrap"
      aria-label="Review task assignment"
    >
      Review
    </a>

  </div>
{/each}

          </div>
        </div>
      </div>
    {/if}

    <!-- Charts Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Attendance Today -->
      <div class="card bg-white/85 backdrop-blur-xl rounded-2xl border border-slate-200/80 shadow-sm overflow-hidden">
        <div class="p-6 border-b border-slate-100">
          <h3 class="text-lg font-semibold text-slate-700 flex items-center gap-2">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-emerald-500">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
            Kehadiran Hari Ini
          </h3>
        </div>
        <div class="p-6">
          <div class="chart-container h-64">
            <canvas id="attendanceTodayChart"></canvas>
          </div>
        </div>
      </div>

      <!-- Attendance Trend -->
      <div class="card bg-white/85 backdrop-blur-xl rounded-2xl border border-slate-200/80 shadow-sm overflow-hidden">
        <div class="p-6 border-b border-slate-100">
          <h3 class="text-lg font-semibold text-slate-700 flex items-center gap-2">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-indigo-500">
              <polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/>
              <polyline points="17 6 23 6 23 12"/>
            </svg>
            Tren Kehadiran Mingguan
          </h3>
        </div>
        <div class="p-6">
          <div class="chart-container h-64">
            <canvas id="attendanceTrendChart"></canvas>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .stat-card {
    background: rgba(255, 255, 255, 0.85) !important;
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }

  .chart-container {
    position: relative;
  }
</style>