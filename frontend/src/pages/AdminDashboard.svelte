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
    pendingInternsList: [],
    pendingTasks: 0,
    attendanceTrend: [],
    submittedTasks: []
  });

  // Date for header
  let currentDate = new Date().toLocaleDateString('id-ID', { 
    weekday: 'long', 
    year: 'numeric', 
    month: 'long', 
    day: 'numeric' 
  });

  // Chart instances
  let taskPieChart = null;
  let attendanceTrendChart = null;

  async function fetchDashboardData() {
    loading = true;
    try {
      const pendingReq = api.getInterns({ status: 'pending', limit: 5 });
      const submittedTasksReq = api.getTasks({ status: 'submitted', limit: 5 });

      let serverStats = null;
      try {
        const res = await api.getAdminDashboard();
        serverStats = res.data;
      } catch (err) {
        console.warn("Backend dashboard endpoint belum siap/error (404).");
      }

      const [pendingRes, tasksRes] = await Promise.all([pendingReq, submittedTasksReq]);

      if (serverStats) {
        const stats = serverStats.stats || {};
        dashboardData = {
            totalInterns: stats.total_interns || 0,
            completedOnTime: stats.completed_on_time || 0,
            completedLate: stats.completed_late || 0,
            presentToday: stats.present_today || 0,
            pendingTasks: stats.pending_tasks || 0,
            attendanceTrend: serverStats.weekly_trend || [],
            pendingRegistrations: pendingRes.meta?.total || (pendingRes.data ? pendingRes.data.length : 0),
            pendingInternsList: pendingRes.data || [],
            submittedTasks: tasksRes.data || []
        };
      } else {
        const allInternsRes = await api.getInterns({ limit: 1000 });
        const allInterns = allInternsRes.data || [];
        const activeCount = allInterns.filter(i => i.status === 'active').length;
        
        let taskOnTime = 0, taskLate = 0, taskPending = 0;
        try {
            const allTasksRes = await api.getTasks({ limit: 1000 });
            const allTasks = allTasksRes.data || [];
            taskOnTime = allTasks.filter(t => t.status === 'completed' && !t.is_late).length;
            taskLate = allTasks.filter(t => t.status === 'completed' && t.is_late).length;
            taskPending = allTasks.filter(t => ['pending', 'in_progress'].includes(t.status)).length;
        } catch(e) { console.log('Gagal hitung task manual'); }

        dashboardData = {
            totalInterns: activeCount,
            completedOnTime: taskOnTime,
            completedLate: taskLate,
            presentToday: 0,
            pendingTasks: taskPending,
            attendanceTrend: [],
            pendingRegistrations: pendingRes.meta?.total || (pendingRes.data ? pendingRes.data.length : 0),
            pendingInternsList: pendingRes.data || [],
            submittedTasks: tasksRes.data || []
        };
      }

      $effect(() => {
        if (!loading) {
          setTimeout(initCharts, 100);
        }
      });

    } catch (err) {
      console.error('Fatal error fetching dashboard:', err);
    } finally {
      loading = false;
    }
  }

  async function handleApprove(id, name) {
    if (!confirm(`Terima siswa "${name}" menjadi Active?`)) return;
    try {
        await api.updateInternStatus(id, 'active');
        dashboardData.pendingInternsList = dashboardData.pendingInternsList.filter(i => i.id !== id);
        dashboardData.pendingRegistrations--;
        dashboardData.totalInterns++;
        alert(`Berhasil menerima ${name}`);
    } catch (err) {
        alert('Gagal approve: ' + (err.response?.data?.message || err.message));
    }
  }

  async function handleDeny(id, name) {
    if (!confirm(`Apakah Anda yakin ingin MENOLAK dan MENGHAPUS pendaftaran "${name}"? Data akan hilang permanen.`)) return;
    try {
        // Menggunakan generic delete endpoint, asumsi RESTful /interns/:id
        await api.delete(`/interns/${id}`);
        dashboardData.pendingInternsList = dashboardData.pendingInternsList.filter(i => i.id !== id);
        dashboardData.pendingRegistrations--;
        alert(`Pendaftaran ${name} telah ditolak dan dihapus.`);
    } catch (err) {
        alert('Gagal menolak: ' + (err.response?.data?.message || err.message));
    }
  }

  function initCharts() {
    const taskPieCtx = document.getElementById('taskPieChart');
    if (taskPieCtx && taskPieCtx instanceof HTMLCanvasElement) {
      if (taskPieChart) taskPieChart.destroy();
      taskPieChart = new Chart(taskPieCtx.getContext('2d'), {
        type: 'doughnut',
        data: {
          labels: ['Tepat Waktu', 'Terlambat', 'Dalam Proses'],
          datasets: [{
            data: [
              dashboardData.completedOnTime,
              dashboardData.completedLate,
              dashboardData.pendingTasks
            ],
            // Warna Hijau Modern: Emerald, Amber, Slate
            backgroundColor: ['#10b981', '#f59e0b', '#cbd5e1'],
            hoverBackgroundColor: ['#059669', '#d97706', '#94a3b8'],
            borderWidth: 2,
            borderColor: '#ffffff',
            hoverOffset: 5
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          cutout: '70%',
          layout: { padding: 10 },
          plugins: {
            legend: {
              position: 'bottom',
              labels: {
                usePointStyle: true,
                pointStyle: 'circle',
                padding: 20,
                font: { family: 'Geist, Inter, sans-serif', size: 12 },
                color: '#64748b'
              }
            }
          }
        }
      });
    }

    const trendCtx = document.getElementById('attendanceTrendChart');
    if (trendCtx && trendCtx instanceof HTMLCanvasElement && dashboardData.attendanceTrend.length > 0) {
      if (attendanceTrendChart) attendanceTrendChart.destroy();
      
      const labels = dashboardData.attendanceTrend.map(d => d.day);
      const presentData = dashboardData.attendanceTrend.map(d => d.present);
      const absentData = dashboardData.attendanceTrend.map(d => d.absent);

      attendanceTrendChart = new Chart(trendCtx.getContext('2d'), {
        type: 'bar',
        data: {
          labels: labels,
          datasets: [
            {
              label: 'Hadir',
              data: presentData,
              backgroundColor: '#10b981',
              hoverBackgroundColor: '#059669',
              borderRadius: 6,
              barThickness: 16
            },
            {
              label: 'Tidak Hadir',
              data: absentData,
              backgroundColor: '#fca5a5',
              hoverBackgroundColor: '#ef4444',
              borderRadius: 6,
              barThickness: 16
            }
          ]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: { display: false },
            tooltip: {
                backgroundColor: 'rgba(15, 23, 42, 0.9)',
                padding: 10,
                cornerRadius: 8,
                titleFont: { family: 'Geist, Inter, sans-serif' },
                bodyFont: { family: 'Geist, Inter, sans-serif' }
            }
          },
          scales: {
            x: { 
                stacked: true, 
                grid: { display: false },
                ticks: { font: { family: 'Geist, Inter, sans-serif' }, color: '#64748b' }
            },
            y: { 
                stacked: true, 
                beginAtZero: true, 
                grid: { color: '#f1f5f9' },
                ticks: { font: { family: 'Geist, Inter, sans-serif' }, color: '#64748b', stepSize: 1 }
            }
          }
        }
      });
    }
  }

  onMount(fetchDashboardData);
</script>

<div class="page-bg">
  <div class="dashboard-wrapper">
    <!-- Header with Settings Button -->
    <div class="header animate-fade-in">
      <div class="header-content">
        <h2 class="welcome-text">
          Halo, <span class="highlight">{auth.user?.name || 'Admin'}</span> 
        </h2>
        <p class="subtitle">
          Pantau progres dan aktivitas magang secara real-time.
        </p>
      </div>
      
      <div class="header-actions">
        <div class="date-pill">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
          {currentDate}
        </div>

        <a href="/settings" class="btn-settings">
            <div class="btn-icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.38a2 2 0 0 0-.73-2.73l-.15-.1a2 2 0 0 1-1-1.72v-.51a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0 2.73.73l.15.08a2 2 0 0 1 2 0l-.43.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"></path>
                  <circle cx="12" cy="12" r="3"></circle>
              </svg>
            </div>
            <span>Pengaturan</span>
        </a>
      </div>
    </div>

    {#if loading}
      <div class="loading-container">
        <div class="spinner"></div>
        <p>Menyiapkan data...</p>
      </div>
    {:else}
    
      <!-- Section Approval (Premium Gradient) -->
      {#if dashboardData.pendingInternsList.length > 0}
        <div class="approval-card animate-slide-up">
          <div class="approval-decoration"></div>
          <div class="approval-header">
              <div class="approval-title-group">
                  <div class="icon-pulse">
                      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                          <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                          <circle cx="8.5" cy="7" r="4"/>
                          <polyline points="17 11 19 13 23 9"/>
                      </svg>
                  </div>
                  <div>
                      <h3>Pendaftaran Baru</h3>
                      <p>Terdapat <span class="count-badge">{dashboardData.pendingRegistrations}</span> calon siswa menunggu persetujuan Anda.</p>
                  </div>
              </div>
              <a href="/interns?status=pending" class="link-view-all">
                Lihat Semua
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14"/><path d="M12 5l7 7-7 7"/></svg>
              </a>
          </div>
          
          <div class="approval-list">
              {#each dashboardData.pendingInternsList as intern}
                  <div class="approval-item">
                      <div class="item-info">
                          <div class="avatar-initial">
                              {intern.full_name.charAt(0)}
                          </div>
                          <div class="item-details">
                              <h4>{intern.full_name}</h4>
                              <span>{intern.School} • {intern.Department}</span>
                          </div>
                      </div>
                      <div class="approval-actions">
                          <button 
                              class="btn-deny"
                              onclick={() => handleDeny(intern.id, intern.full_name)}
                          >
                              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                              Tolak
                          </button>
                          <button 
                              class="btn-approve"
                              onclick={() => handleApprove(intern.id, intern.full_name)}
                          >
                              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
                              Terima
                          </button>
                      </div>
                  </div>
              {/each}
          </div>
        </div>
      {/if}

      <!-- Stats Grid (Modern Cards) -->
      <div class="stats-grid animate-slide-up" style="animation-delay: 0.1s;">
        
        <!-- Total Active -->
        <div class="stat-card">
          <div class="stat-icon-wrapper bg-green-soft">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
              <circle cx="9" cy="7" r="4"/>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
            </svg>
          </div>
          <div class="stat-content">
            <span class="label">Siswa Aktif</span>
            <span class="value">{dashboardData.totalInterns}</span>
          </div>
          <div class="stat-decoration bg-green"></div>
        </div>

        <!-- Hadir Hari Ini -->
        <div class="stat-card">
          <div class="stat-icon-wrapper bg-emerald-soft">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
              <polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
          </div>
          <div class="stat-content">
            <span class="label">Hadir Hari Ini</span>
            <span class="value">{dashboardData.presentToday}</span>
          </div>
          <div class="stat-decoration bg-emerald"></div>
        </div>

        <!-- Tepat Waktu -->
        <div class="stat-card">
          <div class="stat-icon-wrapper bg-teal-soft">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
          </div>
          <div class="stat-content">
            <span class="label">Tugas Tepat Waktu</span>
            <span class="value">{dashboardData.completedOnTime}</span>
          </div>
          <div class="stat-decoration bg-teal"></div>
        </div>

        <!-- Terlambat -->
        <div class="stat-card">
          <div class="stat-icon-wrapper bg-amber-soft">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
          </div>
          <div class="stat-content">
            <span class="label">Tugas Terlambat</span>
            <span class="value">{dashboardData.completedLate}</span>
          </div>
          <div class="stat-decoration bg-amber"></div>
        </div>
      </div>

      <!-- Charts Row -->
      <div class="charts-grid animate-slide-up" style="animation-delay: 0.2s;">
        
        <!-- Task Chart -->
        <div class="chart-card">
          <div class="card-header">
              <h3>Distribusi Tugas</h3>
              <span class="badge-soft">Total: {dashboardData.completedOnTime + dashboardData.completedLate + dashboardData.pendingTasks}</span>
          </div>
          <div class="canvas-wrapper">
            <canvas id="taskPieChart"></canvas>
          </div>
        </div>

        <!-- Attendance Chart -->
        <div class="chart-card">
          <div class="card-header">
              <h3>Tren Kehadiran</h3>
              <select class="chart-select">
                  <option>7 Hari Terakhir</option>
              </select>
          </div>
          {#if dashboardData.attendanceTrend.length > 0}
              <div class="canvas-wrapper">
                  <canvas id="attendanceTrendChart"></canvas>
              </div>
          {:else}
              <div class="empty-chart">
                  <p>Belum ada data kehadiran</p>
              </div>
          {/if}
        </div>
      </div>

      <!-- Submitted Tasks List -->
      {#if dashboardData.submittedTasks.length > 0}
        <div class="tasks-section animate-slide-up" style="animation-delay: 0.3s;">
          <div class="section-header">
              <div class="header-title">
                  <h3>Tugas Baru Dikumpulkan</h3>
                  <span class="badge-count">{dashboardData.submittedTasks.length}</span>
              </div>
              <a href="/tasks?status=submitted" class="link-view-all">Lihat Semua →</a>
          </div>
          <div class="task-grid">
              {#each dashboardData.submittedTasks as task}
                <div class="task-item">
                    <div class="task-icon">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
                    </div>
                    <div class="task-info">
                        <h4>{task.title}</h4>
                        <p>Oleh: <span class="author">{task.intern_name || 'Siswa Magang'}</span></p>
                    </div>
                    <a href="/task-assignments/{task.id}" class="btn-review">
                        Review
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
                    </a>
                </div>
              {/each}
          </div>
        </div>
      {/if}

    {/if}
  </div>
</div>

<style>
  /* --- FONTS & GLOBAL THEME --- */
  :global(body) {
    font-family: 'Geist', 'Inter', sans-serif;
    margin: 0;
    padding: 0;
  }

  .page-bg {
    min-height: 100vh;
    /* Soft premium background: White to very faint green tint */
    background: radial-gradient(circle at 0% 0%, #f0fdf4 0%, #f8fafc 40%, #ffffff 100%);
    background-attachment: fixed;
  }

  .dashboard-wrapper {
    max-width: 1280px;
    margin: 0 auto;
    padding: 40px 24px;
  }

  /* --- HEADER --- */
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 48px;
  }
  .welcome-text {
    font-size: 32px;
    font-weight: 800;
    margin: 0 0 8px 0;
    color: #0f172a;
    letter-spacing: -0.03em;
  }
  .highlight {
    background: linear-gradient(120deg, #10b981, #059669);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .subtitle {
    color: #64748b;
    font-size: 16px;
    font-weight: 500;
    margin: 0;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .date-pill {
    display: flex; align-items: center; gap: 8px;
    padding: 8px 16px;
    background: #ffffff; border: 1px solid #e2e8f0; border-radius: 9999px;
    color: #64748b; font-size: 14px; font-weight: 500;
    box-shadow: 0 2px 4px rgba(0,0,0,0.03);
  }
  
  .btn-settings {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 16px 8px 8px;
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 9999px; /* Pill shape */
    color: #475569;
    font-weight: 600;
    font-size: 14px;
    text-decoration: none;
    transition: all 0.3s ease;
    box-shadow: 0 2px 4px rgba(0,0,0,0.03);
  }
  .btn-settings .btn-icon {
    background: #f1f5f9;
    width: 32px; height: 32px;
    border-radius: 50%;
    display: flex; align-items: center; justify-content: center;
    transition: background 0.3s;
  }
  .btn-settings:hover {
    border-color: #10b981;
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.15);
    color: #0f172a;
  }
  .btn-settings:hover .btn-icon {
    background: #d1fae5;
    color: #059669;
  }

  /* --- APPROVAL CARD (PREMIUM LOOK) --- */
  .approval-card {
    background: #ffffff;
    border-radius: 24px;
    padding: 0;
    margin-bottom: 40px;
    border: 1px solid rgba(16, 185, 129, 0.2);
    box-shadow: 0 10px 30px -10px rgba(16, 185, 129, 0.15);
    position: relative;
    overflow: hidden;
  }
  .approval-decoration {
    position: absolute; top: 0; left: 0; width: 100%; height: 6px;
    background: linear-gradient(90deg, #10b981, #34d399);
  }
  .approval-header {
    padding: 24px 32px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: linear-gradient(to bottom, #ecfdf5, #ffffff);
    border-bottom: 1px solid #e2e8f0;
  }
  .approval-title-group { display: flex; align-items: center; gap: 16px; }
  .icon-pulse {
    width: 48px; height: 48px;
    background: #ffffff;
    color: #10b981;
    border: 1px solid #d1fae5;
    border-radius: 14px;
    display: flex; align-items: center; justify-content: center;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.1);
  }
  .approval-title-group h3 { margin: 0; font-size: 18px; font-weight: 600; color: #064e3b; }
  .approval-title-group p { margin: 4px 0 0 0; font-size: 14px; color: #64748b; }
  .count-badge { background: #fef3c7; color: #b45309; padding: 2px 8px; border-radius: 6px; font-weight: 600; font-size: 12px; border: 1px solid #fde68a; }

  .link-view-all {
    color: #059669; font-weight: 600; font-size: 14px; text-decoration: none;
    display: flex; align-items: center; gap: 6px;
    transition: gap 0.2s;
  }
  .link-view-all:hover { gap: 10px; color: #047857; }

  .approval-list { padding: 16px 24px; }
  .approval-item {
    display: flex; align-items: center; justify-content: space-between;
    padding: 16px; margin-bottom: 8px;
    border-radius: 16px;
    border: 1px solid transparent;
    transition: all 0.2s ease;
  }
  .approval-item:hover {
    background: #f8fafc;
    border-color: #e2e8f0;
    transform: scale(1.01);
  }
  .item-info { display: flex; align-items: center; gap: 16px; }
  .avatar-initial {
    width: 42px; height: 42px;
    background: linear-gradient(135deg, #10b981, #059669);
    color: white; border-radius: 12px;
    display: flex; align-items: center; justify-content: center;
    font-weight: 600; font-size: 16px;
    box-shadow: 0 4px 10px rgba(16, 185, 129, 0.3);
  }
  .item-details h4 { margin: 0; font-size: 15px; font-weight: 600; color: #1e293b; }
  .item-details span { font-size: 13px; color: #64748b; }

  .approval-actions { display: flex; gap: 8px; }

  .btn-approve {
    background: #ffffff; color: #10b981;
    border: 1px solid #10b981;
    padding: 8px 18px; border-radius: 10px;
    font-size: 13px; font-weight: 600; cursor: pointer;
    display: flex; align-items: center; gap: 6px;
    transition: all 0.2s;
  }
  .btn-approve:hover {
    background: #10b981; color: white;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
    transform: translateY(-1px);
  }

  .btn-deny {
    background: #ffffff; color: #ef4444;
    border: 1px solid #ef4444;
    padding: 8px 18px; border-radius: 10px;
    font-size: 13px; font-weight: 600; cursor: pointer;
    display: flex; align-items: center; gap: 6px;
    transition: all 0.2s;
  }
  .btn-deny:hover {
    background: #ef4444; color: white;
    box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
    transform: translateY(-1px);
  }

  /* --- STATS GRID --- */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 24px;
    margin-bottom: 40px;
  }
  .stat-card {
    background: white;
    padding: 24px;
    border-radius: 20px;
    border: 1px solid #f1f5f9;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.01), 0 2px 4px -1px rgba(0,0,0,0.01);
    display: flex;
    align-items: center;
    gap: 20px;
    position: relative;
    overflow: hidden;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }
  .stat-card:hover {
    transform: translateY(-6px);
    box-shadow: 0 20px 25px -5px rgba(0,0,0,0.05), 0 10px 10px -5px rgba(0,0,0,0.02);
    border-color: transparent;
  }
  .stat-icon-wrapper {
    width: 60px; height: 60px;
    border-radius: 18px;
    display: flex; align-items: center; justify-content: center;
    flex-shrink: 0;
  }
  
  /* Colors */
  .bg-green-soft { background: #ecfdf5; color: #10b981; }
  .bg-emerald-soft { background: #f0fdf4; color: #16a34a; }
  .bg-teal-soft { background: #f0f9ff; color: #0ea5e9; }
  .bg-amber-soft { background: #fffbeb; color: #f59e0b; }
  
  .bg-green { background: #10b981; }
  .bg-emerald { background: #16a34a; }
  .bg-teal { background: #0ea5e9; }
  .bg-amber { background: #f59e0b; }

  .stat-content { display: flex; flex-direction: column; z-index: 1; }
  .stat-content .value { font-size: 32px; font-weight: 800; color: #0f172a; line-height: 1.1; letter-spacing: -0.02em; }
  .stat-content .label { font-size: 13px; font-weight: 600; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; margin-bottom: 4px; }
  
  /* Hover Effect: Bottom Line Color */
  .stat-decoration {
    position: absolute; bottom: 0; left: 0; width: 100%; height: 4px;
    opacity: 0; transition: opacity 0.3s;
  }
  .stat-card:hover .stat-decoration { opacity: 1; }

  /* --- CHARTS GRID --- */
  .charts-grid {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 24px;
    margin-bottom: 40px;
  }
  @media (max-width: 1024px) { .charts-grid { grid-template-columns: 1fr; } }

  .chart-card {
    background: white;
    padding: 28px;
    border-radius: 24px;
    border: 1px solid #f1f5f9;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.01);
    transition: transform 0.3s;
  }
  .chart-card:hover { transform: translateY(-2px); box-shadow: 0 10px 15px -3px rgba(0,0,0,0.05); }
  
  .card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
  .card-header h3 { margin: 0; font-size: 18px; font-weight: 600; color: #1e293b; }
  .badge-soft { background: #f1f5f9; color: #64748b; padding: 4px 12px; border-radius: 99px; font-size: 12px; font-weight: 600; }
  .chart-select { border: 1px solid #e2e8f0; border-radius: 8px; padding: 4px 8px; font-size: 12px; color: #64748b; background: transparent; cursor: pointer; }

  .canvas-wrapper { height: 300px; position: relative; }
  .empty-chart { height: 300px; display: flex; align-items: center; justify-content: center; background: #f8fafc; border-radius: 16px; border: 2px dashed #e2e8f0; color: #94a3b8; font-size: 14px; font-weight: 500; }

  /* --- TASKS SECTION --- */
  .tasks-section {
    background: white;
    border-radius: 24px;
    padding: 32px;
    border: 1px solid #f1f5f9;
    box-shadow: 0 10px 15px -3px rgba(0,0,0,0.03);
  }
  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
  }
  .header-title { display: flex; align-items: center; gap: 12px; }
  .header-title h3 { margin: 0; font-size: 20px; font-weight: 600; color: #1e293b; }
  .badge-count { background: #ef4444; color: white; width: 24px; height: 24px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 600; }

  .task-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 20px; }
  .task-item {
    display: flex; align-items: center; gap: 16px;
    padding: 20px;
    background: #ffffff;
    border: 1px solid #f1f5f9;
    border-radius: 16px;
    transition: all 0.3s ease;
    position: relative;
  }
  .task-item:hover {
    border-color: #cbd5e1;
    transform: translateY(-4px);
    box-shadow: 0 10px 20px -5px rgba(0,0,0,0.05);
  }
  .task-icon {
    width: 48px; height: 48px;
    background: #f8fafc; color: #64748b;
    border-radius: 12px;
    display: flex; align-items: center; justify-content: center;
    transition: all 0.3s;
  }
  .task-item:hover .task-icon { background: #e0f2fe; color: #0ea5e9; }
  
  .task-info { flex: 1; }
  .task-info h4 { margin: 0 0 4px 0; font-size: 16px; font-weight: 600; color: #334155; }
  .task-info p { margin: 0; font-size: 13px; color: #94a3b8; }
  .task-info .author { color: #64748b; font-weight: 500; }

  .btn-review {
    padding: 8px 16px;
    background: #f1f5f9;
    color: #475569;
    border-radius: 10px;
    text-decoration: none;
    font-size: 13px;
    font-weight: 600;
    display: flex; align-items: center; gap: 6px;
    transition: all 0.2s;
  }
  .btn-review:hover {
    background: #0f172a; color: white;
  }

  /* --- LOADING SPINNER --- */
  .spinner {
    width: 48px; height: 48px;
    border: 4px solid #e2e8f0;
    border-top-color: #10b981;
    border-radius: 50%;
    margin: 0 auto 16px;
    animation: spin 1s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* --- ANIMATIONS --- */
  .animate-fade-in { opacity: 0; animation: fadeIn 0.8s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards; }

  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(30px); } to { opacity: 1; transform: translateY(0); } }
</style>