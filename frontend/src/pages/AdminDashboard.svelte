<script>
  import { onMount, tick } from "svelte";
  import { api } from "../lib/api.js";
  import { auth } from "../lib/auth.svelte.js";
  import Chart from "chart.js/auto";

  // State
  let loading = $state(true);
  let dashboardData = $state({
    totalInterns: 0,
    completedOnTime: 0,
    completedLate: 0,
    presentToday: 0,
    pendingRegistrations: 0,
    pendingInternsList: [],
    pendingSupervisors: 0,
    pendingSupervisorsList: [],
    overdueTasks: 0,
    inProgressTasks: 0,
    // completedLate: 0,
    attendanceTrend: [],
    submittedTasks: [],
  });

  // Date for header
  let currentDate = new Date().toLocaleDateString("id-ID", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
  });

  // Chart instances
  let taskPieChart = null;
  let attendanceTrendChart = null;

  const presentStatuses = ["present", "late"];
  const absentStatuses = ["absent", "permission", "sick", "on_leave", "leave"];

  function buildAttendanceTrend(records = []) {
    // Aggregate last 7 days (including today)
    const today = new Date();
    const dayKey = (d) => d.toISOString().slice(0, 10);
    const normalizeStatus = (s = "") => s.toLowerCase();

    const buckets = [];
    for (let i = 6; i >= 0; i--) {
      const d = new Date(today);
      d.setDate(today.getDate() - i);
      buckets.push({
        key: dayKey(d),
        label: d.toLocaleDateString("id-ID", {
          weekday: "short",
          day: "2-digit",
        }),
        present: 0,
        absent: 0,
      });
    }

    records.forEach((rec) => {
      const key = rec.date ? rec.date.slice(0, 10) : null;
      if (!key) return;
      const bucket = buckets.find((b) => b.key === key);
      if (!bucket) return;
      const status = normalizeStatus(rec.status);
      if (presentStatuses.includes(status)) bucket.present += 1;
      else if (absentStatuses.includes(status)) bucket.absent += 1;
    });

    return buckets.map(({ label, present, absent }) => ({
      day: label,
      present,
      absent,
    }));
  }

  // Redraw charts whenever dashboardData changes and we aren't loading
  $effect(() => {
    if (!loading && dashboardData) {
      tick().then(initCharts);
    }
  });

  async function fetchDashboardData() {
    loading = true;
    try {
      const pendingReq = api.getInterns({
        status: "pending",
        limit: 5,
        ...(auth.user?.role === "supervisor" || auth.user?.role === "pembimbing"
          ? { supervisor_id: auth.user.id }
          : {}),
      });

      // Only admins can fetch pending supervisors
      const pendingSupervisorsReq =
        auth.user?.role === "admin"
          ? api.getSupervisors({ status: "pending", limit: 5 })
          : Promise.resolve({ data: [], meta: { total: 0 } });

      const submittedTasksReq = api.getTasks({
        status: "submitted",
        limit: 15,
      });
      const attendanceReq = api.getAttendance({ page: 1, limit: 500 });

      let serverStats = null;
      try {
        const res = await api.getAdminDashboard();
        serverStats = res.data;
      } catch (err) {
        console.warn("Backend dashboard endpoint belum siap/error (404).");
      }

      const [pendingRes, pendingSupervisorsRes, tasksRes, attendanceRes] =
        await Promise.all([
          pendingReq,
          pendingSupervisorsReq,
          submittedTasksReq,
          attendanceReq,
        ]);
      const attendanceRecords = attendanceRes?.data || [];

      const todayStr = new Date().toLocaleDateString("en-CA"); // YYYY-MM-DD
      const presentTodayCount = attendanceRecords.filter((r) => {
        const rDate = r.date ? r.date.slice(0, 10) : "";
        return rDate === todayStr && ["present", "late"].includes(r.status);
      }).length;

      if (serverStats) {
        const stats = serverStats.stats || {};
        dashboardData = {
          totalInterns: stats.total_interns || 0,
          completedOnTime: stats.completed_on_time || 0,
          completedLate: stats.completed_late || 0,
          presentToday: stats.present_today || presentTodayCount,
          overdueTasks: stats.overdue_tasks || 0,
          inProgressTasks: stats.in_progress_tasks || 0,
          attendanceTrend:
            serverStats.weekly_trend && serverStats.weekly_trend.length
              ? serverStats.weekly_trend
              : buildAttendanceTrend(attendanceRecords),
          pendingRegistrations:
            stats.pending_registrations ||
            pendingRes.meta?.total ||
            (pendingRes.data ? pendingRes.data.length : 0),
          pendingSupervisors:
            stats.pending_supervisors ||
            pendingSupervisorsRes.meta?.total ||
            (pendingSupervisorsRes.data
              ? pendingSupervisorsRes.data.length
              : 0),
          pendingInternsList: pendingRes.data || [],
          pendingSupervisorsList: pendingSupervisorsRes.data || [],
          submittedTasks: tasksRes.data || [],
        };
      } else {
        const allInternsRes = await api.getInterns({
          limit: 1000,
          ...(auth.user?.role === "supervisor" ||
          auth.user?.role === "pembimbing"
            ? { supervisor_id: auth.user.id }
            : {}),
        });
        const allInterns = allInternsRes.data || [];
        const activeCount = allInterns.filter(
          (i) => i.status === "active",
        ).length;

        let taskOnTime = 0,
          taskLate = 0,
          taskOverdue = 0,
          taskInProgress = 0;
        try {
          const allTasksRes = await api.getTasks({ limit: 1000 });
          const allTasks = allTasksRes.data || [];
          taskOnTime = allTasks.filter(
            (t) => t.status === "completed" && !t.is_late,
          ).length;
          taskLate = allTasks.filter(
            (t) => t.status === "completed" && t.is_late,
          ).length;
          taskOverdue = allTasks.filter(
            (t) => t.status !== "completed" && t.is_late,
          ).length;
          taskInProgress = allTasks.filter(
            (t) => t.status !== "completed" && !t.is_late,
          ).length;
        } catch (e) {
          console.log("Gagal hitung task manual");
        }

        dashboardData = {
          totalInterns: activeCount,
          completedOnTime: taskOnTime,
          completedLate: taskLate,
          presentToday: presentTodayCount,
          overdueTasks: taskOverdue,
          inProgressTasks: taskInProgress,
          attendanceTrend: buildAttendanceTrend(attendanceRecords),
          pendingRegistrations:
            pendingRes.meta?.total ||
            (pendingRes.data ? pendingRes.data.length : 0),
          pendingSupervisors:
            pendingSupervisorsRes.meta?.total ||
            (pendingSupervisorsRes.data
              ? pendingSupervisorsRes.data.length
              : 0),
          pendingInternsList: pendingRes.data || [],
          pendingSupervisorsList: pendingSupervisorsRes.data || [],
          submittedTasks: tasksRes.data || [],
        };
      }
    } catch (err) {
      console.error("Fatal error fetching dashboard:", err);
    } finally {
      loading = false;
    }
  }

  async function handleApprove(id, name) {
    if (!confirm(`Terima siswa "${name}" menjadi Active?`)) return;
    try {
      await api.updateInternStatus(id, "active");
      dashboardData.pendingInternsList =
        dashboardData.pendingInternsList.filter((i) => i.id !== id);
      dashboardData.pendingRegistrations--;
      dashboardData.totalInterns++;
      alert(`Berhasil menerima ${name}`);
    } catch (err) {
      alert("Gagal approve: " + (err.response?.data?.message || err.message));
    }
  }

  async function handleDeny(id, name) {
    if (
      !confirm(
        `Apakah Anda yakin ingin MENOLAK dan MENGHAPUS pendaftaran "${name}"? Data akan hilang permanen.`,
      )
    )
      return;
    try {
      await api.deleteIntern(id);
      dashboardData.pendingInternsList =
        dashboardData.pendingInternsList.filter((i) => i.id !== id);
      dashboardData.pendingRegistrations--;
      alert(`Pendaftaran ${name} telah ditolak dan dihapus.`);
    } catch (err) {
      alert("Gagal menolak: " + (err.response?.data?.message || err.message));
    }
  }

  async function handleApproveSupervisor(id, name) {
    if (!confirm(`Terima pembimbing "${name}" menjadi Active?`)) return;
    try {
      await api.approveSupervisor(id);
      dashboardData.pendingSupervisorsList =
        dashboardData.pendingSupervisorsList.filter((s) => s.id !== id);
      dashboardData.pendingSupervisors--;
      alert(`Berhasil menerima ${name}`);
    } catch (err) {
      alert("Gagal approve: " + (err.response?.data?.message || err.message));
    }
  }

  async function handleDenySupervisor(id, name) {
    if (
      !confirm(
        `Apakah Anda yakin ingin MENOLAK dan MENGHAPUS pendaftaran pembimbing "${name}"? Data akan hilang permanen.`,
      )
    )
      return;
    try {
      await api.rejectSupervisor(id);
      dashboardData.pendingSupervisorsList =
        dashboardData.pendingSupervisorsList.filter((s) => s.id !== id);
      dashboardData.pendingSupervisors--;
      alert(`Pendaftaran ${name} telah ditolak dan dihapus.`);
    } catch (err) {
      alert("Gagal menolak: " + (err.response?.data?.message || err.message));
    }
  }

  function initCharts() {
    const taskPieCtx = document.getElementById("taskPieChart");
    if (taskPieCtx && taskPieCtx instanceof HTMLCanvasElement) {
      if (taskPieChart) taskPieChart.destroy();
      taskPieChart = new Chart(taskPieCtx.getContext("2d"), {
        type: "doughnut",
        data: {
          labels: [
            "Selesai Tepat Waktu",
            "Selesai Terlambat",
            "Overdue",
            "Dalam Proses",
          ],
          datasets: [
            {
              data: [
                dashboardData.completedOnTime,
                dashboardData.completedLate,
                dashboardData.overdueTasks,
                dashboardData.inProgressTasks,
              ],
              backgroundColor: ["#10b981", "#3b82f6", "#ef4444", "#94a3b8"],
              hoverBackgroundColor: [
                "#059669",
                "#2563eb",
                "#dc2626",
                "#64748b",
              ],
              borderWidth: 2,
              borderColor: "#ffffff",
              hoverOffset: 5,
            },
          ],
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          cutout: "70%",
          layout: { padding: 10 },
          plugins: {
            legend: {
              position: "bottom",
              labels: {
                usePointStyle: true,
                pointStyle: "circle",
                padding: 20,
                font: { family: "Geist, Inter, sans-serif", size: 12 },
                color: "#64748b",
              },
            },
          },
        },
      });
    }

    const trendCtx = document.getElementById("attendanceTrendChart");
    if (
      trendCtx &&
      trendCtx instanceof HTMLCanvasElement &&
      dashboardData.attendanceTrend.length > 0
    ) {
      if (attendanceTrendChart) attendanceTrendChart.destroy();

      const labels = dashboardData.attendanceTrend.map((d) => d.day);
      const presentData = dashboardData.attendanceTrend.map((d) => d.present);
      const absentData = dashboardData.attendanceTrend.map((d) => d.absent);

      attendanceTrendChart = new Chart(trendCtx.getContext("2d"), {
        type: "bar",
        data: {
          labels: labels,
          datasets: [
            {
              label: "Hadir",
              data: presentData,
              backgroundColor: "#10b981",
              hoverBackgroundColor: "#059669",
              borderRadius: 6,
              barThickness: 16,
            },
            {
              label: "Tidak Hadir",
              data: absentData,
              backgroundColor: "#fca5a5",
              hoverBackgroundColor: "#ef4444",
              borderRadius: 6,
              barThickness: 16,
            },
          ],
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: { display: false },
            tooltip: {
              backgroundColor: "rgba(15, 23, 42, 0.9)",
              padding: 10,
              cornerRadius: 8,
              titleFont: { family: "Geist, Inter, sans-serif" },
              bodyFont: { family: "Geist, Inter, sans-serif" },
            },
          },
          scales: {
            x: {
              stacked: true,
              grid: { display: false },
              ticks: {
                font: { family: "Geist, Inter, sans-serif" },
                color: "#64748b",
              },
            },
            y: {
              stacked: true,
              beginAtZero: true,
              grid: { color: "#f1f5f9" },
              ticks: {
                font: { family: "Geist, Inter, sans-serif" },
                color: "#64748b",
                stepSize: 1,
              },
            },
          },
        },
      });
    }
  }

  onMount(fetchDashboardData);
