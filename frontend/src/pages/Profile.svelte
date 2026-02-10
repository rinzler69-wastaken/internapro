<script>
  import { onMount, onDestroy, tick } from 'svelte';
  import Chart from 'chart.js/auto';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let loading = $state(true);
  let error = $state(null);

  let user = $state(null);
  let intern = $state(null);
  let supervisor = $state(null);

  let stats = $state({
    completedTasks: 0,
    totalTasks: 0,
    attendancePercentage: 0,
    averageSpeed: 0,
    overallScore: 0
  });

  let taskStatusData = $state({
    completed: 0,
    in_progress: 0,
    pending: 0,
    revision: 0
  });

  let attendanceData = $state({
    present: 0,
    late: 0,
    absent: 0,
    sick: 0,
    permission: 0
  });

  let assessmentData = $state(null);

  let taskChartEl;
  let attendanceChartEl;
  let radarChartEl;
  let taskChart, attendanceChart, radarChart;

  /** @type {import('chart.js').ChartOptions<'doughnut'>} */
  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'bottom',
        labels: {
          color: '#475569',
          font: { size: 12, weight: 600 },
          usePointStyle: true,
          padding: 18
        }
      }
    },
    cutout: '70%',
    layout: { padding: 18 }
  };

  onMount(load);
  onDestroy(cleanupCharts);

  async function load() {
    loading = true;
    error = null;
    cleanupCharts();

    try {
      const res = await api.getProfile();
      const payload = res.data || res;
      user = payload.user || payload;
      intern = payload.intern && payload.intern.id ? payload.intern : null;
      supervisor = payload.supervisor && payload.supervisor.id ? payload.supervisor : null;

      // Enrich intern details (supervisor name, institution, etc.)
      if (intern?.id) {
        try {
          const detail = await api.getIntern(intern.id);
          const detailData = detail.data || detail;
          intern = { ...intern, ...detailData, supervisor: detailData.supervisor || null, supervisor_name: detailData.supervisor_name || detailData.SupervisorName };
        } catch (err) {
          console.warn('Failed to fetch intern detail', err);
        }
      }

      // Enrich supervisor details if needed
      if (supervisor?.id && !supervisor.nip) {
        try {
          const detail = await api.getSupervisors({ id: supervisor.id });
          const detailData = detail.data?.[0] || detail.data || detail;
          supervisor = { ...supervisor, ...detailData };
        } catch (err) {
          console.warn('Failed to fetch supervisor detail', err);
        }
      }

      if (auth.hydrate && user) auth.hydrate(user);

      if (intern) {
        await loadInternMetrics();
      }
    } catch (err) {
      console.error(err);
      error = err.message || 'Gagal memuat profil';
    } finally {
      loading = false;
      await tick();
      if (intern) drawCharts();
    }
  }

  async function loadInternMetrics() {
    let tasks = [];
    let attendance = [];
    let assessments = [];
    let dashboardData = null;

    try {
      const results = await Promise.allSettled([
        api.get('/dashboard/intern'),
        api.getTasks({ limit: 400 }),
        api.getAttendance({ limit: 400 }),
        api.getAssessments({ limit: 30 })
      ]);

      if (results[0].status === 'fulfilled') dashboardData = results[0].value.data?.data || results[0].value.data || results[0].value;
      if (results[1].status === 'fulfilled') tasks = results[1].value.data?.data || results[1].value.data || results[1].value || [];
      if (results[2].status === 'fulfilled') attendance = results[2].value.data?.data || results[2].value.data || results[2].value.records || results[2].value || [];
      if (results[3].status === 'fulfilled') assessments = results[3].value.data?.data || results[3].value.data || results[3].value || [];
    } catch (err) {
      console.warn('Partial profile metrics failed:', err);
    }

    deriveStats({ tasks, attendance, assessments, dashboard: dashboardData });
  }

  function deriveStats({ tasks, attendance, assessments, dashboard }) {
    const tasksArr = Array.isArray(tasks) ? tasks : [];
    const attendanceArr = Array.isArray(attendance) ? attendance : [];
    const assessmentsArr = Array.isArray(assessments) ? assessments : [];

    const totalTasks = tasksArr.length;
    const completedTasks = tasksArr.filter((t) => t.status === 'completed').length;
    const inProgress = tasksArr.filter((t) => t.status === 'in_progress').length;
    const pending = tasksArr.filter((t) => t.status === 'pending').length;
    const revision = tasksArr.filter((t) => t.status === 'revision').length;

    const completedArr = tasksArr.filter((t) => t.status === 'completed');
    const onTime = completedArr.filter(
      (t) =>
        t.is_late === false ||
        t.is_late === 0 ||
        t.is_late === null ||
        t.is_late === undefined
    );
    const averageSpeed = dashboard?.task_stats?.percentage ?? (completedTasks
      ? Math.round((onTime.length / completedArr.length) * 100)
      : 0);

    const present = attendanceArr.filter((a) => a.status === 'present').length;
    const late = attendanceArr.filter((a) => a.status === 'late').length;
    const absent = attendanceArr.filter((a) => a.status === 'absent').length;
    const sick = attendanceArr.filter((a) => a.status === 'sick').length;
    const permission = attendanceArr.filter((a) => a.status === 'permission').length;
    const attTotal = attendanceArr.length || 1;
    const attendancePercentage =
      dashboard?.attendance_percentage ?? Math.round(((present + late) / attTotal) * 100);

    const radar = buildAssessmentRadar(assessmentsArr);

    const tasksWithScores = tasksArr.filter((t) => t.score !== null && t.score !== undefined && t.score !== '');
    const taskAverageScore = tasksWithScores.length
      ? Math.round(tasksWithScores.reduce((acc, t) => acc + Number(t.score), 0) / tasksWithScores.length)
      : 0;

    stats = {
      completedTasks: dashboard?.task_stats?.completed ?? dashboard?.completed ?? completedTasks,
      totalTasks: dashboard?.task_stats?.total ?? dashboard?.total ?? totalTasks,
      attendancePercentage,
      averageSpeed,
      overallScore: taskAverageScore
    };
    taskStatusData = dashboard?.task_breakdown
      ? {
          completed: dashboard.task_breakdown.completed ?? 0,
          in_progress: dashboard.task_breakdown.in_progress ?? 0,
          pending: dashboard.task_breakdown.pending ?? 0,
          revision: dashboard.task_breakdown.revision ?? 0
        }
      : { completed: completedTasks, in_progress: inProgress, pending, revision };
    attendanceData = { present, late, absent, sick, permission };
    assessmentData = radar;
  }

  function buildAssessmentRadar(list = []) {
    if (!list.length) return null;
    const keys = [
      'quality_score',
      'speed_score',
      'initiative_score',
      'teamwork_score',
      'communication_score'
    ];
    const labelMap = {
      quality_score: 'Kualitas',
      speed_score: 'Kecepatan',
      initiative_score: 'Inisiatif',
      teamwork_score: 'Kerjasama',
      communication_score: 'Komunikasi'
    };

    const sum = {};
    const count = {};
    keys.forEach((k) => {
      sum[k] = 0;
      count[k] = 0;
    });

    list.slice(0, 5).forEach((a) => {
      keys.forEach((k) => {
        const v = a[k];
        if (v !== null && v !== undefined) {
          sum[k] += Number(v);
          count[k] += 1;
        }
      });
    });

    const labels = [];
    const values = [];
    keys.forEach((k) => {
      labels.push(labelMap[k]);
      values.push(count[k] ? Math.round(sum[k] / count[k]) : 0);
    });

    const overallScore = values.length
      ? Math.round(values.reduce((a, b) => a + b, 0) / values.length)
      : 0;

    return { labels, values, overallScore };
  }

  function drawCharts() {
    cleanupCharts();

    if (taskChartEl) {
      taskChart = new Chart(taskChartEl, {
        type: 'doughnut',
        data: {
          labels: ['Selesai', 'Proses', 'Menunggu', 'Revisi'],
          datasets: [
            {
              data: [
                taskStatusData.completed,
                taskStatusData.in_progress,
                taskStatusData.pending,
                taskStatusData.revision
              ],
              backgroundColor: ['#10b981', '#6366f1', '#94a3b8', '#f59e0b'],
              borderWidth: 0
            }
          ]
        },
        options: chartOptions
      });
    }

    if (attendanceChartEl) {
      attendanceChart = new Chart(attendanceChartEl, {
        type: 'doughnut',
        data: {
          labels: ['Hadir', 'Telat', 'Absen', 'Sakit', 'Izin'],
          datasets: [
            {
              data: [
                attendanceData.present,
                attendanceData.late,
                attendanceData.absent,
                attendanceData.sick,
                attendanceData.permission
              ],
              backgroundColor: ['#10b981', '#f59e0b', '#ef4444', '#06b6d4', '#6366f1'],
              borderWidth: 0
            }
          ]
        },
        options: chartOptions
      });
    }

    if (assessmentData && radarChartEl) {
      radarChart = new Chart(radarChartEl, {
        type: 'radar',
        data: {
          labels: assessmentData.labels,
          datasets: [
            {
              label: 'Skor Rata-rata',
              data: assessmentData.values,
              backgroundColor: 'rgba(99, 102, 241, 0.2)',
              borderColor: 'rgba(99, 102, 241, 1)',
              borderWidth: 2,
              pointBackgroundColor: 'rgba(99, 102, 241, 1)'
            }
          ]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          scales: {
            r: {
              beginAtZero: true,
              max: 100,
              grid: { color: 'rgba(148, 163, 184, 0.25)' },
              angleLines: { color: 'rgba(148, 163, 184, 0.25)' },
              pointLabels: {
                font: { size: 12, weight: 600 },
                color: '#475569'
              },
              ticks: { backdropColor: 'transparent', color: '#94a3b8', stepSize: 20 }
            }
          },
          plugins: { legend: { display: false } }
        }
      });
    }
  }

  function cleanupCharts() {
    taskChart && taskChart.destroy();
    attendanceChart && attendanceChart.destroy();
    radarChart && radarChart.destroy();
    taskChart = attendanceChart = radarChart = null;
  }

  function formatDate(dateStr) {
    if (!dateStr) return '-';
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return '-';
    return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  function daysBetween(startStr, endStr) {
    if (!startStr || !endStr) return '-';
    const start = new Date(startStr);
    const end = new Date(endStr);
    if (isNaN(start.getTime()) || isNaN(end.getTime())) return '-';
    const diffMs = end.getTime() - start.getTime();
    return Math.max(0, Math.round(diffMs / (1000 * 60 * 60 * 24)));
  }

  function remainingDays(endStr) {
    if (!endStr) return null;
    const end = new Date(endStr);
    if (isNaN(end.getTime())) return null;
    const diff = Math.round((end.getTime() - Date.now()) / (1000 * 60 * 60 * 24));
    return diff;
  }

  function statusBadge(status) {
    if (status === 'active') return { text: 'Aktif', cls: 'badge-success' };
    if (status === 'completed') return { text: 'Selesai', cls: 'badge-primary' };
    if (status === 'cancelled') return { text: 'Dibatalkan', cls: 'badge-danger' };
    return { text: status || '-', cls: 'badge-muted' };
  }

  function avatarUrl(path) {
    if (!path) return null;
    if (path.startsWith('http')) return path;
    const base = path.startsWith('/uploads/') ? path : `/uploads/${path}`;
    const qs = [];
    if (auth.token) qs.push(`token=${auth.token}`);
    qs.push(`t=${Date.now()}`);
    return `${base}${base.includes('?') ? '&' : '?'}${qs.join('&')}`;
  }

  const isSupervisor = $derived(user?.role === 'supervisor' || user?.role === 'pembimbing');

  const personalInfo = $derived(
    intern
      ? [
          { label: 'NIS/NIM', value: intern.nis || intern.student_id || intern.nim || '-' },
          { label: 'Asal Sekolah', value: intern.school || '-' },
          { label: 'Jurusan', value: intern.department || '-' },
          { label: 'No. Telepon', value: intern.phone || '-' },
          { label: 'Alamat', value: intern.address || '-' },
          { label: 'Pembimbing', value: intern.supervisor_name || (intern.supervisor?.full_name || intern.supervisor?.name) || '-' },
          { label: 'Institusi', value: intern.institution_name || (typeof intern.institution === 'string' ? intern.institution : intern.institution?.name) || '-' }
        ]
      : [
          { label: 'Nama', value: user?.name },
          { label: 'Email', value: user?.email },
          { label: 'Role', value: user?.role },
          { label: 'Bergabung', value: formatDate(user?.created_at) }
        ]
  );

  const magangInfo = $derived(
    intern
      ? [
          { label: 'Periode', value: `${formatDate(intern.start_date)} - ${formatDate(intern.end_date)}` },
          { label: 'Durasi', value: `${daysBetween(intern.start_date, intern.end_date)} Hari` },
          {
            label: 'Sisa Waktu',
            value: (() => {
              const left = remainingDays(intern.end_date);
              if (left === null) return '-';
              if (left < 0) return 'Telah Berakhir';
              return `${left} Hari Lagi`;
            })()
          },
          { label: 'Status', value: statusBadge(intern.status).text }
        ]
      : [
          { label: 'Peran', value: user?.role },
          { label: 'Status', value: 'Aktif' },
          { label: 'Bergabung', value: formatDate(user?.created_at) },
          { label: 'Institusi', value: user?.institution_name || '-' }
        ]
  );

  function displayValue(val) {
    if (val === undefined || val === null) return '-';
    if (val instanceof Date) return formatDate(val.toISOString());
    if (typeof val === 'object') {
      if (val.full_name) return String(val.full_name);
      if (val.name) return String(val.name);
      if (val.title) return String(val.title);
      const firstString = Object.values(val).find((v) => typeof v === 'string');
      return firstString ? String(firstString) : '-';
    }
    if (val === '') return '-';
    return String(val);
  }
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
  <link
    href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap"
    rel="stylesheet"
  />
      <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@48,300,0,0" />
</svelte:head>

<div class="page">
  <div class="shell">
    <header class="header">
      <div class="title-stack">
        <!-- <h1>Profil Saya</h1> -->
      </div>
      <!-- <div class="actions"> -->
        <!-- <a class="btn ghost" href="/dashboard">Kembali</a> -->
      <!-- </div> -->
    </header>

    {#if loading}
      <div class="loader-card">
        <div class="spinner"></div>
        <p>Memuat profil...</p>
      </div>
    {:else if error}
      <div class="error-card">
        <p>{error}</p>
        <button class="btn primary" on:click={load}>Coba Lagi</button>
      </div>
    {:else}
      <div class="card hero">
        <div class="hero-left">
          <div class="avatar">
            {#if avatarUrl(user?.avatar)}
              <img src={avatarUrl(user?.avatar)} alt="avatar" referrerpolicy="no-referrer" />
            {:else}
              <div class="avatar-placeholder">{user?.name?.[0]?.toUpperCase() || 'U'}</div>
            {/if}
          </div>
          <div>
            <h2>{user?.name}</h2>
            <p class="muted">{user?.email}</p>
            <div class="pill-row">
              <span class="pill role">{user?.role}</span>
              {#if intern}
                {#if statusBadge(intern.status)}
                  <span class={`pill ${statusBadge(intern.status).cls}`}>
                    {statusBadge(intern.status).text}
                  </span>
                {/if}
              {:else if supervisor}
                <span class="pill badge-success">Aktif</span>
              {/if}
            </div>
          </div>
        </div>
        <div class="hero-right">
          <div class="joined">
            <p class="muted">Bergabung sejak</p>
            <p class="muted-bold">{formatDate(user?.created_at)}</p>
          </div>
                <!-- <div class="actions"> -->
                  <a class="btn ghost" href="/profile/edit">Edit Profil</a>
        </div>
      </div>

      {#if intern}
        <!-- Stats -->
        <div class="grid stats">
          <div class="stat-card indigo">
            <div class="stat-label">Tugas Selesai</div>
            <div class="stat-value">
              {stats.completedTasks}
              <span class="stat-sub">/ {stats.totalTasks}</span>
            </div>
            <div class="progress"><div style={`width:${stats.totalTasks ? (stats.completedTasks / stats.totalTasks) * 100 : 0}%`}></div></div>
          </div>
          <div class="stat-card green">
            <div class="stat-label">Kehadiran</div>
            <div class="stat-value">{stats.attendancePercentage}%</div>
            <div class="progress"><div style={`width:${stats.attendancePercentage}%`}></div></div>
          </div>
          <div class="stat-card sky">
            <div class="stat-label">Kecepatan</div>
            <div class="stat-value">{stats.averageSpeed}%</div>
            <div class="progress"><div style={`width:${Math.min(stats.averageSpeed, 100)}%`}></div></div>
          </div>
          <div class="stat-card amber">
            <div class="stat-label">Skor Rata-rata</div>
            <div class="stat-value">{stats.overallScore}</div>
            <div class="progress"><div style={`width:${Math.min(stats.overallScore, 100)}%`}></div></div>
          </div>
        </div>

        <!-- Charts -->
        <div class="grid charts">
          <div class="card chart-card">
            <div class="card-head">
              <h3>Status Pekerjaan</h3>
              <!-- <span class="badge blue">Doughnut</span> -->
            </div>
            <div class="chart-wrap">
              <canvas bind:this={taskChartEl}></canvas>
            </div>
          </div>
          <div class="card chart-card">
            <div class="card-head">
              <h3>Status Kehadiran</h3>
              <!-- <span class="badge green">Doughnut</span> -->
            </div>
            <div class="chart-wrap">
              <canvas bind:this={attendanceChartEl}></canvas>
            </div>
          </div>
        </div>

        {#if assessmentData}
          <div class="card chart-card radar">
            <div class="card-head">
              <h3>Radar Penilaian (5 terbaru)</h3>
              <span class="badge indigo">Radar</span>
            </div>
            <div class="chart-wrap tall">
              <canvas bind:this={radarChartEl}></canvas>
            </div>
          </div>
        {/if}

      {/if}

      {#if intern || isSupervisor}
        <!-- Details -->
        <div class="grid details">
          <div class="card">
            <div class="card-head">
              <h3>Informasi Pribadi</h3>
            </div>
            <div class="detail-list">
              {#each personalInfo as item}
                <div class="detail-row">
                  <span class="detail-label">{item.label}</span>
                  <span class="detail-value">{displayValue(item.value)}</span>
                </div>
              {/each}
            </div>
          </div>

          <div class="card">
            <div class="card-head">
              <h3>{intern ? 'Informasi Magang' : 'Informasi Tambahan'}</h3>
            </div>
            <div class="detail-list">
              {#each magangInfo as item}
                <div class="detail-row">
                  <span class="detail-label">{item.label}</span>
                  <span class="detail-value">{displayValue(item.value)}</span>
                </div>
              {/each}
            </div>
          </div>
        </div>
      {:else}
        <div class="card">
          <div class="card-head">
            <h3>Informasi Akun</h3>
          </div>
          <div class="detail-list">
            <div class="detail-row"><span class="detail-label">Nama</span><span class="detail-value">{displayValue(user?.name)}</span></div>
            <div class="detail-row"><span class="detail-label">Email</span><span class="detail-value">{displayValue(user?.email)}</span></div>
            <div class="detail-row"><span class="detail-label">Role</span><span class="detail-value">{displayValue(user?.role)}</span></div>
            <div class="detail-row"><span class="detail-label">Dibuat</span><span class="detail-value">{formatDate(user?.created_at)}</span></div>
          </div>
        </div>
      {/if}
    {/if}
  </div>
</div>

<style>
  :global(body) {
    font-family: 'Plus Jakarta Sans', 'Inter', system-ui, -apple-system, sans-serif;
    background: #f8fafc;
    color: #0f172a;
  }

  .page {
    min-height: 100vh;
    background:
      radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.06) 0, transparent 45%),
      radial-gradient(at 100% 20%, rgba(16, 185, 129, 0.07) 0, transparent 35%),
      #f8fafc;
    padding: 0px;
  }

  .shell {
    max-width: 1200px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 12px;
  }

  .title-stack h1 {
    margin: 4px 0 6px;
    font-size: 20px;
    font-weight: 600;
    letter-spacing: -0.02em;
  }

  .eyebrow {
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-size: 11px;
    font-weight: 700;
    color: #6366f1;
    margin: 0;
  }

  .sub {
    margin: 0;
    color: #64748b;
  }

  .actions {
    display: flex;
    gap: 10px;
  }

  .btn {
    padding: 10px 0px;
    border-radius: 999px;
    font-weight: 700;
    border: 1px solid transparent;
    cursor: pointer;
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    transition: transform 0.12s ease, box-shadow 0.12s ease, background 0.12s ease;
    font-size: 14px;
  }

  .btn.primary {
    color:oklch(12.9% 0.042 264.695);
    background: oklch(96.8% 0.007 247.896);
    border-color: oklch(82.3% 0.034 264.695);
    /* box-shadow: 0 10px 30px rgba(16, 185, 129, 0.25); */
  }

  .btn.ghost {
    background: #fff;
    border: 2px solid #e2e8f0;
    color: #0f172a;
    font-weight: 600;
    display: inline-flex;
  }

  .btn:hover {
    transform: translateY(-1px);
  }

  .card {
    background: rgba(255, 255, 255, 0.9);
    border: 1px solid oklch(92.9% 0.013 255.508);
    border-radius: 20px;
    box-shadow: 0 10px 30px rgba(15, 23, 42, 0.04);
    padding: 20px;
  }

  .card-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 14px;
  }

  .card-head h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 700;
    color: #0f172a;
  }

  .hero {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 22px;
    background: linear-gradient(135deg, rgba(99, 102, 241, 0.08), rgba(16, 185, 129, 0.08));
  }

  .hero-left {
    display: flex;
    gap: 16px;
    align-items: center;
  }

  .hero-right {
    display: grid;
    grid-template-columns: repeat(1, minmax(0, 1fr));
    gap: 16px;
  }

  .hero-right .btn {
  grid-column: 1 / -1;   /* span full width */
  /* center horizontally */
  align-self: center;    /* center vertically */
}


  .avatar {
    width: 78px;
    height: 78px;
    border-radius: 50%;
    overflow: hidden;
    border: 4px solid #ffffff;
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
    background: #0f172a;
    font-weight: 800;
  }

  .avatar img,
  .avatar-placeholder {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 28px;
    font-weight: 800;
  }

  @media (max-width: 800px){
    .joined {
    display: flex;
    gap: 6px;
    align-items: baseline;
    justify-content: center;
    text-align: center;
    white-space: nowrap;
    margin-top: 8px;
  }
  }

  .joined .muted,
  .joined .big {
    margin: 0;
  }

  @media (max-width: 640px) {
  .hero-right {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }
  }

  @media (max-width: 640px) {
  .hero-right .btn {
    width: 100%;
    justify-content: center;
  }
}

  .muted {
    color: #64748b;
    margin: 0;
    font-size: 14px;
    text-align: center;
  }

  .muted-bold {
    color: #475569;
    margin: 0;
    font-size: 14px;
    font-weight: 600;
    text-align: center;
  }

  .pill-row {
    display: flex;
    gap: 8px;
    margin-top: 8px;
    flex-wrap: wrap;
  }

  .pill {
    padding: 6px 12px;
    border-radius: 999px;
    font-weight: 700;
    font-size: 12px;
    text-transform: capitalize;
    border: 1px solid transparent;
  }

  .pill.role {
    background: #eef2ff;
    color: #4f46e5;
    border-color: #e0e7ff;
  }

  .pill.badge-success {
    background: #ecfdf3;
    color: #047857;
    border-color: #bbf7d0;
  }

  .pill.badge-primary {
    background: #eef2ff;
    color: #4338ca;
    border-color: #c7d2fe;
  }

  .pill.badge-danger {
    background: #fef2f2;
    color: #b91c1c;
    border-color: #fecaca;
  }

  .pill.badge-muted {
    background: #f8fafc;
    color: #475569;
    border-color: #e2e8f0;
  }

  .pill.green {
    background: #ecfdf3;
    color: #047857;
    border-color: #bbf7d0;
  }

  .pill.blue {
    background: #e0f2fe;
    color: #075985;
    border-color: #bae6fd;
  }

  .pill.indigo {
    background: #eef2ff;
    color: #4338ca;
    border-color: #c7d2fe;
  }

  .grid {
    display: grid;
    gap: 16px;
  }

  .stats {
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  }

  .charts {
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  }

  .details {
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  }

  .stat-card {
    border-radius: 18px;
    padding: 18px;
    color: #0f172a;
    border: 1px solid rgba(15, 23, 42, 0.05);
    box-shadow: 0 14px 35px rgba(15, 23, 42, 0.06);
    background: #fff;
    position: relative;
    overflow: hidden;
  }

  .stat-card::after {
    content: '';
    position: absolute;
    inset: 0;
    background: radial-gradient(circle at 80% 20%, rgba(255, 255, 255, 0.6), transparent 55%);
    pointer-events: none;
  }

  .stat-card .stat-label {
    font-size: 12px;
    font-weight: 700;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .stat-card .stat-value {
    font-size: 28px;
    font-weight: 800;
    margin: 6px 0;
    display: flex;
    align-items: baseline;
    gap: 6px;
  }

  .stat-sub {
    font-size: 14px;
    color: #94a3b8;
    font-weight: 600;
  }

  .stat-card .progress {
    height: 8px;
    background: #f1f5f9;
    border-radius: 999px;
    overflow: hidden;
  }

  .stat-card .progress div {
    height: 100%;
    background: linear-gradient(90deg, #10b981, #059669);
  }

  .stat-card.indigo .progress div {
    background: linear-gradient(90deg, #6366f1, #4338ca);
  }

  .stat-card.green .progress div {
    background: linear-gradient(90deg, #10b981, #16a34a);
  }

  .stat-card.sky .progress div {
    background: linear-gradient(90deg, #06b6d4, #0ea5e9);
  }

  .stat-card.amber .progress div {
    background: linear-gradient(90deg, #f59e0b, #d97706);
  }

  .chart-card {
    padding: 20px;
  }

  .chart-wrap {
    height: 280px;
  }

  .chart-wrap.tall {
    height: 340px;
  }

  .badge {
    padding: 6px 10px;
    border-radius: 999px;
    font-weight: 700;
    font-size: 12px;
    border: 1px solid transparent;
  }

  .badge.blue {
    background: #e0f2fe;
    color: #075985;
    border-color: #bae6fd;
  }

  .badge.green {
    background: #ecfdf3;
    color: #047857;
    border-color: #bbf7d0;
  }

  .badge.indigo {
    background: #eef2ff;
    color: #4338ca;
    border-color: #c7d2fe;
  }

  .detail-list {
    display: grid;
    gap: 14px;
  }

  .detail-row {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    border-bottom: 1px dashed #e2e8f0;
    padding-bottom: 10px;
  }

  .detail-label {
    font-size: 12px;
    font-weight: 700;
    color: #94a3b8;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .detail-value {
    font-weight: 500;
    color: #0f172a;
    text-align: right;
    font-size: 12px;
  }

  .loader-card,
  .error-card {
    padding: 32px;
    text-align: center;
    background: #fff;
    border-radius: 16px;
    border: 1px solid #e2e8f0;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border-radius: 999px;
    border: 4px solid #e2e8f0;
    border-top-color: #6366f1;
    animation: spin 1s linear infinite;
    margin: 0 auto 12px;
  }

  .big {
    font-size: 16px;
    font-weight: 800;
    margin: 4px 0 0;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  @media (max-width: 900px) {
    .header {
      flex-direction: column;
      align-items: flex-start;
    }
    .hero {
      flex-direction: column;
      align-items: flex-start;
      gap: 16px;
    }
    .hero-right {
      width: 100%;
      grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    }
  }
</style>
