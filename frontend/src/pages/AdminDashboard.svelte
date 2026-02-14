<script>
  import { onMount, onDestroy, tick } from "svelte";
  import { api } from "../lib/api.js";
  import { auth } from "../lib/auth.svelte.js";
  import { getAvatarUrl } from "../lib/utils.js";
  import Chart from "chart.js/auto";
  import L from "leaflet";

  // ==================== VIEW MODE STATE ====================
  let viewMode = $state("stats"); // 'stats' | 'geofence'
  $effect(() => {
    if (!isAdmin && viewMode === "geofence") {
      viewMode = "stats";
    }
  });

  // ==================== STATS TAB STATE (from AdminDashboard) ====================
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

  // ==================== GEOFENCE TAB STATE ====================
  // Office Configuration
  let OFFICE = $state({
    lat: -7.0355,
    lng: 110.4746,
    name: "Kantor Pusat",
    maxDistance: 1000, // meters
  });

  // GPS State (for debugging simulation)
  let gpsStatus = $state("loading"); // loading, ready, error, denied
  let gps = $state({ lat: null, lng: null, error: null });
  let simulatingLocation = $state(false);

  // Map State
  let map = $state(null);
  let userMarker = $state(null);
  let officeMarker = $state(null);
  let radiusCircle = $state(null);
  let watchId = null;

  // Office Management State (from Settings)
  let officeList = $state([]);
  let activeOfficeId = $state(null);
  let selectedOfficeId = $state("");
  let loadingOffices = $state(false);
  let creatingOffice = $state(false);
  let deletingOffice = $state(false);
  let settingActive = $state(false);
  let showAddModal = $state(false);
  let savingSettings = $state(false);

  // Search State
  let searchingPlaces = $state(false);
  let searchResults = $state([]);

  let newOffice = $state({
    name: "",
    latitude: "",
    longitude: "",
    radius_meters: 1000,
    address: "",
  });

  // Global settings
  let globalSettings = $state({
    max_checkin_distance: 1000,
  });

  // ==================== DERIVED STATE ====================
  const isAdmin = $derived(auth.user?.role === "admin");
  const isSupervisor = $derived(
    auth.user?.role === "supervisor" || auth.user?.role === "pembimbing",
  );

  // Current selected office
  let currentOffice = $derived(
    selectedOfficeId && selectedOfficeId !== "new"
      ? officeList.find((o) => o.id == selectedOfficeId)
      : null,
  );

  // Calculate distance from office
  const distance = $derived(
    gps.lat && gps.lng
      ? Math.round(getDistance(gps.lat, gps.lng, OFFICE.lat, OFFICE.lng))
      : null,
  );

  // Check if location is within bounds
  const isWithinBounds = $derived(
    distance !== null && distance <= OFFICE.maxDistance,
  );

  // ==================== STATS TAB FUNCTIONS ====================

  function buildAttendanceTrend(records = []) {
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

  async function fetchDashboardData() {
    loading = true;
    try {
      const pendingReq = isAdmin
        ? api.getInterns({ status: "pending", limit: 5 })
        : Promise.resolve({ data: [] });
      const pendingSupervisorsReq = isAdmin
        ? api.getSupervisors({
            status: "pending",
            limit: 5,
          })
        : Promise.resolve({ data: [] });
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
        console.warn(
          "Backend dashboard endpoint not ready or restricted:",
          err,
        );
      }

      const [pendingRes, pendingSupervisorsRes, tasksRes, attendanceRes] =
        await Promise.all([
          pendingReq.catch(() => ({ data: [] })),
          pendingSupervisorsReq.catch(() => ({ data: [] })),
          submittedTasksReq.catch(() => ({ data: [] })),
          attendanceReq.catch(() => ({ data: [] })),
        ]);

      const attendanceRecords = attendanceRes?.data || [];
      const todayStr = new Date().toLocaleDateString("en-CA");
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
            stats.pending_registrations || pendingRes.meta?.total || 0,
          pendingSupervisors:
            stats.pending_supervisors || pendingSupervisorsRes.meta?.total || 0,
          pendingInternsList: pendingRes.data || [],
          pendingSupervisorsList: pendingSupervisorsRes.data || [],
          submittedTasks: tasksRes.data || [],
        };
      } else {
        // Fallback manual calculation
        const allInternsRes = await api.getInterns({ limit: 1000 });
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
          console.warn("Failed to calculate tasks manually");
        }

        dashboardData = {
          totalInterns: activeCount,
          completedOnTime: taskOnTime,
          completedLate: taskLate,
          presentToday: presentTodayCount,
          overdueTasks: taskOverdue,
          inProgressTasks: taskInProgress,
          attendanceTrend: buildAttendanceTrend(attendanceRecords),
          pendingRegistrations: pendingRes.meta?.total || 0,
          pendingSupervisors: pendingSupervisorsRes.meta?.total || 0,
          pendingInternsList: pendingRes.data || [],
          pendingSupervisorsList: pendingSupervisorsRes.data || [],
          submittedTasks: tasksRes.data || [],
        };
      }
    } catch (err) {
      console.error("Failed to fetch dashboard:", err);
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
        `Apakah Anda yakin ingin MENOLAK dan MENGHAPUS pendaftaran "${name}"?`,
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
    if (!confirm(`Apakah Anda yakin ingin MENOLAK pembimbing "${name}"?`))
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

  // Re-init charts when stats view is shown
  $effect(() => {
    if (!loading && dashboardData && viewMode === "stats") {
      tick().then(initCharts);
    }
  });

  // ==================== GEOFENCE TAB FUNCTIONS ====================

  // Haversine formula
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

  function initGeolocation() {
    if (!navigator.geolocation) {
      gpsStatus = "error";
      gps.error = "Geolocation not supported";
      return;
    }

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

  function initMap() {
    if (typeof window === "undefined") return;

    const mapElement = document.getElementById("map");
    if (!mapElement) return;

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

  async function fetchSettings() {
    try {
      const res = await api.getSettings();
      const list = res?.data || [];
      const map = {};
      list.forEach((item) => {
        if (item?.key) map[item.key] = item.value;
      });

      if (map.active_office_id) {
        activeOfficeId = Number(map.active_office_id);
        if (!selectedOfficeId) selectedOfficeId = activeOfficeId;
      }

      OFFICE = {
        lat: parseFloat(map.office_latitude || OFFICE.lat),
        lng: parseFloat(map.office_longitude || OFFICE.lng),
        name: map.office_name || OFFICE.name,
        maxDistance: parseInt(
          map.max_checkin_distance || map.office_radius || OFFICE.maxDistance,
          10,
        ),
      };

      globalSettings = {
        max_checkin_distance: parseInt(map.max_checkin_distance || 100, 10),
      };

      console.log("Office settings loaded:", OFFICE);
    } catch (err) {
      console.warn("Failed to load office settings:", err);
    }
  }

  async function fetchOffices() {
    loadingOffices = true;
    try {
      const res = await api.getOffices();
      if (res.success) {
        officeList = res.data || [];
        if (!selectedOfficeId && officeList.length > 0) {
          selectedOfficeId = activeOfficeId || officeList[0].id;
        }
      }
    } catch (err) {
      console.error("Failed to load offices", err);
    } finally {
      loadingOffices = false;
    }
  }

  function handleLocationChange() {
    if (selectedOfficeId === "new") {
      selectedOfficeId = "";
      showAddModal = true;
    }
  }

  function closeAddModal() {
    showAddModal = false;
    newOffice = {
      name: "",
      latitude: "",
      longitude: "",
      radius_meters: 100,
      address: "",
    };
    searchResults = [];
  }

  async function searchPlaces(query) {
    if (!query) return;
    searchingPlaces = true;
    searchResults = [];
    try {
      const res = await api.searchPlaces(query);
      if (res.success) {
        searchResults = res.data || [];
        if (searchResults.length === 0) {
          alert("Tidak ditemukan lokasi dengan kata kunci tersebut.");
        }
      }
    } catch (err) {
      console.error(err);
      alert("Gagal mencari lokasi: " + err.message);
    } finally {
      searchingPlaces = false;
    }
  }

  function selectPlace(place) {
    newOffice = {
      ...newOffice,
      name: place.name,
      latitude: String(place.latitude),
      longitude: String(place.longitude),
      address: place.address,
    };
    searchResults = [];
  }

  async function handleAddOffice() {
    if (!newOffice.name || !newOffice.latitude || !newOffice.longitude) {
      alert("Mohon lengkapi nama dan koordinat.");
      return;
    }
    creatingOffice = true;
    try {
      const payload = {
        ...newOffice,
        latitude: parseFloat(newOffice.latitude),
        longitude: parseFloat(newOffice.longitude),
        radius_meters: parseInt(String(newOffice.radius_meters)) || 100,
      };
      const res = await api.createOffice(payload);
      if (res.success) {
        await fetchOffices();
        selectedOfficeId = res.data.id;
        closeAddModal();
      }
    } catch (err) {
      alert("Gagal menambah lokasi: " + err.message);
    } finally {
      creatingOffice = false;
    }
  }

  async function handleDeleteOffice() {
    if (!currentOffice) return;
    if (!confirm(`Hapus lokasi "${currentOffice.name}"?`)) return;

    deletingOffice = true;
    try {
      await api.deleteOffice(currentOffice.id);
      await fetchOffices();
      selectedOfficeId = activeOfficeId || officeList[0]?.id || "";
    } catch (err) {
      alert("Gagal menghapus lokasi: " + err.message);
    } finally {
      deletingOffice = false;
    }
  }

  async function handleSetActive() {
    if (!currentOffice) return;
    settingActive = true;
    try {
      await api.setActiveOffice(currentOffice.id);
      activeOfficeId = currentOffice.id;
      await fetchSettings();
      alert(`Lokasi aktif diubah ke "${currentOffice.name}"`);
    } catch (err) {
      alert("Gagal mengatur lokasi aktif: " + err.message);
    } finally {
      settingActive = false;
    }
  }

  async function saveGlobalSettings() {
    savingSettings = true;
    try {
      await api.updateSettings({
        max_checkin_distance: String(globalSettings.max_checkin_distance),
      });
      alert("Pengaturan radius global tersimpan.");
      await fetchSettings();
    } catch (err) {
      alert("Gagal menyimpan: " + (err.message || "unknown error"));
    } finally {
      savingSettings = false;
    }
  }

  function handleSimulateLocation() {
    simulatingLocation = true;
    setTimeout(() => {
      simulatingLocation = false;
    }, 1000);
  }

  function updateMapOffice() {
    if (!map || !OFFICE.lat || !OFFICE.lng) return;

    map.setView([OFFICE.lat, OFFICE.lng], 15);

    if (officeMarker) {
      officeMarker
        .setLatLng([OFFICE.lat, OFFICE.lng])
        .bindPopup(OFFICE.name)
        .openPopup();
    }

    if (radiusCircle) {
      radiusCircle
        .setLatLng([OFFICE.lat, OFFICE.lng])
        .setRadius(OFFICE.maxDistance);
    }
  }

  // Re-init map or charts when view mode changes OR when OFFICE data updates
  $effect(() => {
    // Dependency tracking
    viewMode;
    OFFICE.lat;
    OFFICE.lng;
    OFFICE.maxDistance;

    if (!loading) {
      if (viewMode === "geofence") {
        setTimeout(() => {
          initMap();
          updateMapPosition();
          updateMapOffice();
        }, 50);
      } else if (viewMode === "stats") {
        setTimeout(initCharts, 50);
      }
    }
  });

  function handleKeydown(e) {
    if (e.key === "Escape" && showAddModal) closeAddModal();
  }

  // ==================== LIFECYCLE ====================

  onMount(async () => {
    loading = true;
    try {
      const tasks = [fetchSettings(), fetchDashboardData()];
      if (isAdmin) {
        tasks.push(fetchOffices());
      }
      await Promise.all(tasks.map((t) => t.catch((e) => console.error(e))));

      if (isAdmin) {
        initGeolocation();
      }
    } finally {
      loading = false;
    }

    setTimeout(() => {
      if (isAdmin && viewMode === "geofence") {
        initMap();
        updateMapPosition();
      } else if (viewMode === "stats") {
        initCharts();
      }
    }, 100);
  });

  onDestroy(() => {
    if (watchId) {
      navigator.geolocation.clearWatch(watchId);
    }
    if (map) {
      map.remove();
    }
    if (taskPieChart) taskPieChart.destroy();
    if (attendanceTrendChart) attendanceTrendChart.destroy();
  });
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
            >{auth.user?.name || "Admin"}</span
          >
        </h2>
      </div>
      <div class="flex items-center gap-3">
        <button
          onclick={() => (viewMode = "stats")}
          class="px-5 py-2 rounded-full text-sm font-semibold transition-all flex items-center justify-center gap-2 {viewMode ===
          'stats'
            ? 'bg-slate-900 text-white border border-transparent'
            : 'bg-white text-slate-900 border border-slate-200 hover:border-slate-300'}"
        >
          <span class="material-symbols-outlined text-[18px]">analytics</span>
          Stats
        </button>
        {#if isAdmin}
          <button
            onclick={() => (viewMode = "geofence")}
            class="px-5 py-2 rounded-full text-sm font-semibold transition-all flex items-center justify-center gap-2 {viewMode ===
            'geofence'
              ? 'bg-slate-900 text-white border border-transparent'
              : 'bg-white text-slate-900 border border-slate-200 hover:border-slate-300'}"
          >
            <span class="material-symbols-outlined text-[18px]"
              >location_on</span
            >
            Geofence
          </button>
        {/if}
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
          >
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
            <line x1="16" y1="2" x2="16" y2="6"></line>
            <line x1="8" y1="2" x2="8" y2="6"></line>
            <line x1="3" y1="10" x2="21" y2="10"></line>
          </svg>
          {currentDate}
        </div>
      </div>
    </div>

    <!-- Mobile Header -->
    <div class="md:hidden flex flex-col gap-4">
      <h2 class="text-xl font-semibold text-slate-800 tracking-tight">
        Admin Dashboard, <span class="highlight"
          >{auth.user?.name || "Admin"}</span
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
          >
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
            <line x1="16" y1="2" x2="16" y2="6"></line>
            <line x1="8" y1="2" x2="8" y2="6"></line>
            <line x1="3" y1="10" x2="21" y2="10"></line>
          </svg>
          {new Date().toLocaleDateString("id-ID", {
            day: "2-digit",
            month: "short",
            year: "numeric",
          })}
        </div>
      </div>
      <div class="flex gap-2 w-full">
        <button
          onclick={() => (viewMode = "stats")}
          class="flex-1 px-5 py-2 rounded-full text-sm font-semibold transition-all flex items-center justify-center gap-2 {viewMode ===
          'stats'
            ? 'bg-slate-900 text-white border border-transparent'
            : 'bg-white text-slate-900 border border-slate-200 hover:border-slate-300'}"
        >
          <span class="material-symbols-outlined text-[18px]">analytics</span>
          Stats
        </button>
        {#if isAdmin}
          <button
            onclick={() => (viewMode = "geofence")}
            class="flex-1 px-5 py-2 rounded-full text-sm font-semibold transition-all flex items-center justify-center gap-2 {viewMode ===
            'geofence'
              ? 'bg-slate-900 text-white border border-transparent'
              : 'bg-white text-slate-900 border border-slate-200 hover:border-slate-300'}"
          >
            <span class="material-symbols-outlined text-[18px]"
              >location_on</span
            >
            Geofence
          </button>
        {/if}
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
    <!-- STATS MODE -->
    {#if viewMode === "stats"}
      <div class="space-y-5 animate-fade-in">
        <!-- Pending Registrations -->
        {#if isAdmin && (dashboardData.pendingInternsList.length > 0 || dashboardData.pendingSupervisorsList.length > 0)}
          <div class="tasks-section approval-card animate-slide-up">
            <div class="flex flex-col mb-4">
              <div class="flex justify-between items-center mb-2">
                <div class="flex items-center gap-3">
                  <div class="p-2 bg-amber-50 rounded-lg text-amber-600">
                    <svg
                      width="20"
                      height="20"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                      <circle cx="8.5" cy="7" r="4" />
                      <polyline points="17 11 19 13 23 9" />
                    </svg>
                  </div>
                  <h3 class="font-bold text-base sm:text-lg text-slate-800">
                    Pendaftaran Baru
                  </h3>
                </div>
                <a href="/interns?status=pending" class="link-view-all">
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
              <p class="pl-12 text-xs sm:text-sm text-slate-600 mt-1">
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

            <div class="task-grid">
              {#each dashboardData.pendingInternsList as intern}
                <div class="task-item">
                  <div
                    class="task-icon {intern.avatar
                      ? 'p-0 overflow-hidden'
                      : 'task-icon-emerald bg-emerald-50 border-emerald-600 border-2 text-emerald-600'}"
                  >
                    {#if intern.avatar}
                      <img
                        src={getAvatarUrl(intern.avatar)}
                        alt={intern.full_name}
                        class="w-full h-full object-cover"
                      />
                    {:else}
                      <div class="avatar-placeholder">
                        {intern.full_name?.charAt(0) || "I"}
                      </div>
                    {/if}
                  </div>
                  <div class="task-info">
                    <h4 class="font-semibold text-slate-800 text-sm truncate">
                      {intern.full_name}
                    </h4>
                    <p class="text-xs text-slate-500">
                      {intern.school || "-"} • {intern.department || "-"}
                    </p>
                  </div>
                  <div class="flex gap-1 shrink-0">
                    <button
                      class="btn-icon deny"
                      onclick={() => handleDeny(intern.id, intern.full_name)}
                      title="Tolak"
                    >
                      <svg
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                      </svg>
                    </button>
                    <button
                      class="btn-icon approve"
                      onclick={() => handleApprove(intern.id, intern.full_name)}
                      title="Terima"
                    >
                      <svg
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <polyline points="20 6 9 17 4 12" />
                      </svg>
                    </button>
                  </div>
                </div>
              {/each}

              {#each dashboardData.pendingSupervisorsList as supervisor}
                <div class="task-item">
                  <div
                    class="task-icon {supervisor.avatar
                      ? 'p-0 overflow-hidden'
                      : 'task-icon-purple bg-purple-100 border-purple-600 border-2 text-purple-600'}"
                  >
                    {#if supervisor.avatar}
                      <img
                        src={getAvatarUrl(supervisor.avatar)}
                        alt={supervisor.full_name}
                        class="w-full h-full object-cover"
                      />
                    {:else}
                      {supervisor.full_name?.charAt(0) || "P"}
                    {/if}
                  </div>
                  <div class="task-info">
                    <h4 class="font-semibold text-slate-800 text-sm truncate">
                      {supervisor.full_name}
                    </h4>
                    <p class="text-xs text-slate-500">
                      {supervisor.institution || "-"} • Pembimbing
                    </p>
                  </div>
                  <div class="flex gap-1 shrink-0">
                    <button
                      class="btn-icon deny"
                      onclick={() =>
                        handleDenySupervisor(
                          supervisor.id,
                          supervisor.full_name,
                        )}
                      title="Tolak"
                    >
                      <svg
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                      </svg>
                    </button>
                    <button
                      class="btn-icon approve"
                      onclick={() =>
                        handleApproveSupervisor(
                          supervisor.id,
                          supervisor.full_name,
                        )}
                      title="Terima"
                    >
                      <svg
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <polyline points="20 6 9 17 4 12" />
                      </svg>
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Stats Grid -->
        <div class="stats-grid animate-slide-up" style="animation-delay: 0.1s;">
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

        <!-- Charts -->
        <div
          class="charts-grid animate-slide-up"
          style="animation-delay: 0.2s;"
        >
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
                <p class="text-sm text-slate-500 mt-3">
                  Belum ada data kehadiran
                </p>
              </div>
            {/if}
          </div>
        </div>

        <!-- Submitted Tasks -->
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
      </div>
    {:else}
      <!-- GEOFENCE MODE -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 animate-fade-in">
        <!-- Map Card (Left 2/3) -->
        <div class="lg:col-span-2 space-y-6">
          <div class="card p-0 overflow-hidden">
            <div
              class="px-3 py-2 sm:px-6 sm:py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50"
            >
              <h3
                class="font-bold text-lg text-slate-800 flex items-center gap-2"
              >
                <span class="material-symbols-outlined text-indigo-500"
                  >location_searching</span
                >
                Geofence Debugging
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

              <!-- Office Info -->
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
                  <span class="font-bold text-slate-800">{OFFICE.name}</span>
                </span>
              </div>

              <!-- Distance Info -->
              <div class="info-bar">
                <div class="info-item">
                  <span class="label">Jarak Kantor</span>
                  <span
                    class="value {isWithinBounds
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

              <!-- Debug Status -->
              <div class="bg-slate-50 p-4 rounded-xl border border-slate-200">
                <h4
                  class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-3"
                >
                  Debug Status
                </h4>
                <div class="space-y-2">
                  <div class="flex justify-between items-center">
                    <span class="text-sm text-slate-600">GPS Coordinates:</span>
                    <code
                      class="text-xs bg-white px-2 py-1 rounded border border-slate-200"
                    >
                      {gps.lat && gps.lng
                        ? `${gps.lat.toFixed(6)}, ${gps.lng.toFixed(6)}`
                        : "N/A"}
                    </code>
                  </div>
                  <div class="flex justify-between items-center">
                    <span class="text-sm text-slate-600">Within Bounds:</span>
                    <span
                      class="text-xs font-bold px-2 py-1 rounded {isWithinBounds
                        ? 'bg-emerald-100 text-emerald-700'
                        : 'bg-rose-100 text-rose-700'}"
                    >
                      {isWithinBounds ? "YES ✓" : "NO ✗"}
                    </span>
                  </div>
                  <div class="flex justify-between items-center">
                    <span class="text-sm text-slate-600">Office Coords:</span>
                    <code
                      class="text-xs bg-white px-2 py-1 rounded border border-slate-200"
                    >
                      {OFFICE.lat}, {OFFICE.lng}
                    </code>
                  </div>
                </div>
              </div>

              <!-- Simulate Button -->
              <button
                onclick={handleSimulateLocation}
                disabled={simulatingLocation || gpsStatus !== "ready"}
                class="w-full py-3 px-4 bg-indigo-600 text-white rounded-xl font-semibold hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-sm flex items-center justify-center gap-2"
              >
                {#if simulatingLocation}
                  <span
                    class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"
                  ></span>
                  Simulating...
                {:else}
                  <span class="material-symbols-outlined">refresh</span>
                  Refresh Location
                {/if}
              </button>
            </div>
          </div>
        </div>

        <!-- Location Settings Card (Right 1/3) -->
        <div class="space-y-6">
          <div class="card p-4 sm:p-6">
            <h3
              class="font-bold text-base text-slate-800 mb-4 flex items-center gap-2"
            >
              <span class="material-symbols-outlined text-emerald-500"
                >location_city</span
              >
              Lokasi Absensi
            </h3>

            <!-- Global Radius -->
            <div class="form-grid mb-4 border-b border-slate-100 pb-4">
              <div class="form-group">
                <label class="label" for="max-checkin-distance"
                  >Radius Maksimal (meter)</label
                >
                <input
                  id="max-checkin-distance"
                  class="input"
                  type="number"
                  min="10"
                  bind:value={globalSettings.max_checkin_distance}
                />
                <p class="help-text">
                  Jarak maksimal dari titik kantor (Global).
                </p>
              </div>
              <button
                onclick={saveGlobalSettings}
                disabled={savingSettings}
                class="px-4 py-2 bg-indigo-600 text-white rounded-lg font-medium text-sm hover:bg-indigo-700 transition-colors shadow-sm"
              >
                {savingSettings ? "Menyimpan..." : "Simpan Radius"}
              </button>
            </div>

            <!-- Location Selector -->
            <div class="flex items-end gap-3 mb-4">
              <div class="form-group flex-1">
                <label class="label" for="locationSelect">Pilih Lokasi</label>
                <div class="relative">
                  <select
                    id="locationSelect"
                    class="input appearance-none pr-8 cursor-pointer"
                    bind:value={selectedOfficeId}
                    onchange={handleLocationChange}
                    disabled={loadingOffices}
                  >
                    {#if loadingOffices}
                      <option value="" disabled>Memuat lokasi...</option>
                    {:else}
                      {#each officeList as office}
                        <option value={office.id}
                          >{office.name}
                          {office.id === activeOfficeId
                            ? "(Aktif)"
                            : ""}</option
                        >
                      {/each}
                      <option value="new" class="text-blue-600 font-semibold"
                        >+ Tambah Lokasi Baru...</option
                      >
                    {/if}
                  </select>
                  <div
                    class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-slate-500"
                  >
                    <span class="material-symbols-outlined text-sm"
                      >expand_more</span
                    >
                  </div>
                </div>
              </div>
            </div>

            <!-- Active Indicator -->
            {#if selectedOfficeId === activeOfficeId && activeOfficeId}
              <div
                class="bg-emerald-50 border border-emerald-200 rounded-lg p-3 flex items-center gap-3 mb-4"
              >
                <span class="material-symbols-outlined text-emerald-600"
                  >check_circle</span
                >
                <p class="text-sm text-emerald-700 font-medium">
                  Lokasi ini sedang aktif digunakan.
                </p>
              </div>
            {/if}

            <!-- Location Preview -->
            {#if currentOffice}
              <div class="preview-box mb-4">
                <div
                  class="p-3 border-b border-slate-200 flex justify-between items-center"
                >
                  <h4 class="font-bold text-slate-700">{currentOffice.name}</h4>
                  <span
                    class="text-xs bg-slate-100 text-slate-600 px-2 py-1 rounded-full border border-slate-200"
                  >
                    Radius: {currentOffice.radius_meters}m
                  </span>
                </div>
                <div class="p-3 grid grid-cols-2 gap-4">
                  <div>
                    <span class="text-xs text-slate-500 block mb-1"
                      >Latitude</span
                    >
                    <code
                      class="text-sm block bg-white border border-slate-200 rounded px-2 py-1"
                      >{currentOffice.latitude}</code
                    >
                  </div>
                  <div>
                    <span class="text-xs text-slate-500 block mb-1"
                      >Longitude</span
                    >
                    <code
                      class="text-sm block bg-white border border-slate-200 rounded px-2 py-1"
                      >{currentOffice.longitude}</code
                    >
                  </div>
                  <div class="col-span-2">
                    <span class="text-xs text-slate-500 block mb-1">Alamat</span
                    >
                    <p class="text-sm text-slate-700">
                      {currentOffice.address || "-"}
                    </p>
                  </div>
                </div>
              </div>

              <!-- Actions -->
              <div
                class="flex gap-3 justify-end border-t border-slate-100 pt-4"
              >
                {#if selectedOfficeId !== activeOfficeId}
                  <button
                    class="px-4 py-2 rounded-lg bg-slate-100 text-slate-600 font-medium text-sm hover:bg-slate-200 transition-colors"
                    onclick={handleDeleteOffice}
                    disabled={deletingOffice}
                  >
                    {deletingOffice ? "Menghapus..." : "Hapus"}
                  </button>
                  <button
                    class="px-4 py-2 rounded-lg bg-indigo-600 text-white font-medium text-sm hover:bg-indigo-700 transition-colors shadow-sm hover:shadow-md"
                    onclick={handleSetActive}
                    disabled={settingActive}
                  >
                    {settingActive ? "Mengatur..." : "Atur Sebagai Aktif"}
                  </button>
                {/if}
              </div>
            {:else}
              <div
                class="text-center py-8 text-slate-400 bg-slate-50 rounded-xl border-dashed border-2 border-slate-200"
              >
                <span class="material-symbols-outlined text-4xl mb-2 opacity-50"
                  >location_off</span
                >
                <p>Tidak ada lokasi yang dipilih.</p>
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  {/if}
</div>

<!-- Add Location Modal -->
{#if showAddModal}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-fade-in"
    onclick={closeAddModal}
    role="button"
    tabindex="0"
    onkeydown={handleKeydown}
  >
    <div
      class="bg-white rounded-2xl shadow-xl w-full max-w-md overflow-hidden animate-scale-up"
      onclick={(e) => e.stopPropagation()}
      role="button"
      tabindex="0"
      onkeydown={handleKeydown}
    >
      <div
        class="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50"
      >
        <h3 class="font-bold text-lg text-slate-800">Tambah Lokasi Baru</h3>
        <button
          onclick={closeAddModal}
          class="text-slate-400 hover:text-slate-600 transition-colors"
        >
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>
      <div class="p-6 space-y-4">
        <!-- Search -->
        <div class="form-group relative">
          <label class="label" for="search-place"
            >Cari Lokasi (Google Maps)</label
          >
          <div class="flex gap-2">
            <input
              id="search-place"
              class="input"
              type="text"
              placeholder="Contoh: Simpang Lima Semarang"
              onkeydown={(e) =>
                e.key === "Enter" && searchPlaces(e.currentTarget.value)}
            />
            <button
              class="px-3 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              onclick={(e) => {
                const input = /** @type {HTMLInputElement | null} */ (
                  document.getElementById("search-place")
                );
                if (input) searchPlaces(input.value);
              }}
              disabled={searchingPlaces}
            >
              {#if searchingPlaces}
                <span class="material-symbols-outlined animate-spin text-sm"
                  >refresh</span
                >
              {:else}
                <span class="material-symbols-outlined text-sm">search</span>
              {/if}
            </button>
          </div>
          {#if searchResults.length > 0}
            <div
              class="absolute z-10 top-full left-0 right-0 mt-1 bg-white border border-slate-200 rounded-lg shadow-lg max-h-48 overflow-y-auto"
            >
              {#each searchResults as place}
                <button
                  class="w-full text-left px-4 py-2 hover:bg-slate-50 border-b border-slate-100 last:border-0"
                  onclick={() => selectPlace(place)}
                >
                  <div class="font-medium text-slate-700 text-sm">
                    {place.name}
                  </div>
                  <div class="text-xs text-slate-500 truncate">
                    {place.address}
                  </div>
                </button>
              {/each}
            </div>
          {/if}
        </div>

        <div class="relative py-2">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-slate-200"></div>
          </div>
          <div class="relative flex justify-center text-sm">
            <span class="px-2 bg-white text-slate-500">Atau isi manual</span>
          </div>
        </div>

        <div class="form-group">
          <label class="label" for="new-name">Nama Kantor</label>
          <input
            id="new-name"
            class="input"
            type="text"
            bind:value={newOffice.name}
            placeholder="Contoh: Kantor Cabang A"
          />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="form-group">
            <label class="label" for="new-lat">Latitude</label>
            <input
              id="new-lat"
              class="input"
              type="text"
              bind:value={newOffice.latitude}
              placeholder="-7.xxxxx"
            />
          </div>
          <div class="form-group">
            <label class="label" for="new-lng">Longitude</label>
            <input
              id="new-lng"
              class="input"
              type="text"
              bind:value={newOffice.longitude}
              placeholder="110.xxxxx"
            />
          </div>
        </div>
        <div class="form-group">
          <label class="label" for="new-address">Alamat Lengkap</label>
          <textarea
            id="new-address"
            class="input min-h-[80px]"
            bind:value={newOffice.address}
            placeholder="Jalan..."
          ></textarea>
        </div>
      </div>
      <div
        class="px-6 py-4 bg-slate-50 border-t border-slate-100 flex justify-end gap-3"
      >
        <button
          class="px-4 py-2 rounded-lg text-slate-600 font-medium hover:bg-slate-200 transition-colors"
          onclick={closeAddModal}>Batal</button
        >
        <button
          class="px-4 py-2 rounded-lg bg-emerald-600 text-white font-medium hover:bg-emerald-700 transition-colors shadow-sm"
          onclick={handleAddOffice}
          disabled={creatingOffice}
        >
          {creatingOffice ? "Menambahkan..." : "Tambah Lokasi"}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Import all styles from both original files */
  :global(body) {
    font-family: "Geist", "Inter", sans-serif;
    color: #0f172a;
    background: #f8fafc;
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

  .card {
    background: rgba(255, 255, 255, 0.85) !important;
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
    overflow: hidden;
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

  /* Stats Tab Styles */
  .approval-card {
    position: relative;
    border: 1px solid rgba(245, 158, 11, 0.5) !important;
    border-left: 4px solid #f59e0b !important;
    background: #fffbeb !important;
    box-shadow: 0 10px 30px -10px rgba(245, 158, 11, 0.2);
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
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.05);
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

  .chart-card:hover {
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

  .header-title {
    display: flex;
    align-items: center;
    gap: 10px;
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
    box-shadow: 0 10px 20px -5px rgba(0, 0, 0, 0.05);
  }

  .task-icon {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    border: 1px solid #f1f5f9;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: all 0.3s;
  }

  .task-icon-emerald {
    border: 1px solid #0ac78b;
  }
  .task-icon-emerald:hover {
    background: #0ac78b;
    color: white;
  }
  .task-icon-purple {
    border: 1px solid #9333ea;
  }
  .task-icon-purple:hover {
    background: #9333ea;
    color: white;
  }

  .avatar-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    color: inherit;
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

  .btn-icon {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.2s;
    background: #fff;
  }

  .btn-icon.approve {
    border-color: #d1fae5;
    color: #10b981;
  }

  .btn-icon.approve:hover {
    background: #10b981;
    border-color: #10b981;
    color: #fff;
    transform: translateY(-2px);
    box-shadow: 0 4px 6px -1px rgba(16, 185, 129, 0.2);
  }

  .btn-icon.deny {
    border-color: #fee2e2;
    color: #ef4444;
  }

  .btn-icon.deny:hover {
    background: #ef4444;
    border-color: #ef4444;
    color: #fff;
    transform: translateY(-2px);
    box-shadow: 0 4px 6px -1px rgba(239, 68, 68, 0.2);
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

  /* Geofence Tab Styles */
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .label {
    font-size: 13px;
    font-weight: 600;
    color: #334155;
  }

  .input {
    width: 100%;
    padding: 10px 14px;
    border: 1px solid #cbd5e1;
    border-radius: 10px;
    font-size: 14px;
    color: #0f172a;
    background: #fff;
    transition: all 0.2s;
  }

  .input:focus {
    outline: none;
    border-color: #10b981;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
  }

  .help-text {
    font-size: 12px;
    color: #94a3b8;
    margin: 0;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
  }

  .preview-box {
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    overflow: hidden;
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

  .slide-up {
    opacity: 0;
    animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  }

  .animate-scale-up {
    opacity: 0;
    transform: scale(0.95);
    animation: scaleUp 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
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

  @keyframes scaleUp {
    from {
      opacity: 0;
      transform: scale(0.95);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }
</style>
