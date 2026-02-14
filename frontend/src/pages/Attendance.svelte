<script>
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import { api } from "../lib/api.js";
  import { portal } from "../lib/portal.js";
  import { auth } from "../lib/auth.svelte.js";
  import { getAvatarUrl } from "../lib/utils.js";

  // State
  let records = $state([]);
  let loading = $state(false);
  let searchQuery = $state("");
  let filterStatus = $state("");
  let filterDate = $state("");
  let currentPage = $state(1);
  let totalPages = $state(1);
  let totalItems = $state(0);
  let searchTimeout;
  let expandedRecords = $state({});

  // Detail Modal State
  let detailOpen = $state(false);
  let detailLoading = $state(false);
  let detail = $state(null);

  // Document Modal State
  let docOpen = $state(false);
  let docUrl = $state("");

  // Keep overlay-root click-through state in sync with our modals
  $effect(() => {
    const root =
      typeof document !== "undefined"
        ? document.querySelector("#overlay-root")
        : null;
    if (!(root instanceof HTMLElement)) return;
    const hasModal = detailOpen || docOpen;
    root.style.pointerEvents = hasModal ? "auto" : "none";
    if (!hasModal) {
      root.dataset.portalCount = "0";
    }
  });

  // Labels & Colors
  const statusLabels = {
    present: "Hadir",
    late: "Terlambat",
    absent: "Tidak Hadir",
    sick: "Sakit",
    permission: "Izin",
  };

  function getStatusColor(status) {
    switch (status) {
      case "present":
        return "bg-emerald-100 text-emerald-700 border-emerald-200";
      case "late":
        return "bg-yellow-100 text-yellow-700 border-yellow-200";
      case "sick":
        return "bg-blue-100 text-blue-700 border-blue-200";
      case "permission":
        return "bg-purple-100 text-purple-700 border-purple-200";
      case "absent":
        return "bg-red-100 text-red-600 border-red-200";
      default:
        return "bg-slate-100 text-slate-600 border-slate-200";
    }
  }

  function formatDate(value) {
    if (!value) return "—";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    });
  }

  function formatTime(value) {
    if (!value) return "--:--";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      if (typeof value === "string" && value.includes(":"))
        return value.slice(0, 5);
      return "--:--";
    }
    return date.toLocaleTimeString("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function withBase(url) {
    if (!url) return "";
    const needsToken = url.startsWith("/uploads") || url.startsWith("uploads/");
    const token = auth.token;
    const tokenQS =
      needsToken && token
        ? `${url.includes("?") ? "&" : "?"}token=${token}`
        : "";
    if (url.startsWith("http")) return `${url}${tokenQS}`;
    const base = import.meta.env.VITE_API_URL || "";
    const normalized = url.startsWith("/") ? url : `/${url}`;
    return `${base}${normalized}${tokenQS}`;
  }

  function isImageFile(url) {
    if (!url) return false;
    return /\.(jpg|jpeg|png|webp|gif|svg)($|\?)/i.test(url);
  }

  // --- Fetch Data ---
  async function fetchAttendance() {
    loading = true;
    try {
      const params = { page: currentPage, limit: 50 };
      if (searchQuery) params.search = searchQuery;
      if (filterStatus) params.status = filterStatus;
      if (filterDate) params.date = filterDate;

      const res = await api.getAttendance(params);
      records = res.data || [];
      const pagination = res.pagination || {};
      totalPages = Math.max(pagination.total_pages || 1, 1);
      totalItems = pagination.total_items || 0;
      currentPage = pagination.page || currentPage;
      console.log("Fetched attendance:", records);
    } catch (err) {
      console.error("Failed to fetch attendance:", err);
      alert("Gagal memuat data presensi: " + err.message);
    } finally {
      loading = false;
    }
  }

  function goToPreviousPage() {
    currentPage = Math.max(1, currentPage - 1);
    fetchAttendance();
  }

  function goToNextPage() {
    if (currentPage >= totalPages) return;
    currentPage += 1;
    fetchAttendance();
  }

  function handleSearchInput() {
    clearTimeout(searchTimeout);
    currentPage = 1;
    searchTimeout = setTimeout(() => {
      fetchAttendance();
    }, 500);
  }

  async function openDetail(id) {
    detailOpen = true;
    detailLoading = true;
    detail = null;
    try {
      const res = await api.getAttendanceById(id);
      detail = res.data;
    } catch (err) {
      console.error(err);
      alert("Gagal memuat detail: " + err.message);
    } finally {
      detailLoading = false;
    }
  }

  function closeDetail() {
    detailOpen = false;
    detail = null;
  }

  function openDoc(url) {
    const full = withBase(url);
    docUrl = full;
    docOpen = true;
  }

  function closeDoc() {
    docOpen = false;
    docUrl = "";
  }

  async function handleDelete(id) {
    if (!confirm("Apakah Anda yakin ingin menghapus data presensi ini?"))
      return;
    try {
      await api.deleteAttendance(id);
      records = records.filter((r) => r.id !== id);
      alert("Data presensi berhasil dihapus.");
    } catch (err) {
      console.error(err);
      alert("Gagal menghapus data: " + err.message);
    }
  }

  function toggleExpand(id) {
    expandedRecords[id] = !expandedRecords[id];
  }

  onMount(async () => {
    await fetchAttendance();
  });