</script>

<svelte:head>
  <link
    rel="stylesheet"
    href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20,400,0,0"
  />
</svelte:head>

<div class="max-w-[1200px] mx-auto space-y-5 px-3 py-0 sm:px-6 animate-fade-in">
  <!-- Header -->
  <div
    class="flex flex-col md:flex-row md:justify-between md:items-center gap-4"
  >
    <div>
      <h2 class="text-header font-bold text-slate-800 tracking-tight">
        Halo, <span class="text-emerald-600">{auth.user?.name || "Admin"}</span>
      </h2>
      <p class="text-sm text-slate-600 mt-1 pb-4">
        Pantau progres dan aktivitas magang secara real-time.
      </p>
    </div>

    <div class="date-pill">
      <svg
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
        <line x1="16" y1="2" x2="16" y2="6"></line>
        <line x1="8" y1="2" x2="8" y2="6"></line>
        <line x1="3" y1="10" x2="21" y2="10"></line>
      </svg>
      <span class="hidden sm:inline">{currentDate}</span>
      <span class="sm:hidden"
        >{new Date().toLocaleDateString("id-ID", {
          day: "2-digit",
          month: "short",
          year: "numeric",
        })}</span
      >
    </div>
  </div>

  {#if loading}
    <div class="flex flex-col items-center justify-center">
      <div class="spinner"></div>
      <p class="text-slate-600 mt-4">Menyiapkan data...</p>
    </div>
  {:else}
    <!-- Pending Registrations Card -->
    {#if auth.user?.role === "admin" && (dashboardData.pendingInternsList.length > 0 || dashboardData.pendingSupervisorsList.length > 0)}
      <div class="card approval-card animate-slide-up">
        <div class="approval-decoration"></div>
        <div
          class="pt-4 border-b border-slate-100 flex flex-col sm:flex-row sm:justify-between sm:items-center gap-3 bg-emerald-50/30"
        >
          <div class="flex items-center gap-3">
            <div class="icon-pulse shrink-0">
              <svg
                width="22"
                height="22"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                <circle cx="8.5" cy="7" r="4" />
                <polyline points="17 11 19 13 23 9" />
              </svg>
            </div>
            <div>
              <h3 class="font-bold text-base sm:text-lg text-slate-800">
                Pendaftaran Baru
              </h3>
              <p class="text-xs sm:text-sm text-slate-600 mt-1">
                {#if dashboardData.pendingRegistrations > 0 && dashboardData.pendingSupervisors > 0}
                  <span class="count-badge"
                    >{dashboardData.pendingRegistrations}</span
                  >
                  siswa,
                  <span class="count-badge"
                    >{dashboardData.pendingSupervisors}</span
                  > pembimbing menunggu persetujuan.
                {:else if dashboardData.pendingRegistrations > 0}
                  <span class="count-badge"
                    >{dashboardData.pendingRegistrations}</span
                  > siswa menunggu persetujuan.
                {:else}
                  <span class="count-badge"
                    >{dashboardData.pendingSupervisors}</span
                  > pembimbing menunggu persetujuan.
                {/if}
              </p>
            </div>
          </div>
          <a href="/interns?status=pending" class="link-view-all shrink-0">
            Lihat Semua
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <path d="M5 12h14" /><path d="M12 5l7 7-7 7" />
            </svg>
          </a>
        </div>

        <div class="p-3 sm:p-5 space-y-2">
          {#each dashboardData.pendingInternsList as intern}
            <div class="approval-item">
              <div class="flex items-center gap-2 sm:gap-3 flex-1 min-w-0">
                <div class="avatar-initial shrink-0">
                  {intern.full_name?.charAt(0) || "I"}
                </div>
                <div class="min-w-0 flex-1">
                  <h4 class="font-semibold text-slate-800 text-sm truncate">
                    {intern.full_name}
                  </h4>
                  <span class="text-xs text-slate-500 truncate block"
                    >{intern.school || "-"} • {intern.department || "-"}</span
                  >
                </div>
              </div>
              <div class="flex gap-2 shrink-0 approval-buttons">
                <button
                  class="btn-deny"
                  onclick={() => handleDeny(intern.id, intern.full_name)}
                  title="Tolak & Hapus"
                >
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                  <span class="hidden sm:inline">Tolak</span>
                </button>
                <button
                  class="btn-approve"
                  onclick={() => handleApprove(intern.id, intern.full_name)}
                  title="Terima"
                >
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                  <span class="hidden sm:inline">Terima</span>
                </button>
              </div>
            </div>
          {/each}
          {#each dashboardData.pendingSupervisorsList as supervisor}
            <div class="approval-item">
              <div class="flex items-center gap-2 sm:gap-3 flex-1 min-w-0">
                <div class="avatar-initial supervisor-avatar shrink-0">
                  {supervisor.full_name?.charAt(0) || "P"}
                </div>
                <div class="min-w-0 flex-1">
                  <h4 class="font-semibold text-slate-800 text-sm truncate">
                    {supervisor.full_name}
                  </h4>
                  <span class="text-xs text-slate-500 truncate block"
                    >{supervisor.institution || "-"} • Pembimbing</span
                  >
                </div>
              </div>
              <div class="flex gap-2 shrink-0 approval-buttons">
                <button
                  class="btn-deny"
                  onclick={() =>
                    handleDenySupervisor(supervisor.id, supervisor.full_name)}
                  title="Tolak & Hapus"
                >
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                  <span class="hidden sm:inline">Tolak</span>
                </button>
                <button
                  class="btn-approve"
                  onclick={() =>
                    handleApproveSupervisor(
                      supervisor.id,
                      supervisor.full_name,
                    )}
                  title="Terima"
                >
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                  <span class="hidden sm:inline">Terima</span>
                </button>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Stats Grid -->
    <div class="stats-grid animate-slide-up" style="animation-delay: 0.1s;">
      <!-- Total Active -->
      <div class="stat-card">
        <div class="stat-icon-wrapper bg-emerald-soft">
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
            <circle cx="9" cy="7" r="4" />
            <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
            <path d="M16 3.13a4 4 0 0 1 0 7.75" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="label">Siswa Aktif</span>
          <span class="value">{dashboardData.totalInterns}</span>
        </div>
        <div class="stat-decoration bg-emerald"></div>
      </div>

      <!-- Hadir Hari Ini -->
      <div class="stat-card">
        <div class="stat-icon-wrapper bg-teal-soft">
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
            <polyline points="22 4 12 14.01 9 11.01" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="label">Hadir Hari Ini</span>
          <span class="value">{dashboardData.presentToday}</span>
        </div>
        <div class="stat-decoration bg-teal"></div>
      </div>

      <!-- Tepat Waktu -->
      <div class="stat-card">
        <div class="stat-icon-wrapper bg-green-soft">
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="20 6 9 17 4 12"></polyline>
          </svg>
        </div>
        <div class="stat-content">
          <span class="label">Tugas Tepat Waktu</span>
          <span class="value">{dashboardData.completedOnTime}</span>
        </div>
        <div class="stat-decoration bg-green"></div>
      </div>

      <!-- Terlambat -->
      <div class="stat-card">
        <div class="stat-icon-wrapper bg-amber-soft">
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <path d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="label">Tugas Overdue</span>
          <span class="value">{dashboardData.overdueTasks}</span>
        </div>
        <div class="stat-decoration bg-amber"></div>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="charts-grid animate-slide-up" style="animation-delay: 0.2s;">
      <!-- Task Chart -->
      <div class="chart-card">
        <div class="card-header">
          <h3 class="font-bold text-base sm:text-lg text-slate-800">
            Distribusi Tugas
          </h3>
          <span class="badge-soft"
            >Total: {dashboardData.completedOnTime +
              dashboardData.completedLate +
              dashboardData.overdueTasks +
              dashboardData.inProgressTasks}</span
          >
        </div>
        <div class="canvas-wrapper">
          <canvas id="taskPieChart"></canvas>
        </div>
      </div>

      <!-- Attendance Chart -->
      <div class="chart-card">
        <div class="card-header">
          <h3 class="font-bold text-base sm:text-lg text-slate-800">
            Tren Kehadiran
          </h3>
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
            <svg
              width="48"
              height="48"
              viewBox="0 0 24 24"
              fill="none"
              stroke="#cbd5e1"
              stroke-width="2"
            >
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <line x1="9" y1="9" x2="15" y2="15" />
              <line x1="15" y1="9" x2="9" y2="15" />
            </svg>
            <p class="text-sm text-slate-500 mt-3">Belum ada data kehadiran</p>
          </div>
        {/if}
      </div>
    </div>

    <!-- Submitted Tasks List -->
    {#if dashboardData.submittedTasks.length > 0}
      <div
        class="tasks-section animate-slide-up"
        style="animation-delay: 0.3s;"
      >
        <div class="section-header">
          <div class="header-title">
            <h3 class="font-bold text-base sm:text-lg text-slate-800">
              Tugas Baru Dikumpulkan
            </h3>
            <span class="badge-count"
              >{dashboardData.submittedTasks.length}</span
            >
          </div>
          <a href="/tasks?status=submitted" class="link-view-all">
            Lihat Semua
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <path d="M5 12h14" /><path d="M12 5l7 7-7 7" />
            </svg>
          </a>
        </div>
        <div class="task-grid">
          {#each dashboardData.submittedTasks as task}
            <div class="task-item">
              <div class="task-icon">
                <svg
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path
                    d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                  />
                  <polyline points="14 2 14 8 20 8" />
                  <line x1="16" y1="13" x2="8" y2="13" />
                  <line x1="16" y1="17" x2="8" y2="17" />
                  <polyline points="10 9 9 9 8 9" />
                </svg>
              </div>
              <div class="task-info">
                <h4 class="font-semibold text-slate-800 text-sm">
                  {task.title}
                </h4>
                <p class="text-xs text-slate-500">
                  Oleh: <span class="font-medium text-slate-700"
                    >{task.intern_name || "Siswa Magang"}</span
                  >
                </p>
              </div>
              <a href="/tasks/{task.id}" class="btn-review">
                <span class="hidden sm:inline">Review</span>
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path
                    d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"
                  />
                  <polyline points="15 3 21 3 21 9" />
                  <line x1="10" y1="14" x2="21" y2="3" />
                </svg>
              </a>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  {/if}
</div>

<style>
  /* Base styles */
  :global(body) {
    font-family: "Geist", "Inter", sans-serif;
  }

  /* Header */
  .date-pill {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 9999px;
    color: #64748b;
    font-size: 14px;
    font-weight: 500;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.03);
  }

  /* Approval Card */
  .card {
    background: white;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
    overflow: hidden;
  }

  .approval-card {
    position: relative;
    border-color: rgba(16, 185, 129, 0.2);
    box-shadow: 0 10px 30px -10px rgba(16, 185, 129, 0.15);
  }

  .approval-decoration {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 6px;
    background: linear-gradient(90deg, #10b981, #34d399);
  }

  .icon-pulse {
    width: 40px;
    height: 40px;
    background: #ffffff;
    color: #10b981;
    border: 1px solid #d1fae5;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.1);
  }

  @media (min-width: 640px) {
    .icon-pulse {
      width: 48px;
      height: 48px;
    }
  }

  .count-badge {
    background: #fef3c7;
    color: #b45309;
    padding: 2px 8px;
    border-radius: 6px;
    font-weight: 600;
    font-size: 12px;
    border: 1px solid #fde68a;
  }

  .link-view-all {
    color: #059669;
    font-weight: 600;
    font-size: 14px;
    text-decoration: none;
    display: flex;
    align-items: center;
    gap: 6px;
    transition: gap 0.2s;
  }

  .link-view-all:hover {
    gap: 10px;
    color: #047857;
  }

  .approval-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 0px;
    border-radius: 12px;
    border: 1px solid transparent;
    transition: all 0.2s ease;
    gap: 8px;
  }

  @media (min-width: 640px) {
    .approval-item {
      padding: 12px 16px;
      gap: 12px;
    }
  }

  .approval-item:hover {
    background: #f8fafc;
    border-color: #e2e8f0;
  }

  .avatar-initial {
    width: 36px;
    height: 36px;
    background: linear-gradient(135deg, #10b981, #059669);
    color: white;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 14px;
    box-shadow: 0 4px 10px rgba(16, 185, 129, 0.3);
  }

  @media (min-width: 640px) {
    .avatar-initial {
      width: 40px;
      height: 40px;
      font-size: 16px;
    }
  }

  .btn-approve {
    background: #ffffff;
    color: #10b981;
    border: 1px solid #10b981;
    padding: 6px 10px;
    border-radius: 10px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 6px;
    transition: all 0.2s;
  }

  @media (min-width: 640px) {
    .btn-approve {
      padding: 6px 12px;
    }
  }

  .btn-approve:hover {
    background: #10b981;
    color: white;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
    transform: translateY(-1px);
  }

  .btn-deny {
    background: #ffffff;
    color: #ef4444;
    border: 1px solid #ef4444;
    padding: 6px 10px;
    border-radius: 10px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 6px;
    transition: all 0.2s;
  }

  @media (min-width: 640px) {
    .btn-deny {
      padding: 6px 12px;
    }
  }

  .btn-deny:hover {
    background: #ef4444;
    color: white;
    box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
    transform: translateY(-1px);
  }

  /* Mobile responsive approval buttons */
  @media (max-width: 640px) {
    .approval-item {
      flex-direction: column;
      align-items: stretch;
    }

    .approval-buttons {
      width: 100%;
      margin-top: 8px;
    }

    .btn-approve,
    .btn-deny {
      flex: 1;
      justify-content: center;
      padding: 8px 12px;
    }

    .btn-approve span,
    .btn-deny span {
      display: inline !important;
    }
  }

  /* Supervisor avatar styling */
  .supervisor-avatar {
    background: linear-gradient(135deg, #6366f1, #4f46e5) !important;
  }

  /* Stats Grid */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 12px;
  }

  @media (min-width: 640px) {
    .stats-grid {
      gap: 16px;
    }
  }

  .stat-card {
    background: white;
    padding: 10px;
    border-radius: 16px;
    border: 2px solid #e0e0e0;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.01);
    display: flex;
    align-items: center;
    gap: 12px;
    position: relative;
    overflow: hidden;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  @media (min-width: 640px) {
    .stat-card {
      padding: 10px;
      gap: 16px;
    }
  }

  .stat-card:hover {
    /* transform: translateY(-4px); */
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.05);
    /* border-color: transparent; */
  }

  .stat-icon-wrapper {
    width: 48px;
    height: 48px;
    border-radius: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  @media (min-width: 640px) {
    .stat-icon-wrapper {
      width: 56px;
      height: 56px;
    }
  }

  .bg-emerald-soft {
    background: #ecfdf5;
    color: #10b981;
  }
  .bg-teal-soft {
    background: #f0f9ff;
    color: #0ea5e9;
  }
  .bg-green-soft {
    background: #f0fdf4;
    color: #16a34a;
  }
  .bg-amber-soft {
    background: #fffbeb;
    color: #f59e0b;
  }

  .bg-emerald {
    background: #10b981;
  }
  .bg-teal {
    background: #0ea5e9;
  }
  .bg-green {
    background: #16a34a;
  }
  .bg-amber {
    background: #f59e0b;
  }

  .stat-content {
    display: flex;
    flex-direction: column;
    z-index: 1;
  }

  .stat-content .value {
    font-size: 24px;
    font-weight: 800;
    color: #0f172a;
    line-height: 1.1;
    letter-spacing: -0.02em;
  }

  @media (min-width: 640px) {
    .stat-content .value {
      font-size: 28px;
    }
  }

  .stat-content .label {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    color: #94a3b8;
    letter-spacing: 0.05em;
    margin-bottom: 4px;
  }

  @media (min-width: 640px) {
    .stat-content .label {
      font-size: 12px;
    }
  }

  .stat-decoration {
    position: absolute;
    bottom: 0;
    left: 0;
    width: 100%;
    height: 4px;
    opacity: 0;
    transition: opacity 0.3s;
  }

  .stat-card:hover .stat-decoration {
    opacity: 1;
  }

  /* Charts Grid */
  .charts-grid {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 16px;
  }

  @media (max-width: 900px) {
    .charts-grid {
      grid-template-columns: 1fr;
    }
  }

  .chart-card {
    background: white;
    padding: 15px;
    border-radius: 20px;
    border: 1px solid #e0e0e0;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.01);
    transition: transform 0.3s;
  }

  @media (min-width: 640px) {
    .chart-card {
      padding: 15px;
    }
  }

  .chart-card:hover {
    /* transform: translateY(-2px); */
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    flex-wrap: wrap;
    gap: 8px;
  }

  @media (min-width: 640px) {
    .card-header {
      margin-bottom: 10px;
      gap: 12px;
    }
  }

  .badge-soft {
    background: #f1f5f9;
    color: #64748b;
    padding: 4px 12px;
    border-radius: 99px;
    font-size: 12px;
    font-weight: 600;
  }

  .chart-select {
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 4px 8px;
    font-size: 12px;
    color: #64748b;
    background: transparent;
    cursor: pointer;
  }

  .canvas-wrapper {
    height: 280px;
    position: relative;
  }

  @media (min-width: 640px) {
    .canvas-wrapper {
      height: 300px;
    }
  }

  .empty-chart {
    height: 280px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: #f8fafc;
    border-radius: 16px;
    border: 2px dashed #e2e8f0;
  }

  @media (min-width: 640px) {
    .empty-chart {
      height: 300px;
    }
  }

  /* Tasks Section */
  .tasks-section {
    background: white;
    border-radius: 20px;
    border: 1px solid #e0e0e0;
    padding: 16px;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.03);
  }

  @media (min-width: 640px) {
    .tasks-section {
      padding: 15px;
    }
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    flex-wrap: wrap;
    gap: 8px;
  }

  @media (min-width: 640px) {
    .section-header {
      margin-bottom: 20px;
      gap: 12px;
    }
  }

  .text-header {
    font-size: 20px;
    font-weight: 600;
    color: #0f172a;
    line-height: 1.1;
    letter-spacing: -0.02em;
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  @media (min-width: 640px) {
    .header-title {
      gap: 12px;
    }
  }

  .badge-count {
    background: #ef4444;
    color: white;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 600;
  }

  .task-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 12px;
  }

  @media (min-width: 640px) {
    .task-grid {
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
      gap: 16px;
    }
  }

  .task-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px;
    background: #ffffff;
    border: 1px solid #f1f5f9;
    border-radius: 12px;
    transition: all 0.3s ease;
  }

  @media (min-width: 640px) {
    .task-item {
      gap: 12px;
      padding: 16px;
    }
  }

  .task-item:hover {
    border-color: #cbd5e1;
    transform: translateY(-2px);
    box-shadow: 0 10px 20px -5px rgba(0, 0, 0, 0.05);
  }

  .task-icon {
    width: 40px;
    height: 40px;
    background: #f8fafc;
    color: #64748b;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: all 0.3s;
  }

  @media (min-width: 640px) {
    .task-icon {
      width: 44px;
      height: 44px;
    }
  }

  .task-item:hover .task-icon {
    background: #e0f2fe;
    color: #0ea5e9;
  }

  .task-info {
    flex: 1;
    min-width: 0;
  }

  .task-info h4 {
    margin: 0 0 4px 0;
    font-size: 14px;
    font-weight: 600;
    color: #334155;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .task-info p {
    margin: 0;
    font-size: 12px;
    color: #94a3b8;
  }

  .btn-review {
    padding: 6px 10px;
    background: #f1f5f9;
    color: #475569;
    border-radius: 8px;
    text-decoration: none;
    font-size: 12px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 6px;
    transition: all 0.2s;
    flex-shrink: 0;
  }

  @media (min-width: 640px) {
    .btn-review {
      padding: 6px 12px;
    }
  }

  .btn-review:hover {
    background: #0f172a;
    color: white;
  }

  /* Loading Spinner */
  .spinner {
    width: 48px;
    height: 48px;
    border: 4px solid #e2e8f0;
    border-top-color: #10b981;
    border-radius: 50%;
    margin: 0 auto;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  /* Animations */
  .animate-fade-in {
    opacity: 0;
    animation: fadeIn 0.6s ease-out forwards;
  }

  .animate-slide-up {
    opacity: 0;
    animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
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

  /* Responsive adjustments */
  @media (max-width: 900px) {
    .stats-grid {
      grid-template-columns: repeat(2, 1fr);
    }

    .stat-card {
      flex-direction: column;
      text-align: center;
      align-items: center;
    }

    .stat-content {
      align-items: center;
    }

    .task-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
