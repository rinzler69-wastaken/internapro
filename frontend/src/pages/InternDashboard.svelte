<script>
  import { onMount, onDestroy } from "svelte";
  import { api } from "../lib/api.js";
  import { auth } from "../lib/auth.svelte.js";
  import Chart from "chart.js/auto";
  import L from "leaflet";

  // Office Configuration - Loaded from settings
  let OFFICE = $state({
    lat: -7.0355,
    lng: 110.4746,
    name: "Kantor Pusat",
    maxDistance: 1000, // meters
  });

  // ==================== STATE ====================
  let loading = $state(true);
  let dashboardLoading = $state(false);
  let viewMode = $state("presensi"); // 'presensi' | 'tugas'

  // Attendance State
  let todayAttendance = $state(null);
  let gpsStatus = $state("loading"); // loading, ready, error, denied
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
    taskStats: {
      pending: 0,
      in_progress: 0,
      submitted: 0,
      completed: 0,
      revision: 0,
    },
    taskBreakdown: {
      pending: 0,
      in_progress: 0,
      submitted: 0,
      completed: 0,
      revision: 0,
    },
    recentTasks: [],
    weeklyAttendance: { days: [], counts: [], colors: [] },
    attendancePercentage: 0,
    attendanceHistory: [], // weekly list
    attendanceLast30: [], // last 30 days raw
    office: null,
  });

  // Modals
  let showLateModal = $state(false);
  let showPermissionModal = $state(false);
  let lateReason = $state("");
  let permissionForm = $state({
    type: "permission",
    reason: "",
    document: null,
  });

  // Charts
  /** @type {any} */
  let taskProgressChart = null;
  /** @type {any} */
  let weeklyAttendanceChart = null;
  let watchId = null;

  // Permission Form State
  let permissionStatus = $state("sick");
  let permissionNotes = $state("");
  let permissionFile = $state(null);
  let fileInput = $state(null);

  let isSubmittingPermission = $state(false); // Khusus Kirim Izin

  // ==================== DERIVED STATE ====================

  // Calculate distance from office
  const distance = $derived(
    gps.lat && gps.lng
      ? Math.round(getDistance(gps.lat, gps.lng, OFFICE.lat, OFFICE.lng))
      : null,
  );

  // Check if user can check in
  const canCheckIn = $derived(
    !loading &&
      !checkingIn &&
      gpsStatus === "ready" &&
      distance !== null &&
      distance <= OFFICE.maxDistance &&
      todayAttendance === null,
  );

  // Check if user can check out
  const canCheckOut = $derived(
    !loading &&
      !checkingOut &&
      todayAttendance !== null &&
      (todayAttendance.checked_in === true ||
        !!todayAttendance.check_in_time) &&
      (todayAttendance.checked_out === false ||
        !todayAttendance.check_out_time) &&
      !["permission", "sick"].includes(todayAttendance.status),
  );

  // Check if user is late
  const isLate = $derived(() => {
    const now = new Date();
    const hour = now.getHours();
    const minute = now.getMinutes();
    return hour > 8 || (hour === 8 && minute > 0);
  });

  // Attendance completion status
  const attendanceComplete = $derived(
    todayAttendance !== null &&
      (todayAttendance.checked_out === true ||
        !!todayAttendance.check_out_time ||
        ["permission", "sick", "absent", "off"].includes(
          todayAttendance.status,
        )),
  );

  // ==================== GEOLOCATION ====================

  // Haversine formula for distance calculation
  function getDistance(lat1, lon1, lat2, lon2) {
    const R = 6371e3; // metres
    const φ1 = (lat1 * Math.PI) / 180;
    const φ2 = (lat2 * Math.PI) / 180;
    const Δφ = ((lat2 - lat1) * Math.PI) / 180;
    const Δλ = ((lon2 - lon1) * Math.PI) / 180;
    const a =
      Math.sin(Δφ / 2) * Math.sin(Δφ / 2) +
      Math.cos(φ1) * Math.cos(φ2) * Math.sin(Δλ / 2) * Math.sin(Δλ / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c;
  }

  // Initialize geolocation tracking
  function initGeolocation() {
    if (!navigator.geolocation) {
      gpsStatus = "error";
      gps.error = "Geolocation not supported";
      return;
    }

    // Get initial position
    navigator.geolocation.getCurrentPosition(
      (position) => {
        gps = {
          lat: position.coords.latitude,
          lng: position.coords.longitude,
          error: null,
        };
        gpsStatus = "ready";
        updateMapPosition();
      },
      (error) => {
        console.error("Geolocation error:", error);
        gpsStatus = error.code === 1 ? "denied" : "error";
        gps.error = error.message;
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 },
    );

    // Watch position for continuous updates
    watchId = navigator.geolocation.watchPosition(
      (position) => {
        gps = {
          lat: position.coords.latitude,
          lng: position.coords.longitude,
          error: null,
        };
        if (gpsStatus !== "ready") gpsStatus = "ready";
        updateMapPosition();
      },
      (error) => {
        console.error("Watch position error:", error);
      },
      { enableHighAccuracy: true, maximumAge: 30000 },
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
          className: "user-marker",
          html: '<div style="background:#10b981; width:24px; height:24px; border-radius:50%; border:3px solid white; box-shadow:0 2px 8px rgba(0,0,0,0.3); animation: pulse 2s infinite;"></div>',
          iconSize: [24, 24],
        }),
      }).addTo(map);
    }
  }

  // ==================== MAP INITIALIZATION ====================

  function initMap() {
    if (typeof window === "undefined") return;

    const mapElement = document.getElementById("map");
    if (!mapElement) return;

    // If map exists but DOM was destroyed (view switched), recreate
    if (map && map._container !== mapElement) {
      map.remove();
      map = null;
    }

    if (!map) {
      map = L.map("map").setView([OFFICE.lat, OFFICE.lng], 15);

      L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution: "&copy; OpenStreetMap contributors",
      }).addTo(map);

      officeMarker = L.marker([OFFICE.lat, OFFICE.lng], {
        icon: L.divIcon({
          className: "office-marker",
          html: '<div style="background:#10b981; width:32px; height:32px; border-radius:50%; border:4px solid white; box-shadow:0 2px 8px rgba(0,0,0,0.3);"></div>',
          iconSize: [32, 32],
        }),
      })
        .addTo(map)
        .bindPopup(OFFICE.name);

      radiusCircle = L.circle([OFFICE.lat, OFFICE.lng], {
        color: "#6366f1",
        fillColor: "#6366f1",
        fillOpacity: 0.1,
        radius: OFFICE.maxDistance,
      }).addTo(map);
    }

    setTimeout(() => map.invalidateSize(), 100);
  }

  // ==================== DATA FETCHING ====================

  async function fetchSettings() {
    try {
      // Use the new office-info endpoint that's accessible to all authenticated users
      const res = await fetch("/api/office-info", {
        headers: {
          Authorization: `Bearer ${auth.token}`,
        },
      });

      if (!res.ok) {
        throw new Error(`HTTP ${res.status}: ${res.statusText}`);
      }

      const json = await res.json();
      const map = json?.data || {};

      // Update OFFICE configuration from settings
      OFFICE = {
        lat: parseFloat(map.office_latitude || OFFICE.lat),
        lng: parseFloat(map.office_longitude || OFFICE.lng),
        name: map.office_name || OFFICE.name,
        maxDistance: parseInt(
          map.max_checkin_distance || map.office_radius || OFFICE.maxDistance,
          10,
        ),
      };

      console.log("Office settings loaded:", OFFICE);
    } catch (err) {
      console.warn("Failed to load office settings, using defaults:", err);
    }
  }

  function formatLocalDateKey(d) {
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${year}-${month}-${day}`;
  }

  function getCurrentWeekRange() {
    const today = new Date();
    // Make Monday = 0, Sunday = 6
    const dayIndex = (today.getDay() + 6) % 7;
    const start = new Date(today);
    start.setDate(today.getDate() - dayIndex);
    start.setHours(0, 0, 0, 0);

    const end = new Date(start);
    end.setDate(start.getDate() + 6);
    end.setHours(23, 59, 59, 999);

    return {
      start,
      end,
      startStr: formatLocalDateKey(start),
      endStr: formatLocalDateKey(end),
    };
  }

  function getLast30Range() {
    const today = new Date();
    const start = new Date(today);
    start.setDate(today.getDate() - 29);
    start.setHours(0, 0, 0, 0);
    const end = new Date(today);
    end.setHours(23, 59, 59, 999);
    return {
      startStr: formatLocalDateKey(start),
      endStr: formatLocalDateKey(end),
    };
  }

  function isInCurrentWeek(record) {
    if (!record) return false;
    const { startStr, endStr } = getCurrentWeekRange();
    const key = formatLocalDateKey(
      new Date(
        record.date || record.check_in_time || record.created_at || Date.now(),
      ),
    );
    return key >= startStr && key <= endStr;
  }

  const weeklyAttendanceList = $derived(
    (dashboardData.attendanceLast30?.length
      ? dashboardData.attendanceLast30
      : dashboardData.attendanceHistory || []
    ).filter(isInCurrentWeek),
  );

  async function fetchDashboardData() {
    dashboardLoading = true;
    try {
      const { startStr, endStr } = getCurrentWeekRange();
      const { startStr: start30, endStr: end30 } = getLast30Range();

      const [tasksRes, todayRes, weeklyRes, last30Res] = await Promise.all([
        api.getTasks({ limit: 100 }),
        api.getTodayAttendance(),
        api.getAttendance({ start: startStr, end: endStr, limit: 200 }),
        api.getAttendance({ start: start30, end: end30, limit: 400 }),
      ]);

      const tasks = tasksRes.data || [];
      const weeklyHistory = weeklyRes.data || [];
      const history30 = last30Res.data || [];

      // Parse today's attendance (align with Attendance.svelte structure)
      const rawToday = todayRes.data || {};
      const todayData = rawToday.attendance || (rawToday.id ? rawToday : null);

      if (rawToday.off_day) {
        todayAttendance = {
          status: rawToday.status || "off",
          off_day: true,
          date: rawToday.date,
          message: rawToday.message || "Tidak ada jadwal kantor!",
          note: rawToday.note || "Selamat beristirahat!",
        };
      } else if (!rawToday.checked_in && rawToday.closed_today) {
        todayAttendance = {
          status: "absent",
          date: rawToday.date,
          check_in_time: null,
          check_out_time: null,
          closed_today: true,
        };
      } else if (todayData) {
        // Actual attendance record exists
        todayAttendance = todayData;
      } else {
        // No attendance yet today; keep null so check-in stays enabled
        todayAttendance = null;
      }

      // Calculate task stats
      const taskStats = {
        pending: 0,
        in_progress: 0,
        submitted: 0,
        completed: 0,
        revision: 0,
      };
      tasks.forEach((t) => {
        if (taskStats[t.status] !== undefined) taskStats[t.status]++;
      });

      // Weekly stats for current week
      const weeklyStats = processWeeklyAttendance(weeklyHistory);

      // Attendance percentage from last 30 days
      const presentCount = history30.filter((h) =>
        ["present", "late"].includes(h.status),
      ).length;
      const totalDays = history30.length || 1;
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
        attendanceHistory: weeklyHistory,
        attendanceLast30: history30,
        office: null,
      };

      // Initialize charts after data is loaded
      setTimeout(initCharts, 100);
    } catch (err) {
      console.error("Failed to fetch dashboard data:", err);
    } finally {
      dashboardLoading = false;
    }
  }

  function processWeeklyAttendance(history) {
    const today = new Date();
    const start = new Date(today);
    start.setDate(today.getDate() - today.getDay() + 1); // Monday
    start.setHours(0, 0, 0, 0);

    const days = [];
    const counts = [];
    const colors = [];

    console.log("Processing weekly attendance, total records:", history.length);

    for (let i = 0; i < 7; i++) {
      const d = new Date(start);
      d.setDate(start.getDate() + i);
      const dateStr = d.toISOString().split("T")[0];
      const dayName = d.toLocaleDateString("id-ID", { weekday: "short" });
      days.push(dayName);

      // Find matching record - check multiple possible date fields
      const record = history.find((h) => {
        // Try different date field formats
        let hDate = "";

        if (h.date) {
          hDate = h.date.split("T")[0];
        } else if (h.check_in_time) {
          hDate = h.check_in_time.split("T")[0];
        }

        return hDate === dateStr;
      });

      console.log(
        `Day ${dayName} (${dateStr}):`,
        record ? `Found - ${record.status}` : "No record",
      );

      if (record && record.status) {
        counts.push(1);
        // Color based on status
        switch (record.status) {
          case "present":
            colors.push("#10b981"); // green
            break;
          case "late":
            colors.push("#f59e0b"); // amber
            break;
          case "sick":
            colors.push("#6366f1"); // indigo
            break;
          case "permission":
            colors.push("#8b5cf6"); // violet
            break;
          case "absent":
            colors.push("#ef4444"); // red
            break;
          default:
            colors.push("#ef4444"); // red
        }
      } else {
        counts.push(0);
        colors.push("#e2e8f0"); // gray for no attendance
      }
    }

    console.log("Weekly data processed:", { days, counts, colors });
    return { days, counts, colors };
  }

  // ==================== CHART INITIALIZATION ====================

  function initCharts() {
    // Destroy existing charts
    if (taskProgressChart) taskProgressChart.destroy();
    if (weeklyAttendanceChart) weeklyAttendanceChart.destroy();

    // Task Progress Pie Chart
    const taskCanvas = document.getElementById("taskProgressChart");
    if (
      taskCanvas instanceof HTMLCanvasElement &&
      dashboardData.taskBreakdown
    ) {
      const ctx = taskCanvas.getContext("2d");
      const breakdown = dashboardData.taskBreakdown;

      taskProgressChart = new Chart(ctx, {
        type: "doughnut",
        data: {
          labels: [
            "Pending",
            "In Progress",
            "Submitted",
            "Completed",
            "Revision",
          ],
          datasets: [
            {
              data: [
                breakdown.pending || 0,
                breakdown.in_progress || 0,
                breakdown.submitted || 0,
                breakdown.completed || 0,
                breakdown.revision || 0,
              ],
              backgroundColor: [
                "#f59e0b", // amber
                "#6366f1", // indigo
                "#8b5cf6", // violet
                "#10b981", // emerald
                "#ef4444", // red
              ],
              borderWidth: 4,
            },
          ],
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          layout: {
            padding: 50,
          },
          plugins: {
            legend: {
              position: "bottom",
              labels: {
                padding: 12,
                font: { size: 14, weight: 600 },
                usePointStyle: true,
                pointStyle: "circle",
              },
            },
          },
        },
      });
    }
  }

  // Re-init charts or map when the visible mode changes
  $effect(() => {
    viewMode;
    if (!dashboardLoading && !loading) {
      if (viewMode === "tugas") {
        setTimeout(initCharts, 50);
      } else if (viewMode === "presensi") {
        setTimeout(() => {
          initMap();
          updateMapPosition();
        }, 50);
      }
    }
  });

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
        longitude: gps.lng,
      };

      if (lateReason && lateReason.trim()) {
        payload.reason = lateReason.trim();
      }

      const res = await api.checkIn(gps.lat, gps.lng, lateReason || null);

      if (res.success !== false) {
        // Refresh dashboard data to get updated attendance
        await fetchDashboardData();
        showLateModal = false;
        lateReason = "";
      }
    } catch (err) {
      console.error("Check-in failed:", err);
      alert(err.message || "Check-in failed");
    } finally {
      checkingIn = false;
    }
  }

  async function handleCheckOut() {
    if (!canCheckOut) return;

    if (!confirm("Konfirmasi check-out?")) return;

    checkingOut = true;
    try {
      await api.checkOut(gps.lat || 0, gps.lng || 0);
      await fetchDashboardData();
    } catch (err) {
      console.error("Check-out failed:", err);
      alert(err.message || "Check-out failed");
    } finally {
      checkingOut = false;
    }
  }

  async function handlePermissionSubmit() {
    if (!permissionNotes?.trim()) {
      alert("Mohon isi keterangan.");
      return;
    }

    isSubmittingPermission = true;
    try {
      await api.submitPermission({
        status: permissionStatus,
        notes: permissionNotes,
        proof_file: permissionFile,
      });

      showPermissionModal = false;
      permissionNotes = "";
      permissionFile = null;
      await fetchDashboardData();
      alert("Pengajuan izin berhasil dikirim.");
    } catch (err) {
      console.error("Permission submission failed:", err);
      alert(err.message || "Gagal mengirim pengajuan");
    } finally {
      isSubmittingPermission = false;
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

    // Fetch settings first to get office configuration
    await fetchSettings();

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
      pending: "bg-amber-50 text-amber-700 border-amber-200",
      in_progress: "bg-indigo-50 text-indigo-700 border-indigo-200",
      submitted: "bg-violet-50 text-violet-700 border-violet-200",
      completed: "bg-emerald-50 text-emerald-700 border-emerald-200",
      revision: "bg-red-50 text-red-700 border-red-200",
    };
    return classes[status] || "bg-slate-50 text-slate-700 border-slate-200";
  }

  function getStatusColor(status) {
    switch (status) {
      case "present":
        return "bg-emerald-100 text-emerald-700 border-emerald-200";
      case "late":
        return "bg-amber-100 text-amber-700 border-amber-200";
      case "sick":
        return "bg-blue-100 text-blue-700 border-blue-200";
      case "permission":
        return "bg-purple-100 text-purple-700 border-purple-200";
      case "absent":
        return "bg-red-100 text-red-700 border-red-200";
      default:
        return "bg-slate-100 text-slate-600 border-slate-200";
    }
  }

  function getStatusLabel(status) {
    const labels = {
      pending: "Pending",
      in_progress: "In Progress",
      submitted: "Submitted",
      completed: "Completed",
      revision: "Revision",
      present: "Hadir",
      late: "Terlambat",
      sick: "Sakit",
      permission: "Izin",
      absent: "Tidak Hadir",
    };
    return labels[status] || status;
  }

  function formatDate(dateStr) {
    if (!dateStr) return "-";
    const date = new Date(dateStr);
    return date.toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    });
  }

  function formatTime(dateStr) {
    if (!dateStr) return "-";
    const date = new Date(dateStr);
    return date.toLocaleTimeString("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  let showModal = $state(false);
  let modalData = $state(null);

  function closeModal() {
    showModal = false;
    modalData = null;
  }

  function handleKeydown(e) {
    if (e.key === "Escape") {
      if (showPermissionModal) showPermissionModal = false;
      if (showLateModal) showLateModal = false;
      if (showModal) closeModal();
    }
  }
</script>

<svelte:head>
  <link
    rel="stylesheet"
    href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"
    integrity="sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY="
    crossorigin=""
  />
  <link
    rel="stylesheet"
    href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20,400,0,0"
  />
</svelte:head>

<svelte:window onkeydown={handleKeydown} />

<div class="slide-up max-w-[1200px] mx-auto space-y-6">
  <!-- Header -->
  <div class="flex flex-col gap-4">
    <!-- Desktop Header -->
    <div class="hidden md:flex justify-between items-center">
      <div>
        <h2 class="text-xl font-semibold text-slate-800 tracking-tight">
          Selamat Datang, <span class="highlight"
            >{auth.user?.name || "Intern"}</span
          >
        </h2>
      </div>
      <div class="flex items-center gap-3">
        <button
          onclick={() => (viewMode = "presensi")}
          class="px-5 py-2 rounded-full text-sm font-semibold transition-all flex items-center justify-center gap-2 {viewMode ===
          'presensi'
            ? 'bg-slate-900 text-white border border-transparent'
            : 'bg-white text-slate-900 border border-slate-200 hover:border-slate-300'}"
        >
          <span class="material-symbols-outlined text-[18px]">map</span>
          Presensi
        </button>
        <button
          onclick={() => (viewMode = "tugas")}
          class="px-5 py-2 rounded-full text-sm font-semibold transition-all flex items-center justify-center gap-2 {viewMode ===
          'tugas'
            ? 'bg-slate-900 text-white border border-transparent'
            : 'bg-white text-slate-900 border border-slate-200 hover:border-slate-300'}"
        >
          <span class="material-symbols-outlined text-[18px]">assignment</span>
          Tugas
        </button>
        <div class="date-pill h-[38px]">
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line
              x1="16"
              y1="2"
              x2="16"
              y2="6"
            ></line><line x1="8" y1="2" x2="8" y2="6"></line><line
              x1="3"
              y1="10"
              x2="21"
              y2="10"
            ></line></svg
          >
          {new Date().toLocaleDateString("id-ID", {
            weekday: "long",
            day: "2-digit",
            month: "long",
            year: "numeric",
          })}
        </div>
      </div>
    </div>

    <!-- Mobile Header -->
    <div class="md:hidden flex flex-col gap-4">
      <h2 class="text-xl font-semibold text-slate-800 tracking-tight">
        Selamat Datang, <span class="highlight"
          >{auth.user?.name || "Intern"}</span
        >
      </h2>
      <div class="flex justify-center">
        <div class="date-pill w-full justify-center h-[38px]">
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line
              x1="16"
              y1="2"
              x2="16"
              y2="6"
            ></line><line x1="8" y1="2" x2="8" y2="6"></line><line
              x1="3"
              y1="10"
              x2="21"
              y2="10"
            ></line></svg
          >
          {new Date().toLocaleDateString("id-ID", {
            weekday: "long",
            day: "2-digit",
            month: "long",
            year: "numeric",
          })}
        </div>
      </div>
      <div class="flex gap-2 w-full">
        <button
          onclick={() => (viewMode = "presensi")}
          class="flex-1 px-5 py-2 rounded-full text-sm font-semibold transition-all flex items-center justify-center gap-2 {viewMode ===
          'presensi'
            ? 'bg-slate-900 text-white border border-transparent'
            : 'bg-white text-slate-900 border border-slate-200 hover:border-slate-300'}"
        >
          <span class="material-symbols-outlined text-[18px]">map</span>
          Presensi
        </button>
        <button
          onclick={() => (viewMode = "tugas")}
          class="flex-1 px-5 py-2 rounded-full text-sm font-semibold transition-all flex items-center justify-center gap-2 {viewMode ===
          'tugas'
            ? 'bg-slate-900 text-white border border-transparent'
            : 'bg-white text-slate-900 border border-slate-200 hover:border-slate-300'}"
        >
          <span class="material-symbols-outlined text-[18px]">assignment</span>
          Tugas
        </button>
      </div>
    </div>
  </div>

  {#if loading}
    <div class="text-center py-12">
      <div
        class="inline-block w-12 h-12 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin"
      ></div>
      <p class="mt-4 text-slate-600">Loading dashboard...</p>
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 animate-fade-in">
      {#if viewMode === "presensi"}
        <!-- PRESENSI MODE: Map (Left 2/3) & Attendance Stats (Right 1/3) -->
        <div class="lg:col-span-2 space-y-6">
          <!-- ATTENDANCE CARD -->
          <div class="card p-0 overflow-hidden">
            <div
              class="px-3 py-2 sm:px-6 sm:py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50"
            >
              <h3
                class="font-bold text-lg text-slate-800 flex items-center gap-2"
              >
                <span class="material-symbols-outlined text-indigo-500"
                  >map</span
                > Presensi Harian
              </h3>
              <div
                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold"
                class:bg-emerald-100={gpsStatus === "ready"}
                class:text-emerald-700={gpsStatus === "ready"}
                class:bg-slate-100={gpsStatus === "loading"}
                class:text-slate-500={gpsStatus === "loading"}
                class:bg-red-100={gpsStatus === "error" ||
                  gpsStatus === "denied"}
                class:text-red-700={gpsStatus === "error" ||
                  gpsStatus === "denied"}
              >
                <div
                  class="w-1.5 h-1.5 rounded-full mr-1.5"
                  class:bg-emerald-500={gpsStatus === "ready"}
                  class:bg-slate-400={gpsStatus === "loading"}
                  class:bg-red-500={gpsStatus === "error" ||
                    gpsStatus === "denied"}
                  class:animate-pulse={gpsStatus === "loading"}
                ></div>
                {#if gpsStatus === "ready"}GPS Ready
                {:else if gpsStatus === "loading"}Mencari Lokasi...
                {:else if gpsStatus === "denied"}GPS Denied
                {:else}GPS Error{/if}
              </div>
            </div>

            <div class="p-4 sm:p-5 space-y-4">
              <!-- Map -->
              <div
                class="relative h-[250px] sm:h-[300px] w-full rounded-xl overflow-hidden border-slate-200 shadow-inner"
              >
                <div id="map" class="h-full w-full z-0"></div>
                <button
                  class="map-reset"
                  style="bottom: 4rem;"
                  onclick={() => {
                    if (gps.lat && gps.lng)
                      map?.setView([gps.lat, gps.lng], 15);
                  }}
                  aria-label="My Location"
                  title="My Location"
                >
                  <span class="material-symbols-outlined text-slate-600"
                    >location_on</span
                  >
                  <span class="text-slate-600 ml-2 text-sm">My Location</span>
                </button>
                <button
                  class="map-reset"
                  onclick={() => map?.setView([OFFICE.lat, OFFICE.lng], 15)}
                  aria-label="Target"
                  title="Target"
                >
                  <span class="material-symbols-outlined text-slate-600"
                    >gps_fixed</span
                  >
                  <span class="text-slate-600 ml-2 text-sm">Re-center</span>
                </button>
              </div>

              <!-- Office Info & Distance -->
              <div
                class="flex flex-col sm:flex-row justify-between items-center bg-slate-50 p-2 rounded-xl border border-slate-100"
              >
                <span
                  class="text-sm font-medium text-slate-600 flex flex-col sm:flex-row items-center sm:items-center text-center sm:text-left gap-1 sm:gap-2"
                >
                  <span
                    class="flex items-center gap-2 justify-center whitespace-nowrap"
                  >
                    <i class="material-symbols-outlined text-slate-400"
                      >domain</i
                    >
                    Kantor:
                  </span>
                  <span class="font-bold text-slate-800">
                    {OFFICE.name}
                  </span>
                </span>
              </div>

              <div class="info-bar">
                <div class="info-item">
                  <span class="label">Jarak Kantor</span>
                  <span
                    class="value {distance !== null &&
                    distance <= OFFICE.maxDistance
                      ? 'text-emerald-600'
                      : 'text-rose-500'}"
                  >
                    {distance !== null ? distance + " m" : "--"}
                  </span>
                </div>
                <div class="info-item border-l border-slate-200 pl-6">
                  <span class="label">Batas Maksimal</span>
                  <span class="value text-slate-700"
                    >{OFFICE.maxDistance} m</span
                  >
                </div>
              </div>

              <!-- Action Buttons -->
              <div>
                {#if todayAttendance?.off_day}
                  <div
                    class="flex items-center gap-4 p-5 bg-emerald-50/70 border border-emerald-200 rounded-xl justify-center"
                  >
                    <div
                      class="w-12 h-12 bg-emerald-100 text-emerald-600 rounded-full flex items-center justify-center shadow-sm shrink-0"
                    >
                      <span class="material-symbols-outlined text-2xl"
                        >check_circle</span
                      >
                    </div>
                    <div>
                      <h4
                        class="font-bold text-emerald-800 text-lg leading-tight"
                      >
                        {todayAttendance.message || "Tidak ada jadwal kantor!"}
                      </h4>
                      <p class="text-emerald-600 font-medium text-sm mt-0.5">
                        {todayAttendance.note || "Selamat beristirahat!"}
                      </p>
                    </div>
                  </div>
                {:else if attendanceComplete}
                  {#if todayAttendance?.status === "absent"}
                    <div
                      class="flex items-center gap-4 p-5 bg-rose-50/70 border border-rose-200 rounded-xl justify-center"
                    >
                      <div
                        class="w-12 h-12 bg-rose-100 text-rose-600 rounded-full flex items-center justify-center shadow-sm shrink-0"
                      >
                        <span class="material-symbols-outlined text-2xl"
                          >close</span
                        >
                      </div>
                      <div>
                        <h4
                          class="font-bold text-rose-800 text-lg leading-tight"
                        >
                          Presensi Telah Ditutup
                        </h4>
                        <p class="text-rose-600 font-medium text-sm mt-0.5">
                          Status: Tidak Hadir (lewat batas waktu)
                        </p>
                      </div>
                    </div>
                  {:else}
                    <div
                      class="flex items-center gap-4 p-5 bg-emerald-50/50 border border-emerald-100 rounded-xl justify-center"
                    >
                      <div
                        class="w-12 h-12 bg-emerald-100 text-emerald-600 rounded-full flex items-center justify-center shadow-sm shrink-0"
                      >
                        <span class="material-symbols-outlined text-2xl"
                          >check</span
                        >
                      </div>
                      <div>
                        <h4
                          class="font-bold text-emerald-800 text-lg leading-tight"
                        >
                          Selesai Hari Ini
                        </h4>
                        <p class="text-emerald-600 font-medium text-sm mt-0.5">
                          Status: <span class="capitalize"
                            >{todayAttendance?.status || "Completed"}</span
                          >
                        </p>
                      </div>
                    </div>
                  {/if}
                {:else if canCheckOut}
                  <!-- Show Check-In Time & Check-Out Button -->
                  <div class="flex flex-col sm:flex-row gap-3">
                    <div
                      class="flex-1 action-pill pill-success text-emerald-700 border-emerald-100 shadow-sm text-sm"
                    >
                      <span class="material-symbols-outlined mr-2"
                        >schedule</span
                      >
                      <span
                        >Masuk Pada: <strong class="font-mono text-base"
                          >{formatTime(todayAttendance?.check_in_time)}</strong
                        ></span
                      >
                    </div>
                    <button
                      onclick={handleCheckOut}
                      disabled={checkingOut}
                      class="flex-1 action-pill pill-primary text-sm sm:text-base disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
                    >
                      {#if checkingOut}
                        <span
                          class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"
                        ></span>
                        Processing...
                      {:else}
                        <span class="material-symbols-outlined mr-2"
                          >logout</span
                        > PRESENSI KELUAR
                      {/if}
                    </button>
                  </div>
                {:else}
                  <div class="flex flex-col sm:flex-row gap-3">
                    <!-- Check-In Button -->
                    <button
                      onclick={handleCheckIn}
                      disabled={!canCheckIn || checkingIn}
                      class="flex-1 action-pill pill-primary text-sm sm:text-base disabled:cursor-not-allowed shadow-sm"
                    >
                      {#if checkingIn}
                        <span
                          class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"
                        ></span>
                        Processing...
                      {:else if gpsStatus === "loading"}
                        <span
                          class="material-symbols-outlined animate-spin mr-2"
                          >progress_activity</span
                        > Menunggu GPS...
                      {:else if gpsStatus === "denied"}
                        <span class="material-symbols-outlined mr-2"
                          >warning</span
                        > GPS Ditolak
                      {:else if gpsStatus === "error"}
                        <span class="material-symbols-outlined mr-2">error</span
                        > GPS Error
                      {:else if distance === null}
                        <span class="material-symbols-outlined mr-2"
                          >near_me</span
                        > Menunggu Lokasi...
                      {:else if distance > OFFICE.maxDistance}
                        <span class="material-symbols-outlined mr-2"
                          >location_on</span
                        >
                        Terlalu Jauh ({distance}m)
                      {:else}
                        <span class="material-symbols-outlined mr-2">login</span
                        > PRESENSI MASUK
                      {/if}
                    </button>

                    <!-- Permission Button -->
                    <button
                      onclick={() => (showPermissionModal = true)}
                      class="flex-1 action-pill pill-secondary text-sm sm:text-base shadow-sm"
                    >
                      <span class="material-symbols-outlined mr-2">sick</span> Izin
                      / Sakit
                    </button>
                  </div>
                {/if}
              </div>
            </div>
          </div>
        </div>

        <!-- Right Column for Presensi Mode -->
        <div class="space-y-6">
          <!-- Weekly Attendance -->
          <div class="card p-4 sm:p-6 h-[425px] flex flex-col">
            <h3
              class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2 shrink-0"
            >
              <span class="material-symbols-outlined text-indigo-500"
                >date_range</span
              > Presensi Minggu Ini
            </h3>
            <div class="flex-1 overflow-y-auto pr-1 space-y-3">
              {#if weeklyAttendanceList.length}
                {#each weeklyAttendanceList as item (item.date || item.id)}
                  <div
                    class="border border-slate-100 rounded-lg px-3 py-2 flex items-center justify-between"
                  >
                    <div class="flex flex-col">
                      <span class="text-sm font-semibold text-slate-800">
                        {formatDate(
                          item.date || item.check_in_time || item.created_at,
                        )}
                      </span>
                      <span class="text-xs text-slate-500">
                        Masuk: {formatTime(item.check_in_time)} · Pulang: {formatTime(
                          item.check_out_time,
                        )}
                      </span>
                    </div>
                    <span
                      class={`text-xs font-bold px-2 py-1 rounded-full border ${getStatusColor(item.status)}`}
                    >
                      {getStatusLabel(item.status)}
                    </span>
                  </div>
                {/each}
              {:else}
                <p class="text-sm text-slate-500">
                  Belum ada data presensi minggu ini.
                </p>
              {/if}
            </div>
          </div>

          <!-- Attendance Percentage -->
          <div class="card p-4 sm:p-6">
            <h3
              class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2"
            >
              <span class="material-symbols-outlined text-emerald-500"
                >percent</span
              > Persentase Kehadiran
            </h3>
            <div class="text-center">
              <div class="text-5xl font-black text-slate-800 mb-2">
                {dashboardData.attendancePercentage}%
              </div>
              <div class="h-2 bg-slate-100 rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-emerald-500 to-teal-500 rounded-full transition-all duration-500"
                  style="width: {dashboardData.attendancePercentage}%"
                ></div>
              </div>
            </div>
          </div>
        </div>
      {:else}
        <!-- TUGAS MODE: Tasks Card (Left 2/3) & Task Charts (Right 1/3) -->
        <div class="lg:col-span-2 space-y-6">
          <!-- TASKS CARD -->
          <div class="card p-0 overflow-hidden">
            <div
              class="px-5 py-5 sm:px-8 sm:py-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50"
            >
              <h3
                class="font-bold text-lg text-slate-800 flex items-center gap-2"
              >
                <span class="material-symbols-outlined text-teal-500"
                  >assignment</span
                > Tugas Aktif
              </h3>
              <a
                href="/tasks"
                class="text-xs font-bold text-teal-600 hover:text-teal-700 uppercase tracking-wider hover:underline"
              >
                Lihat Semua
              </a>
            </div>

            <div class="p-4 sm:p-6 space-y-6">
              <!-- Task Stats -->
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
                <div
                  class="bg-slate-50 p-3 sm:p-4 rounded-xl border border-slate-100 text-center hover:bg-white hover:shadow-md transition-all"
                >
                  <div class="text-2xl font-black text-slate-700">
                    {dashboardData.totalTasks}
                  </div>
                  <div
                    class="text-[10px] font-bold text-slate-400 uppercase tracking-wider mt-1"
                  >
                    Total
                  </div>
                </div>
                <div
                  class="bg-amber-50 p-3 sm:p-4 rounded-xl border border-amber-100 text-center hover:bg-white hover:shadow-md transition-all"
                >
                  <div class="text-2xl font-black text-amber-600">
                    {dashboardData.pendingTasks}
                  </div>
                  <div
                    class="text-[10px] font-bold text-amber-600/70 uppercase tracking-wider mt-1"
                  >
                    Pending
                  </div>
                </div>
                <div
                  class="bg-indigo-50 p-3 sm:p-4 rounded-xl border border-indigo-100 text-center hover:bg-white hover:shadow-md transition-all"
                >
                  <div class="text-2xl font-black text-indigo-600">
                    {dashboardData.inProgressTasks}
                  </div>
                  <div
                    class="text-[10px] font-bold text-indigo-600/70 uppercase tracking-wider mt-1"
                  >
                    Proses
                  </div>
                </div>
                <div
                  class="bg-emerald-50 p-3 sm:p-4 rounded-xl border border-emerald-100 text-center hover:bg-white hover:shadow-md transition-all"
                >
                  <div class="text-2xl font-black text-emerald-600">
                    {dashboardData.completedTasks}
                  </div>
                  <div
                    class="text-[10px] font-bold text-emerald-600/70 uppercase tracking-wider mt-1"
                  >
                    Selesai
                  </div>
                </div>
              </div>

              <!-- Progress Bar -->
              {#if dashboardData.totalTasks > 0}
                <div>
                  <div class="flex justify-between items-center mb-2">
                    <span
                      class="text-xs font-bold text-slate-600 uppercase tracking-wider"
                      >Progress</span
                    >
                    <span class="text-sm font-black text-slate-800">
                      {Math.round(
                        (dashboardData.completedTasks /
                          dashboardData.totalTasks) *
                          100,
                      )}%
                    </span>
                  </div>
                  <div class="h-3 bg-slate-100 rounded-full overflow-hidden">
                    <div
                      class="h-full bg-gradient-to-r from-emerald-500 to-teal-500 rounded-full transition-all duration-500"
                      style="width: {(dashboardData.completedTasks /
                        dashboardData.totalTasks) *
                        100}%"
                    ></div>
                  </div>
                </div>
              {/if}

              <!-- Recent Tasks -->
              {#if dashboardData.recentTasks?.length > 0}
                <div class="space-y-3">
                  <h4
                    class="text-xs font-bold text-slate-500 uppercase tracking-wider"
                  >
                    Tugas Terbaru
                  </h4>
                  {#each dashboardData.recentTasks as task}
                    <a
                      href="/tasks/{task.id}"
                      class="block p-3 bg-slate-50 rounded-lg border border-slate-200 hover:shadow-sm hover:bg-white transition-all"
                    >
                      <div class="flex justify-between items-start mb-2">
                        <h5 class="font-semibold text-slate-800 text-sm">
                          {task.title}
                        </h5>
                        <span
                          class="px-2 py-0.5 rounded text-[10px] font-bold border {getStatusBadgeClass(
                            task.status,
                          )}"
                        >
                          {getStatusLabel(task.status)}
                        </span>
                      </div>
                      <div
                        class="flex items-center gap-3 text-xs text-slate-500"
                      >
                        {#if task.deadline}
                          <span class="flex items-center"
                            ><span
                              class="material-symbols-outlined text-[16px] mr-1"
                              >calendar_today</span
                            >{formatDate(task.deadline)}</span
                          >
                        {/if}
                        {#if task.priority}
                          <span class="capitalize flex items-center">
                            <span
                              class="material-symbols-outlined text-[16px] mr-1"
                              class:text-red-500={task.priority === "high"}
                              class:text-amber-500={task.priority === "medium"}
                              class:text-slate-400={task.priority === "low"}
                              >flag</span
                            >
                            {task.priority}
                          </span>
                        {/if}
                      </div>
                    </a>
                  {/each}
                </div>
              {:else}
                <div class="text-center py-8 text-slate-400">
                  <span class="material-symbols-outlined text-4xl mb-2"
                    >inbox</span
                  >
                  <p class="text-sm">Tidak ada tugas</p>
                </div>
              {/if}
            </div>
          </div>
        </div>

        <!-- Right Column for Tugas Mode -->
        <div class="space-y-6">
          <!-- Task Breakdown Chart -->
          <div class="card p-4 sm:p-6">
            <h3
              class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2"
            >
              <span class="material-symbols-outlined text-violet-500"
                >pie_chart</span
              > Status Tugas
            </h3>
            <div class="h-[500px]">
              <canvas id="taskProgressChart"></canvas>
            </div>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Late Reason Modal -->
{#if showLateModal}
  <div
    class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
    role="button"
    tabindex="0"
    onclick={(e) => e.target === e.currentTarget && (showLateModal = false)}
    onkeydown={(e) =>
      (e.key === "Enter" || e.key === " ") && (showLateModal = false)}
  >
    <div class="bg-white rounded-2xl p-6 max-w-md w-full shadow-2xl">
      <h3 class="text-xl font-bold text-slate-800 mb-4">Alasan Terlambat</h3>
      <p class="text-sm text-slate-600 mb-4">
        Anda terlambat. Mohon berikan alasan:
      </p>
      <textarea
        bind:value={lateReason}
        placeholder="Tulis alasan..."
        class="w-full px-4 py-3 border border-slate-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
        rows="3"
      ></textarea>
      <div class="flex gap-3 mt-4">
        <button
          onclick={() => (showLateModal = false)}
          class="btn flex-1 bg-slate-100 text-slate-700 hover:bg-slate-200"
        >
          Batal
        </button>
        <button
          onclick={handleCheckIn}
          class="btn flex-1 bg-indigo-600 text-white hover:bg-indigo-700"
        >
          Lanjutkan Check-In
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Permission Modal -->
{#if showPermissionModal}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6"
    role="button"
    tabindex="0"
    onclick={(e) =>
      e.target === e.currentTarget && (showPermissionModal = false)}
    onkeydown={(e) =>
      (e.key === "Enter" || e.key === " ") && (showPermissionModal = false)}
  >
    <div
      class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity"
    ></div>

    <div
      class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg flex flex-col overflow-hidden"
    >
      <div
        class="p-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/50"
      >
        <h3 class="font-bold text-lg text-slate-800">Pengajuan Izin / Sakit</h3>
        <button
          onclick={() => (showPermissionModal = false)}
          class="text-slate-400 hover:text-slate-600"
        >
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <div class="p-6 space-y-4">
        <!-- Kategori -->
        <div>
          <label
            class="block text-sm font-semibold text-slate-700 mb-2"
            for="perm-status">Kategori</label
          >
          <div class="flex gap-3">
            <label class="flex-1 cursor-pointer">
              <input
                type="radio"
                value="sick"
                bind:group={permissionStatus}
                class="peer sr-only"
              />
              <div
                class="text-center p-3 rounded-xl border border-slate-200 text-slate-600 peer-checked:bg-indigo-50 peer-checked:border-indigo-500 peer-checked:text-indigo-700 transition-all"
              >
                🤒 Sakit
              </div>
            </label>
            <label class="flex-1 cursor-pointer">
              <input
                type="radio"
                value="permission"
                bind:group={permissionStatus}
                class="peer sr-only"
              />
              <div
                class="text-center p-3 rounded-xl border border-slate-200 text-slate-600 peer-checked:bg-indigo-50 peer-checked:border-indigo-500 peer-checked:text-indigo-700 transition-all"
              >
                📝 Izin
              </div>
            </label>
          </div>
        </div>

        <!-- Keterangan -->
        <div>
          <label
            class="block text-sm font-semibold text-slate-700 mb-2"
            for="perm-notes">Keterangan</label
          >
          <textarea
            bind:value={permissionNotes}
            rows="3"
            class="w-full px-4 py-3 border border-slate-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none text-sm"
            placeholder="Jelaskan alasan Anda..."
          ></textarea>
        </div>

        <!-- File Upload -->
        <div>
          <label
            class="block text-sm font-semibold text-slate-700 mb-2"
            for="file-upload">Bukti Pendukung (Opsional)</label
          >
          <label
            class="block w-full p-4 border-2 border-dashed border-slate-300 rounded-xl text-center cursor-pointer hover:bg-slate-50 hover:border-indigo-400 transition-all"
          >
            {#if permissionFile}
              <span class="text-indigo-600 font-medium text-sm"
                >{permissionFile.name}</span
              >
            {:else}
              <span class="text-slate-500 text-sm"
                >Klik untuk upload surat dokter/dokumen</span
              >
            {/if}
            <input
              id="file-upload"
              type="file"
              hidden
              bind:this={fileInput}
              onchange={(e) =>
                (permissionFile = e.currentTarget.files?.[0] || null)}
            />
          </label>
        </div>

        <!-- Submit -->
        <button
          onclick={handlePermissionSubmit}
          disabled={isSubmittingPermission}
          class="w-full py-3 px-4 bg-slate-900 text-white rounded-xl font-semibold hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-sm"
        >
          {isSubmittingPermission ? "Mengirim..." : "Kirim Pengajuan"}
        </button>
      </div>
    </div>
  </div>
{/if}

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
      "FILL" 0,
      "wght" 200,
      "GRAD" 0,
      "opsz" 20;
  }

  :global(.user-marker),
  :global(.office-marker) {
    z-index: 1000 !important;
  }

  @keyframes pulse {
    0%,
    100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.1);
      opacity: 0.8;
    }
  }

  .stat-card,
  .card {
    background: rgba(255, 255, 255, 0.85) !important;
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }

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

  .action-pill {
    padding: 0.5rem 0.9rem; /* slimmer */
    border-radius: 9999px;
    border: 1px solid transparent;
    min-height: 2.5rem; /* reduced thickness */
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
    width: 7.5rem;
    height: 2.5rem;
    border-radius: 0.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06);
    transition: all 0.2s;
    cursor: pointer;
  }
  .map-reset:hover {
    background: #f3f4f6;
    transform: scale(1.05);
  }

  .highlight {
    color: #10b981;
  }

  .info-bar {
    display: flex;
    justify-content: space-around;
    background: #f8fafc;
    padding: 12px;
    border-radius: 12px;
    border: 1px solid #f1f5f9;
  }

  .info-item {
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .info-item .label {
    font-size: 11px;
    font-weight: 600;
    color: #94a3b8;
    text-transform: uppercase;
  }
  .info-item .value {
    font-size: 15px;
    font-weight: 600;
    font-family: monospace;
  }
</style>