</script>

<div class="page-container animate-fade-in">
  <div class="flex items-center gap-3 pb-8">
    <h4 class="card-title">Presensi & Kehadiran</h4>
    <span class="badge-count">{records.length} dari {totalItems} Data</span>
  </div>

  <!-- BAGIAN TABEL DAFTAR -->
  <div class="card table-card animate-slide-up" style="animation-delay: 0.1s;">
    <div class="card-header-row border-b">
      <div class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2">
        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2"
          onclick={() => {
            searchQuery = "";
            filterStatus = "";
            filterDate = "";
            currentPage = 1;
            fetchAttendance();
          }}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"
            ></polygon></svg
          >
          <span>Clear</span>
        </button>
        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2"
          onclick={fetchAttendance}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><path
              d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.3"
            /></svg
          >
          <span>Refresh</span>
        </button>
      </div>
      <div class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2">
        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2 {totalPages <=
          1
            ? 'opacity-40 cursor-not-allowed'
            : ''}"
          onclick={goToPreviousPage}
          disabled={currentPage <= 1 || totalPages <= 1}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"><path d="M15 18l-6-6 6-6" /></svg
          >
          <span>Prev</span>
        </button>

        <div
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2 pagination-pill {totalPages <=
          1
            ? 'opacity-40'
            : ''}"
        >
          <span>{currentPage}</span>
          <span class="text-slate-500">of</span>
          <span>{totalPages}</span>
        </div>

        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2 {totalPages <=
          1
            ? 'opacity-40 cursor-not-allowed'
            : ''}"
          onclick={goToNextPage}
          disabled={currentPage >= totalPages || totalPages <= 1}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"><path d="M9 18l6-6-6-6" /></svg
          >
          <span>Next</span>
        </button>
      </div>
    </div>

    <div class="toolbar">
      <div class="search-wrapper">
        <span class="material-symbols-outlined search-icon">search</span>
        <input
          type="text"
          bind:value={searchQuery}
          oninput={handleSearchInput}
          onkeydown={(e) =>
            e.key === "Enter" &&
            (clearTimeout(searchTimeout), fetchAttendance())}
          placeholder="Cari Nama Intern..."
          class="search-input"
        />
      </div>

      <input
        type="date"
        bind:value={filterDate}
        onchange={fetchAttendance}
        class="filter-select"
      />

      <select
        bind:value={filterStatus}
        onchange={fetchAttendance}
        class="filter-select"
      >
        <option value="">Semua Status</option>
        <option value="present">Hadir</option>
        <option value="late">Terlambat</option>
        <option value="permission">Izin</option>
        <option value="sick">Sakit</option>
        <option value="absent">Tidak Hadir</option>
      </select>
    </div>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Memuat data...</p>
      </div>
    {:else if records.length === 0}
      <div class="empty-state">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="48"
          height="48"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#e5e7eb"
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
        <p>Belum ada riwayat presensi.</p>
      </div>
    {:else}
      <div class="table-responsive desktop-only">
        <table class="table">
          <thead>
            <tr>
              <th>Tanggal & Intern</th>
              <th>Waktu Masuk</th>
              <th>Waktu Keluar</th>
              <th class="text-center">Status</th>
              <th class="text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {#each records as r}
              <tr class="table-row">
                <td style="min-width: 220px;">
                  <div class="attendance-info">
                    <div class="attendance-icon-wrapper">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"
                        ></rect>
                        <line x1="16" y1="2" x2="16" y2="6"></line>
                        <line x1="8" y1="2" x2="8" y2="6"></line>
                        <line x1="3" y1="10" x2="21" y2="10"></line>
                      </svg>
                    </div>
                    <div class="attendance-details">
                      <span class="attendance-date">{formatDate(r.date)}</span>
                      {#if auth.user?.role !== "intern"}
                        <div class="user-info-inline">
                          {#if r.intern_avatar && getAvatarUrl(r.intern_avatar)}
                            <img
                              src={getAvatarUrl(r.intern_avatar)}
                              alt={r.intern_name}
                              class="w-6 h-6 rounded-full object-cover"
                            />
                          {:else}
                            <div class="avatar-small">
                              {r.intern_name?.charAt(0) || "U"}
                            </div>
                          {/if}
                          <span class="intern-name-small"
                            >{r.intern_name || "-"}</span
                          >
                        </div>
                      {/if}
                    </div>
                  </div>
                </td>
                <td class="text-muted mono">{formatTime(r.check_in_time)}</td>
                <td class="text-muted mono">{formatTime(r.check_out_time)}</td>
                <td class="text-center">
                  <span
                    class={`status-badge equal-badge ${getStatusColor(r.status)}`}
                  >
                    {statusLabels[r.status] || r.status || "-"}
                  </span>
                </td>
                <td class="text-right">
                  <button
                    class="btn-icon text-sky-600 hover:text-sky-700 bg-sky-50 hover:bg-sky-100"
                    onclick={() => openDetail(r.id)}
                    title="Detail presensi"
                  >
                    <svg
                      width="18"
                      height="18"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                      ></path>
                      <circle cx="12" cy="12" r="3"></circle>
                    </svg>
                  </button>
                  <button
                    class="btn-icon text-slate-600 hover:text-slate-700 bg-slate-50 hover:bg-slate-100 {r.proof_file
                      ? ''
                      : 'opacity-40 cursor-not-allowed'}"
                    onclick={() => r.proof_file && openDoc(r.proof_file)}
                    title="Bukti izin"
                    disabled={!r.proof_file}
                  >
                    <svg
                      width="18"
                      height="18"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path
                        d="m21.44 11.05-9.19 9.19a6 6 0 0 1-8.49-8.49l8.57-8.57A4 4 0 1 1 18 8.84l-8.59 8.57a2 2 0 0 1-2.83-2.83l8.49-8.48"
                      ></path>
                    </svg>
                  </button>
                  <button
                    class="btn-icon text-slate-600 hover:text-slate-700 bg-slate-50 hover:bg-slate-100"
                    onclick={() => handleDelete(r.id)}
                    title="Hapus"
                  >
                    <svg
                      width="18"
                      height="18"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <polyline points="3 6 5 6 21 6"></polyline>
                      <path
                        d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                      ></path>
                    </svg>
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="mobile-list">
        {#each records as r}
          <div class="entry-card">
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="entry-head" onclick={() => toggleExpand(r.id)}>
              <div class="date-row">
                <span class="material-symbols-outlined date-icon"
                  >calendar_month</span
                >
                <span class="date-text">{formatDate(r.date)}</span>
              </div>
              <div class="head-right">
                <span
                  class={`status-badge equal-badge ${getStatusColor(r.status)}`}
                >
                  {statusLabels[r.status] || r.status || "-"}
                </span>
                <button class="expand-btn">
                  <span
                    class="material-symbols-outlined transition-transform duration-200 {expandedRecords[
                      r.id
                    ]
                      ? 'rotate-180'
                      : ''}">expand_more</span
                  >
                </button>
              </div>
            </div>

            {#if expandedRecords[r.id]}
              <div class="entry-details" transition:slide={{ duration: 200 }}>
                {#if auth.user?.role !== "intern"}
                  <div class="intern-grid">
                    <div class="intern-box">
                      {#if r.intern_avatar && getAvatarUrl(r.intern_avatar)}
                        <img
                          src={getAvatarUrl(r.intern_avatar)}
                          alt={r.intern_name}
                          class="w-6 h-6 rounded-full object-cover"
                        />
                      {:else}
                        <div class="avatar-mini">
                          {r.intern_name?.charAt(0) || "U"}
                        </div>
                      {/if}
                      <span class="intern-box-label"
                        >{r.intern_name || "-"}</span
                      >
                    </div>
                  </div>
                {/if}
                <div class="time-grid">
                  <div class="time-box">
                    <p class="label">Presensi Masuk</p>
                    <p class="time-value">{formatTime(r.check_in_time)}</p>
                  </div>
                  <div class="time-box">
                    <p class="label">Presensi Keluar</p>
                    <p class="time-value">{formatTime(r.check_out_time)}</p>
                  </div>
                </div>
                <div class="mobile-actions">
                  <button
                    class="mini-btn mobile"
                    onclick={(e) => {
                      e.stopPropagation();
                      openDetail(r.id);
                    }}
                  >
                    <span class="material-symbols-outlined">visibility</span>
                    <span class="btn-text">Detail</span>
                  </button>
                  <button
                    class="mini-btn-circle mobile {r.proof_file
                      ? ''
                      : 'disabled'}"
                    onclick={(e) => {
                      e.stopPropagation();
                      r.proof_file && openDoc(r.proof_file);
                    }}
                    disabled={!r.proof_file}
                  >
                    <span class="material-symbols-outlined">attach_file</span>
                  </button>
                  <button
                    class="mini-btn mobile danger"
                    onclick={(e) => {
                      e.stopPropagation();
                      handleDelete(r.id);
                    }}
                    aria-label="Hapus presensi"
                  >
                    <span class="material-symbols-outlined">delete</span>
                    <span class="btn-text">Hapus</span>
                  </button>
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Detail Modal -->
{#if detailOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6"
    use:portal
    onclick={closeDetail}
  >
    <div
      class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity"
    ></div>

    <div
      class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg max-h-[90vh] flex flex-col overflow-hidden"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Header -->
      <div
        class="p-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/50"
      >
        <div>
          <h3 class="text-lg font-bold text-slate-800">Detail Presensi</h3>
          {#if detail}
            <p class="text-slate-500 text-sm mt-0.5">
              {detail.intern_name || auth.user?.name || "Intern"}
            </p>
          {/if}
        </div>
        <button
          onclick={closeDetail}
          class="w-8 h-8 rounded-lg bg-slate-100 text-slate-500 flex items-center justify-center hover:bg-rose-50 hover:text-rose-500 transition-colors"
        >
          <span class="material-symbols-outlined text-[20px]">close</span>
        </button>
      </div>

      <!-- Body -->
      <div class="p-6 overflow-y-auto">
        {#if detailLoading}
          <div
            class="flex flex-col items-center justify-center py-8 text-slate-500"
          >
            <div
              class="w-8 h-8 border-2 border-slate-200 border-t-indigo-500 rounded-full animate-spin mb-2"
            ></div>
            <p class="text-sm">Memuat detail...</p>
          </div>
        {:else if detail}
          <div class="space-y-6">
            <!-- Status Banner -->
            <div
              class="flex items-center justify-between p-4 rounded-xl {getStatusColor(
                detail.status,
              )} border"
            >
              <div class="flex flex-col">
                <span
                  class="text-xs font-bold uppercase tracking-wider opacity-70"
                  >Status</span
                >
                <span class="font-bold text-lg"
                  >{statusLabels[detail.status] || detail.status}</span
                >
              </div>
              <div class="text-right">
                <span
                  class="text-xs font-bold uppercase tracking-wider opacity-70"
                  >Tanggal</span
                >
                <div class="font-bold">
                  {formatDate(detail.date || detail.check_in_time)}
                </div>
              </div>
            </div>

            <!-- Time Grid -->
            <div class="grid grid-cols-2 gap-4">
              <div
                class="p-4 bg-slate-50 rounded-xl border border-slate-100 text-center"
              >
                <div
                  class="text-xs font-bold text-slate-400 uppercase tracking-wider mb-1"
                >
                  Masuk
                </div>
                <div class="text-xl font-mono font-bold text-slate-700">
                  {formatTime(detail.check_in_time)}
                </div>
              </div>
              <div
                class="p-4 bg-slate-50 rounded-xl border border-slate-100 text-center"
              >
                <div
                  class="text-xs font-bold text-slate-400 uppercase tracking-wider mb-1"
                >
                  Keluar
                </div>
                <div class="text-xl font-mono font-bold text-slate-700">
                  {formatTime(detail.check_out_time)}
                </div>
              </div>
            </div>

            <!-- Details -->
            <div class="space-y-4">
              {#if detail.late_reason}
                <div>
                  <h4 class="text-sm font-bold text-slate-700 mb-1">
                    Alasan Terlambat
                  </h4>
                  <p
                    class="text-sm text-slate-600 bg-amber-50 p-3 rounded-lg border border-amber-100"
                  >
                    {detail.late_reason}
                  </p>
                </div>
              {/if}

              {#if detail.notes}
                <div>
                  <h4 class="text-sm font-bold text-slate-700 mb-1">
                    Catatan / Izin
                  </h4>
                  <p
                    class="text-sm text-slate-600 bg-slate-50 p-3 rounded-lg border border-slate-100"
                  >
                    {detail.notes}
                  </p>
                </div>
              {/if}

              {#if detail.proof_file}
                <div>
                  <h4 class="text-sm font-bold text-slate-700 mb-2">
                    Bukti Lampiran
                  </h4>
                  <button
                    onclick={() => openDoc(detail.proof_file)}
                    class="flex items-center gap-3 p-3 rounded-lg border border-slate-200 hover:border-indigo-300 hover:bg-indigo-50 transition-all group w-full"
                  >
                    <div
                      class="w-10 h-10 rounded-lg bg-indigo-100 text-indigo-600 flex items-center justify-center"
                    >
                      <span class="material-symbols-outlined">description</span>
                    </div>
                    <div class="flex-1 text-left">
                      <p
                        class="text-sm font-bold text-slate-700 group-hover:text-indigo-700"
                      >
                        Lihat Dokumen
                      </p>
                      <p class="text-xs text-slate-500">Klik untuk membuka</p>
                    </div>
                    <span
                      class="material-symbols-outlined text-slate-400 group-hover:text-indigo-500"
                      >open_in_new</span
                    >
                  </button>
                </div>
              {/if}
            </div>
          </div>
        {:else}
          <div class="text-center py-8 text-slate-400">
            <p>Detail tidak ditemukan.</p>
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="p-4 border-t border-slate-100 bg-slate-50/50 text-right">
        <button
          onclick={closeDetail}
          class="px-4 py-2 bg-white border border-slate-200 text-slate-600 rounded-lg hover:bg-slate-50 font-medium text-sm transition-colors shadow-sm"
        >
          Tutup
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Document Modal -->
{#if docOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6"
    use:portal
    onclick={closeDoc}
  >
    <div
      class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity"
    ></div>

    <div
      class="relative bg-white rounded-2xl shadow-xl flex flex-col overflow-hidden transition-all duration-200"
      class:w-full={!isImageFile(docUrl)}
      class:max-w-4xl={!isImageFile(docUrl)}
      class:h-[85vh]={!isImageFile(docUrl)}
      class:w-auto={isImageFile(docUrl)}
      class:max-w-[90vw]={isImageFile(docUrl)}
      class:max-h-[90vh]={isImageFile(docUrl)}
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Header -->
      <div
        class="p-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50 shrink-0 min-w-[320px]"
      >
        <div>
          <h3 class="text-lg font-bold text-slate-800">Lampiran Dokumen</h3>
        </div>
        <div class="flex items-center gap-2">
          {#if docUrl}
            <a
              href={docUrl}
              target="_blank"
              rel="noreferrer"
              class="px-3 py-1.5 bg-white border border-slate-200 text-slate-600 rounded-lg hover:bg-slate-50 font-medium text-xs transition-colors shadow-sm flex items-center gap-1"
            >
              <span class="material-symbols-outlined text-[16px]"
                >open_in_new</span
              >
              Fullscreen
            </a>
          {/if}
          <button
            onclick={closeDoc}
            class="w-8 h-8 rounded-lg bg-slate-100 text-slate-500 flex items-center justify-center hover:bg-rose-50 hover:text-rose-500 transition-colors"
          >
            <span class="material-symbols-outlined text-[20px]">close</span>
          </button>
        </div>
      </div>

      <!-- Body -->
      <div
        class="flex-1 bg-slate-100 relative overflow-hidden flex items-center justify-center"
      >
        {#if docUrl}
          {#if isImageFile(docUrl)}
            <img
              src={docUrl}
              alt="Lampiran"
              class="block max-w-full max-h-[calc(90vh-80px)] object-contain"
            />
          {:else}
            <iframe title="Lampiran" src={docUrl} class="w-full h-full border-0"
            ></iframe>
          {/if}
        {:else}
          <div
            class="absolute inset-0 flex items-center justify-center text-slate-400"
          >
            <p>Lampiran tidak tersedia.</p>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  /* Base Layout */

  .page-container {
    animation: fadeIn 0.5s ease-out;
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
    padding: 0 16px;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  /* Card Styles */
  .card {
    background: white;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
    overflow: hidden;
  }

  .table-card {
    padding: 0;
  }

  .text-muted {
    color: #6b7280;
    font-size: 0.875rem;
    margin: 4px 0 0 0;
  }

  .small {
    font-size: 0.75rem;
  }

  /* Table Header Row */
  .card-header-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #f3f4f6;
    background: rgba(248, 250, 252, 0.5);
  }

  .card-title {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #111827;
    display: inline-block;
  }
  @media (max-width: 900px) {
    .card-header-row {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;
    }
    .toolbar {
      padding: 14px 16px;
    }
    .search-wrapper {
      flex: 1 1 100%;
    }
    .filter-select {
      width: 100%;
    }
  }

  .badge-count {
    background: #f1f5f9;
    color: #64748b;
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 14px;
    font-weight: 600;
  }

  /* Filters */
  .toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    padding: 16px 24px;
    border-bottom: 1px solid #f3f4f6;
    background: #fafbfd;
  }
  .search-wrapper {
    flex: 1 1 320px;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 14px;
    border: 1px solid #e5e7eb;
    border-radius: 10px;
    background: #fff;
    box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  }
  .search-icon {
    color: #9ca3af;
    font-variation-settings: "wght" 550;
  }
  .search-input {
    flex: 1;
    border: none;
    outline: none;
    font-size: 0.95rem;
    background: transparent;
    color: #111827;
  }
  .search-input::placeholder {
    color: #9ca3af;
  }

  .filter-select {
    min-width: 180px;
    border-radius: 10px;
    border: 1px solid #e5e7eb;
    padding: 10px 12px;
    background: #fff;
    font-weight: 600;
    color: #334155;
    box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  }

  /* Table Styles */
  .table-responsive {
    overflow-x: auto;
  }

  .table {
    width: 100%;
    min-width: 900px;
    border-collapse: collapse;
    font-size: 0.925rem;
  }
  .desktop-only {
    display: block;
  }
  .mobile-list {
    display: none;
  }

  .table th {
    text-align: left;
    padding: 14px 24px;
    background-color: #f8fafc;
    color: #64748b;
    font-weight: 600;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid #e2e8f0;
  }

  .table td {
    padding: 14px 24px;
    border-bottom: 1px solid #f1f5f9;
    vertical-align: middle;
  }

  .table-row {
    transition: background-color 0.15s ease;
  }

  .table-row:hover {
    background-color: #f8fafc;
  }

  .table-row:last-child td {
    border-bottom: none;
  }

  /* Attendance Info in Table */
  .attendance-info {
    display: grid;
    grid-template-columns: 32px 1fr;
    gap: 12px;
    align-items: center;
  }

  .attendance-icon-wrapper {
    width: 32px;
    height: 32px;
    background: #3b82f6;
    color: white;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
  }

  .attendance-details {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  .attendance-date {
    font-weight: 600;
    color: #1f2937;
    font-size: 0.95rem;
  }

  .user-info-inline {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .avatar-small {
    width: 20px;
    height: 20px;
    background: rgb(15 23 42);
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 0.65rem;
    flex-shrink: 0;
  }

  .intern-name-small {
    font-size: 0.8rem;
    color: #6b7280;
  }

  /* User Info in Table */
  .user-info {
    display: grid;
    grid-template-columns: 32px 1fr;
    gap: 12px;
    align-items: center;
  }

  .avatar-placeholder {
    width: 32px;
    height: 32px;
    background: rgb(15 23 42);
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 0.85rem;
    box-shadow: 0 2px 4px rgba(99, 102, 241, 0.2);
  }

  .mono {
    font-family: ui-monospace, monospace;
    font-size: 0.85rem;
  }

  .text-center {
    text-align: center;
  }
  .text-right {
    text-align: right;
    min-width: 150px;
    white-space: nowrap;
  }

  .pagination-pill {
    min-width: 128px;
  }

  /* Status Badges */
  .status-badge {
    padding: 5px 12px;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: capitalize;
    letter-spacing: 0.03em;
    display: inline-block;
    border: 1px solid transparent;
  }

  .bg-emerald-100 {
    background: #ecfdf5;
    border-color: #a7f3d0;
  }
  .text-emerald-700 {
    color: #047857;
  }
  .bg-yellow-100 {
    background: #fefce8;
    border-color: #fef08a;
  }
  .text-yellow-700 {
    color: #a16207;
  }
  .bg-blue-100 {
    background: #eff6ff;
    border-color: #bfdbfe;
  }
  .text-blue-700 {
    color: #1d4ed8;
  }
  .bg-purple-100 {
    background: #faf5ff;
    border-color: #e9d5ff;
  }
  .text-purple-700 {
    color: #7e22ce;
  }
  .bg-red-100 {
    background: #fef2f2;
    border-color: #fecaca;
  }
  .text-red-600 {
    color: #dc2626;
  }
  .bg-slate-100 {
    background: #f1f5f9;
    border-color: #e2e8f0;
  }
  .text-slate-600 {
    color: #475569;
  }
  .equal-badge {
    min-width: 96px;
    text-align: center;
    justify-content: center;
    display: inline-flex;
  }

  /* States */
  .loading-state {
    padding: 40px;
    text-align: center;
    color: #6b7280;
  }

  .empty-state {
    padding: 40px;
    text-align: center;
    color: #9ca3af;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .spinner {
    width: 24px;
    height: 24px;
    border: 3px solid #e5e7eb;
    border-top: 3px solid #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 12px;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  .btn-icon {
    width: 42px;
    height: 38px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 6px;
    border-radius: 6px;
    transition: all 0.2s;
  }

  .flex {
    display: flex;
  }
  .gap-2 {
    gap: 0.5rem;
  }
  .gap-3 {
    gap: 0.75rem;
  }
  .items-center {
    align-items: center;
  }
  .border-b {
    border-bottom: 1px solid #f1f5f9;
  }

  @media (max-width: 900px) {
    .desktop-only {
      display: none;
    }
    .mobile-list {
      display: flex;
      flex-direction: column;
      /* gap: 12px; */
    }
    .entry-card {
      padding: 14px;
      /* border-radius: 16px; */
      border-top: 1px solid #e2e8f0;
      background: #ffffff;
      box-shadow: 0 6px 20px -18px rgba(15, 23, 42, 0.3);
    }
    .entry-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      cursor: pointer;
    }
    .date-row {
      display: flex;
      align-items: center;
      gap: 8px;
      color: #0f172a;
      font-weight: 700;
      flex: 1;
      min-width: 0;
    }
    .date-icon {
      color: #6366f1;
      flex-shrink: 0;
    }
    .date-text {
      font-size: 0.95rem;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .head-right {
      display: flex;
      align-items: center;
      gap: 8px;
      flex-shrink: 0;
    }

    .entry-details {
      margin-top: 12px;
      padding-top: 12px;
      border-top: 1px solid #f1f5f9;
    }

    .intern-grid {
      display: grid;
      gap: 10px;
      margin-bottom: 12px;
    }
    .intern-box {
      padding: 12px;
      border: 1px solid #e2e8f0;
      border-radius: 14px;
      background: #f8fafc;
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .intern-box .avatar-mini {
      width: 28px;
      height: 28px;
      background: #0f172a;
      color: white;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 11px;
      font-weight: 600;
    }
    .intern-box-label {
      font-size: 14px;
      font-weight: 600;
      color: #334155;
    }

    .time-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 10px;
      margin-bottom: 12px;
    }
    .time-box {
      padding: 12px;
      border: 1px solid #e2e8f0;
      border-radius: 14px;
      background: #f8fafc;
      text-align: center;
    }
    .time-box .label {
      margin: 0;
      font-size: 12px;
      font-weight: 700;
      color: #94a3b8;
      text-transform: uppercase;
      letter-spacing: 0.03em;
    }
    .time-value {
      margin: 6px 0 0 0;
      font-size: 18px;
      font-weight: 800;
      color: #0f172a;
      letter-spacing: -0.02em;
    }

    .expand-btn {
      width: 32px;
      height: 32px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 50%;
      background: #f8fafc;
      color: #64748b;
      border: none;
      flex-shrink: 0;
    }
    .mobile-actions {
      display: flex;
      gap: 8px;
    }
    .mini-btn {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 8px 16px;
      border-radius: 9999px;
      border: 1px solid #0f172a;
      background: #0f172a;
      color: #fff;
      font-weight: 700;
      font-size: 13px;
      cursor: pointer;
      transition: all 0.15s ease;
      flex: 1;
      justify-content: center;
    }

    .mini-btn-circle {
      display: inline-flex;
      align-items: center;
      border-radius: 9999px;
      border: 1px solid #0f172a;
      background: #0f172a;
      color: #fff;
      font-weight: 700;
      font-size: 13px;
      cursor: pointer;
      transition: all 0.15s ease;
      /* flex: 1; */
      width: 42px;
      height: 42px;
      justify-content: center;
    }

    .mini-btn-circle.disabled {
      background: #e2e8f0;
      border-color: #e2e8f0;
      color: #94a3b8;
      cursor: not-allowed;
    }

    .mini-btn .btn-text {
      display: inline;
    }
    .mini-btn.danger {
      background: #ef4444;
      border-color: #ef4444;
    }
    .mini-btn.disabled {
      background: #e2e8f0;
      border-color: #e2e8f0;
      color: #94a3b8;
      cursor: not-allowed;
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

  .mt-4 {
    margin-top: 1rem;
  }
  .pt-4 {
    padding-top: 1rem;
  }
  .border-t {
    border-top: 1px solid;
  }
  .border-slate-100 {
    border-color: #f1f5f9;
  }
</style>
