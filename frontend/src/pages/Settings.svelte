<script>
  import { onMount } from "svelte";
  import { api } from "../lib/api.js";

  let loading = $state(true);
  let savingAttendance = $state(false);

  // Office Management State
  let officeList = $state([]);
  let activeOfficeId = $state(null);
  let selectedOfficeId = $state("");
  let loadingOffices = $state(false);
  let creatingOffice = $state(false);
  let deletingOffice = $state(false);
  let settingActive = $state(false);
  let showAddModal = $state(false);

  // Search State
  let searchingPlaces = $state(false);
  let searchResults = $state([]);

  let newOffice = $state({
    name: "",
    latitude: "",
    longitude: "",
    radius_meters: 1000, // Default, though global setting might override
    address: "",
  });

  // Derived state for current selected office details
  let currentOffice = $derived(
    selectedOfficeId && selectedOfficeId !== "new"
      ? officeList.find((o) => o.id == selectedOfficeId)
      : null,
  );

  let attendanceForm = $state({
    attendance_open_time: "07:00",
    check_in_time: "08:00",
    late_tolerance_minutes: 15,
    check_out_time: "17:00",
    office_latitude: "-7.052683",
    office_longitude: "110.469375",
    max_checkin_distance: 1000,
    office_name: "Kantor Pusat",
    workdays: [1, 2, 3, 4, 5, 6],
    manual_off_date: "",
    allow_intern_unscheduled_logging: false,
  });

  const weekdayOptions = [
    { value: 1, label: "S", name: "Senin" },
    { value: 2, label: "S", name: "Selasa" },
    { value: 3, label: "R", name: "Rabu" },
    { value: 4, label: "K", name: "Kamis" },
    { value: 5, label: "J", name: "Jumat" },
    { value: 6, label: "S", name: "Sabtu" },
  ];

  const todayKey = () => {
    // Normalize to WIB (UTC+7) so manual off matches server-side date
    const now = new Date();
    const utcMs = now.getTime() + now.getTimezoneOffset() * 60000;
    const jakarta = new Date(utcMs + 7 * 60 * 60 * 1000);
    const yyyy = jakarta.getFullYear();
    const mm = String(jakarta.getMonth() + 1).padStart(2, "0");
    const dd = String(jakarta.getDate()).padStart(2, "0");
    return `${yyyy}-${mm}-${dd}`;
  };

  onMount(async () => {
    await Promise.all([fetchSettings(), fetchOffices()]);
    loading = false;
  });

  // ... (time helpers: stripSeconds, addMinutesToTime, parseWorkdays, toggleWorkday, isWorkdaySelected, toggleTodayOff, handleTodayOffChange stay same)
  const stripSeconds = (val) => {
    if (!val) return "";
    const parts = val.split(":");
    if (parts.length >= 2)
      return `${parts[0].padStart(2, "0")}:${parts[1].padStart(2, "0")}`;
    return val;
  };

  const addMinutesToTime = (timeStr, minutes) => {
    if (!timeStr) return "";
    const [h, m] = timeStr.split(":").map(Number);
    if (Number.isNaN(h) || Number.isNaN(m)) return timeStr;
    const base = new Date(0, 0, 0, h, m);
    base.setMinutes(base.getMinutes() + Number(minutes || 0));
    const hh = String(base.getHours()).padStart(2, "0");
    const mm = String(base.getMinutes()).padStart(2, "0");
    return `${hh}:${mm}`;
  };

  const parseWorkdays = (val) => {
    if (!val || typeof val !== "string") return [1, 2, 3, 4, 5, 6];
    const days = val
      .split(",")
      .map((n) => Number(n))
      .filter((n) => !Number.isNaN(n) && n >= 0 && n <= 6);
    return days.length ? days : [1, 2, 3, 4, 5, 6];
  };

  function toggleWorkday(day) {
    const current = new Set(attendanceForm.workdays || []);
    if (current.has(day)) current.delete(day);
    else current.add(day);
    const next = Array.from(current).sort((a, b) => a - b);
    attendanceForm = { ...attendanceForm, workdays: next };
  }

  const isWorkdaySelected = (day) =>
    (attendanceForm.workdays || []).includes(day);

  function toggleTodayOff(flag) {
    attendanceForm = {
      ...attendanceForm,
      manual_off_date: flag ? todayKey() : "",
    };
  }

  /** @param {Event & { currentTarget: HTMLInputElement }} event */
  function handleTodayOffChange(event) {
    const target = event.currentTarget;
    if (target && typeof target.checked === "boolean") {
      toggleTodayOff(target.checked);
    }
  }

  async function fetchSettings() {
    try {
      const res = await api.getSettings();
      const list = res?.data || [];
      const map = {};
      list.forEach((item) => {
        if (item?.key) map[item.key] = item.value;
      });

      // Set active office ID
      if (map.active_office_id) {
        activeOfficeId = Number(map.active_office_id);
        if (!selectedOfficeId) selectedOfficeId = activeOfficeId;
      }

      attendanceForm = {
        attendance_open_time:
          stripSeconds(map.attendance_open_time || map.office_start_time) ||
          "07:00",
        check_in_time:
          stripSeconds(map.check_in_time || map.office_start_time) || "08:00",
        late_tolerance_minutes: Number(map.late_tolerance_minutes ?? 15) || 15,
        check_out_time:
          stripSeconds(map.check_out_time || map.office_end_time) || "17:00",
        office_latitude: map.office_latitude || "-7.052683",
        office_longitude: map.office_longitude || "110.469375",
        max_checkin_distance:
          Number(map.max_checkin_distance ?? map.office_radius ?? 1000) || 1000,
        office_name: map.office_name || "Kantor Pusat",
        workdays: parseWorkdays(map.workdays || map.work_days),
        manual_off_date: map.manual_off_date || "",
        allow_intern_unscheduled_logging:
          map.ALLOW_INTERN_UNSCHEDULED_LOGGING === "true",
      };

      // derive tolerance logic... (same as before)
      if (!map.late_tolerance_minutes && map.late_tolerance_time) {
        // ... (existing logic)
        const start = stripSeconds(
          map.check_in_time || map.office_start_time || "08:00",
        );
        const end = stripSeconds(map.late_tolerance_time);
        const diff = (() => {
          const [sh, sm] = (start || "").split(":").map(Number);
          const [eh, em] = (end || "").split(":").map(Number);
          if ([sh, sm, eh, em].some(Number.isNaN)) return null;
          return eh * 60 + em - (sh * 60 + sm);
        })();
        if (diff && diff > 0) attendanceForm.late_tolerance_minutes = diff;
      }
    } catch (err) {
      console.warn("Gagal memuat pengaturan...", err);
    }
  }

  async function fetchOffices() {
    loadingOffices = true;
    try {
      const res = await api.getOffices();
      if (res.success) {
        officeList = res.data || [];
        // If no selected office, select active or first
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
      selectedOfficeId = ""; // Reset dropdown
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
        radius_meters: parseInt(String(newOffice.radius_meters)) || 1000,
      };
      const res = await api.createOffice(payload);
      if (res.success) {
        await fetchOffices();
        selectedOfficeId = res.data.id; // Select the new office
        closeAddModal();
        // If it's the first office, maybe set it active automatically?
        // Better to let user expicitly set active.
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
      // fetchSettings will refresh the legacy fields if needed, but we have activeOfficeId state now.
      // We might want to reload settings to ensure consistency?
      await fetchSettings();
      alert(`Lokasi aktif diubah ke "${currentOffice.name}"`);
    } catch (err) {
      alert("Gagal mengatur lokasi aktif: " + err.message);
    } finally {
      settingActive = false;
    }
  }

  async function saveAttendanceSettings() {
    savingAttendance = true;
    try {
      const t = (val) => (val && val.length === 5 ? `${val}:00` : val || "");
      await api.updateSettings({
        attendance_open_time: t(attendanceForm.attendance_open_time),
        check_in_time: t(attendanceForm.check_in_time),
        check_out_time: t(attendanceForm.check_out_time),
        late_tolerance_minutes: String(attendanceForm.late_tolerance_minutes),
        // office_latitude: attendanceForm.office_latitude, // NO LONGER USED from here, comes from office_locations
        // office_longitude: attendanceForm.office_longitude, // NO LONGER USED from here
        max_checkin_distance: String(attendanceForm.max_checkin_distance), // This is GLOBAL now
        // office_name: attendanceForm.office_name, // NO LONGER USED from here
        workdays: (attendanceForm.workdays || []).join(","),
        manual_off_date: attendanceForm.manual_off_date || "",
        ALLOW_INTERN_UNSCHEDULED_LOGGING: String(
          attendanceForm.allow_intern_unscheduled_logging,
        ),
        // legacy/compat keys
        office_start_time: t(attendanceForm.check_in_time),
        office_end_time: t(attendanceForm.check_out_time),
        // office_radius: String(attendanceForm.max_checkin_distance), // GLOBAL
        late_tolerance_time: t(
          addMinutesToTime(
            attendanceForm.check_in_time,
            attendanceForm.late_tolerance_minutes,
          ),
        ),
      });
      alert("Pengaturan presensi tersimpan.");
      await fetchSettings();
    } catch (err) {
      alert("Gagal menyimpan pengaturan: " + (err.message || "unknown error"));
    } finally {
      savingAttendance = false;
    }
  }

  function handleKeydown(e) {
    if (e.key === "Escape" && showAddModal) closeAddModal();
  }
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    <div class="header">
      <div>
        <h2>Pengaturan Jam & Lokasi</h2>
      </div>

      <button
        class="btn-blue flex items-center justify-center gap-2 w-full md:w-auto"
        onclick={saveAttendanceSettings}
        disabled={savingAttendance}
      >
        {#if savingAttendance}
          <span class="material-symbols-outlined animate-spin text-sm"
            >refresh</span
          >
          <span>Menyimpan...</span>
        {:else}
          <span class="material-symbols-outlined text-sm">save</span>
          <span>Simpan Pengaturan</span>
        {/if}
      </button>
    </div>

    {#if loading}
      <div class="loading">Memuat pengaturan...</div>
    {:else}
      <div class="two-col">
        <div class="content-card animate-slide-up">
          <div class="card-header">
            <h3>Jam Kerja</h3>
            <p>Jam buka presensi, jam masuk, toleransi, dan jam pulang.</p>
          </div>
          <div class="card-body space-y-4">
            <div class="form-grid">
              <div class="form-group">
                <label class="label" for="attendanceOpen">Presensi Dibuka</label
                >
                <input
                  id="attendanceOpen"
                  class="input"
                  type="time"
                  bind:value={attendanceForm.attendance_open_time}
                />
                <p class="help-text">Jam mulai sistem menerima check-in.</p>
              </div>
              <div class="form-group">
                <label class="label" for="checkInTime">Jam Masuk</label>
                <input
                  id="checkInTime"
                  class="input"
                  type="time"
                  bind:value={attendanceForm.check_in_time}
                />
                <p class="help-text">Waktu mulai absen masuk.</p>
              </div>
            </div>
            <div class="form-grid">
              <div class="form-group">
                <label class="label" for="lateTolerance"
                  >Toleransi Terlambat (menit)</label
                >
                <input
                  id="lateTolerance"
                  class="input"
                  type="number"
                  min="0"
                  bind:value={attendanceForm.late_tolerance_minutes}
                />
                <p class="help-text">
                  Menit tambahan sebelum presensi ditutup.
                </p>
              </div>
              <div class="form-group">
                <label class="label" for="checkOutTime">Jam Pulang</label>
                <input
                  id="checkOutTime"
                  class="input"
                  type="time"
                  bind:value={attendanceForm.check_out_time}
                />
                <p class="help-text">Waktu minimal absen pulang.</p>
              </div>
            </div>
            <div class="preview-box">
              <div class="preview-grid">
                <div class="preview-item">
                  <div class="preview-title success">Tepat Waktu</div>
                  <div class="preview-sub">
                    Sebelum {attendanceForm.check_in_time || "08:00"}
                  </div>
                </div>

                <div class="preview-item">
                  <div class="preview-title warning">Terlambat</div>
                  <div class="preview-sub">
                    Ditutup pukul {addMinutesToTime(
                      attendanceForm.check_in_time,
                      attendanceForm.late_tolerance_minutes,
                    )}
                  </div>
                </div>

                <div class="preview-item">
                  <div class="preview-title success">Boleh Pulang</div>
                  <div class="preview-sub">
                    Mulai {attendanceForm.check_out_time || "17:00"}
                  </div>
                </div>
              </div>
            </div>

            <!-- Tindakan Ekstra Card -->
            <div class="settings-section-card">
              <div class="pb-1 border-b border-slate-100 mb-4">
                <p
                  class="label text-slate-400 uppercase tracking-widest text-[10px]"
                >
                  Kontrol Operasional
                </p>
              </div>

              <!-- Hari Kerja -->
              <div class="form-group pb-6 mb-6 border-b border-slate-100">
                <p class="label">Hari Kerja</p>
                <div class="day-pills">
                  {#each weekdayOptions as day}
                    <button
                      type="button"
                      class="day-pill"
                      class:active={isWorkdaySelected(day.value)}
                      title={day.name}
                      aria-label={day.name}
                      onclick={() => toggleWorkday(day.value)}
                    >
                      {day.label}
                    </button>
                  {/each}
                </div>
                <p class="help-text pt-0.5">
                  Pilih hari aktif. Minggu tidak ditampilkan.
                </p>
              </div>

              <!-- Day Off -->
              <div class="form-group pb-6 mb-6 border-b border-slate-100">
                <p class="label">Liburkan Hari Ini</p>
                <div class="today-off-toggle">
                  <label class="checkbox-wrapper">
                    <input
                      id="todayOff"
                      type="checkbox"
                      checked={attendanceForm.manual_off_date === todayKey()}
                      onchange={handleTodayOffChange}
                    />
                    <span class="checkmark"></span>
                  </label>
                  <div>
                    <div class="toggle-title">Tidak ada jadwal kantor</div>
                    <p class="help-text">
                      Centang untuk meliburkan hari ini. Akan reset sehari
                      kemudian.
                    </p>
                  </div>
                </div>
              </div>

              <!-- Aturan Penugasan -->
              <div class="form-group">
                <p class="label">Aturan Penugasan</p>
                <div class="today-off-toggle">
                  <label class="checkbox-wrapper">
                    <input
                      type="checkbox"
                      bind:checked={
                        attendanceForm.allow_intern_unscheduled_logging
                      }
                    />
                    <span class="checkmark"></span>
                  </label>
                  <div>
                    <div class="toggle-title">Self-Reporting Tugas</div>
                    <p class="help-text">
                      Intern mencatat tugas secara mandiri sesuai arahan
                      penugas.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="content-card animate-slide-up">
          <div class="card-header">
            <h3>Lokasi Absensi</h3>
            <p>Kelola lokasi kantor yang tersedia dan pilih lokasi aktif.</p>
          </div>
          <div class="card-body space-y-4">
            <!-- Global Radius Setting -->
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
                  bind:value={attendanceForm.max_checkin_distance}
                />
                <p class="help-text">
                  Jarak maksimal dari titik kantor untuk bisa absen (Global).
                </p>
              </div>
            </div>

            <!-- Location Selector & Add Button -->
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

            <!-- Active Location Indicator -->
            {#if selectedOfficeId === activeOfficeId && activeOfficeId}
              <div
                class="bg-emerald-50 border border-emerald-200 rounded-lg p-3 flex items-center gap-3 mb-4"
              >
                <span class="material-symbols-outlined text-emerald-600"
                  >check_circle</span
                >
                <p class="text-sm text-emerald-700 font-medium">
                  Lokasi ini sedang aktif digunakan untuk presensi.
                </p>
              </div>
            {/if}

            <!-- Location Preview Grid -->
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

              <!-- Actions for Selected Location -->
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
                    {settingActive
                      ? "Mengatur..."
                      : "Atur Sebagai Lokasi Aktif"}
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

            <div class="info-box mt-4">
              <div class="icon-circle bg-emerald-light text-emerald">
                <span class="material-symbols-outlined">satellite_alt</span>
              </div>
              <div>
                <h4 class="info-title">Tips</h4>
                <p class="info-desc">
                  Pastikan koordinat akurat. Gunakan Google Maps untuk menyalin
                  lat/lng.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
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
          <!-- Search Box -->
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
</div>

<style>
  :global(body) {
    font-family: "Geist", "Inter", sans-serif;
    color: #0f172a;
    background: #f8fafc;
  }
  .page-bg {
    min-height: 100vh;
    padding: 40x 24px;
    background: radial-gradient(
      at 0% 0%,
      rgba(16, 185, 129, 0.03) 0%,
      transparent 50%
    );
  }
  .container {
    max-width: 1240px;
    margin: 0 auto;
  }
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
    margin-bottom: 20px;
  }
  .header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 760;
  }

  .two-col {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 20px;
  }

  .content-card {
    background: white;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
    overflow: hidden;
  }
  .card-header {
    padding: 20px 24px;
    border-bottom: 1px solid #f1f5f9;
  }
  .card-header h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 700;
    color: #1e293b;
  }
  .card-header p {
    margin: 4px 0 0;
    font-size: 14px;
    color: #64748b;
  }
  .card-body {
    padding: 20px 24px;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
  }
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

  .workdays-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 16px;
    align-items: center;
  }
  .day-pills {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
  .day-pill {
    min-width: 36px;
    padding: 10px 12px;
    border-radius: 8px;
    background: #fff;
    border: 1px solid #10b981;
    font-weight: 700;
    color: #0f172a;
    cursor: pointer;
    transition: all 0.2s ease;
    /* box-shadow: 0 2px 4px rgba(15, 23, 42, 0.04); */
  }
  .day-pill:hover {
    transform: translateY(-1px);
  }
  .day-pill.active {
    background: #10b981;
    border-color: #10b981;
    color: #fff;
    /* box-shadow: 0 6px 16px rgba(16, 185, 129, 0.25); */
  }

  .settings-section-card {
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 16px;
    background: #f8fafc;
    margin-top: 16px;
  }

  .today-off-toggle {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .checkbox-wrapper {
    position: relative;
    display: inline-block;
    cursor: pointer;
  }

  .checkbox-wrapper input {
    position: absolute;
    opacity: 0;
    cursor: pointer;
    height: 0;
    width: 0;
  }

  .checkmark {
    display: inline-block;
    width: 20px;
    height: 20px;
    border: 2px solid #cbd5e1;
    border-radius: 4px;
    background: #fff;
    transition: all 0.2s;
  }

  .checkbox-wrapper:hover .checkmark {
    border-color: #10b981;
  }

  .checkbox-wrapper input:checked ~ .checkmark {
    background: #10b981;
    border-color: #10b981;
  }

  .checkbox-wrapper input:checked ~ .checkmark:after {
    content: "";
    position: absolute;
    display: block;
    left: 6px;
    top: 2px;
    width: 5px;
    height: 10px;
    border: solid white;
    border-width: 0 2px 2px 0;
    transform: rotate(45deg);
  }

  .toggle-title {
    font-weight: 700;
    color: #0f172a;
    margin: 0;
  }

  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669);
    color: white;
    border: none;
    padding: 10px 18px;
    border-radius: 999px;
    font-weight: 600;
    font-size: 14px;
    cursor: pointer;
    box-shadow: 0 4px 6px rgba(16, 185, 129, 0.2);
    transition: 0.2s;
  }

  .btn-blue {
    background: #4f46e5;
    color: white;
    border: none;
    padding: 10px 18px;
    border-radius: 999px;
    font-weight: 600;
    font-size: 14px;
    cursor: pointer;
    transition: 0.2s;
  }

  .btn-primary:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 6px 12px rgba(16, 185, 129, 0.3);
  }
  .btn-primary:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .btn-blue:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 6px 12px rgba(79, 70, 229, 0.2);
  }
  .btn-blue:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  @media (max-width: 768px) {
    .btn-primary {
      width: 100%;
    }
  }

  .preview-box {
    /* border: 1px solid #e2e8f0; */
    /* border-radius: 16px; */
    /* padding: 16px; */
    background: #f8fafc;
  }

  .preview-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
  }

  @media (max-width: 768px) {
    .preview-grid {
      grid-template-columns: 1fr;
    }
  }

  .preview-item {
    background: white;
    border-radius: 12px;
    padding: 14px;
    border: 1px solid #e2e8f0;
    text-align: center;
  }
  .badge-heading {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .preview-row {
    display: flex;
    align-items: center;
    gap: 14px;
    flex-wrap: wrap;
  }

  .preview-sub {
    font-size: 13px;
    color: #64748b;
    text-align: center;
  }

  .preview-title {
    font-size: 15px;
    margin-bottom: 4px;
    text-transform: uppercase;
    font-weight: 800;
    letter-spacing: 0.04em;
  }

  .preview-title.success {
    color: #059669;
  }
  .preview-title.warning {
    color: #d97706;
  }

  .divider-vertical {
    width: 2px;
    height: 32px;
    background: #e2e8f0;
  }
  .text-success {
    color: #059669;
  }
  .text-warning {
    color: #d97706;
  }

  .info-box {
    display: flex;
    gap: 12px;
    align-items: flex-start;
    padding: 14px;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    background: #f8fafc;
  }
  .icon-circle {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .bg-emerald-light {
    background: #ecfdf5;
  }
  .text-emerald {
    color: #059669;
  }
  .info-title {
    margin: 0;
    font-size: 14px;
    font-weight: 700;
    color: #0f172a;
  }
  .info-desc {
    margin: 4px 0 0;
    font-size: 13px;
    color: #64748b;
  }

  .loading {
    padding: 20px;
    color: #64748b;
  }

  .info-desc {
    margin: 4px 0 0;
    font-size: 13px;
    color: #64748b;
  }

  /* PREVIEW */
  .badge-heading {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .preview-row {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
  }
  .divider-vertical {
    width: 2px;
    height: 40px;
    background: #e2e8f0;
  }

  /* UTILS */
  .mt-4 {
    margin-top: 16px;
  }
  .icon-circle {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .bg-emerald-light {
    background: #ecfdf5;
  }
  .text-emerald {
    color: #059669;
  }

  .animate-fade-in {
    opacity: 0;
    animation: fadeIn 0.6s forwards;
  }
  .animate-slide-up {
    opacity: 0;
    animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  }
  @keyframes fadeIn {
    to {
      opacity: 1;
    }
  }
  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
