<script>
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import { goto } from "@mateothegreat/svelte5-router";
  import { api } from "../lib/api.js";
  import { portal } from "../lib/portal.js";
  import { auth } from "../lib/auth.svelte.js";
  import { getAvatarUrl } from "../lib/utils.js";

  import ReportCreateModal from "./ReportCreateModal.svelte";
  import ReportEditModal from "./ReportEditModal.svelte";

  // State
  let reports = $state([]);
  let loading = $state(false);
  let isCreateModalOpen = $state(false);
  let isEditModalOpen = $state(false);
  let selectedReport = $state(null);
  let searchQuery = $state("");
  let filterStatus = $state("");
  let filterType = $state("");
  let currentPage = $state(1);
  let totalPages = $state(1);
  let totalItems = $state(0);
  let searchTimeout;
  let expandedReports = $state({});

  // Keep overlay-root click-through state in sync with our modals
  $effect(() => {
    const root =
      typeof document !== "undefined"
        ? document.querySelector("#overlay-root")
        : null;
    if (!(root instanceof HTMLElement)) return;
    const hasModal = isCreateModalOpen || isEditModalOpen;
    root.style.pointerEvents = hasModal ? "auto" : "none";
    if (!hasModal) {
      root.dataset.portalCount = "0";
    }
  });

  // Helpers UI
  function getTypeColor(type) {
    switch (type) {
      case "weekly":
        return "bg-blue-100 text-blue-700 border-blue-200";
      case "monthly":
        return "bg-purple-100 text-purple-700 border-purple-200";
      case "final":
        return "bg-rose-100 text-rose-700 border-rose-200";
      default:
        return "bg-slate-100 text-slate-600 border-slate-200";
    }
  }

  function getStatusColor(status) {
    switch (status) {
      case "approved":
      case "reviewed":
        return "bg-emerald-100 text-emerald-700 border-emerald-200";
      case "submitted":
        return "bg-blue-100 text-blue-700 border-blue-200";
      case "rejected":
        return "bg-rose-100 text-rose-700 border-rose-200";
      case "pending":
      default:
        return "bg-amber-100 text-amber-700 border-amber-200";
    }
  }

  function formatDate(dateStr) {
    if (!dateStr) return "—";
    return new Date(dateStr).toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    });
  }

  // --- Fetch Data ---
  async function fetchReports() {
    loading = true;
    try {
      const params = { page: currentPage, limit: 50 };
      if (searchQuery) params.search = searchQuery;
      if (filterStatus) params.status = filterStatus;
      if (filterType) params.type = filterType;

      const res = await api.getReports(params);
      reports = res.data || [];
      const pagination = res.pagination || {};
      totalPages = Math.max(pagination.total_pages || 1, 1);
      totalItems = pagination.total_items || 0;
      currentPage = pagination.page || currentPage;
      console.log("Fetched reports:", reports);
    } catch (err) {
      console.error("Failed to fetch reports:", err);
      alert("Gagal memuat data laporan: " + err.message);
    } finally {
      loading = false;
    }
  }

  function goToPreviousPage() {
    currentPage = Math.max(1, currentPage - 1);
    fetchReports();
  }

  function goToNextPage() {
    if (currentPage >= totalPages) return;
    currentPage += 1;
    fetchReports();
  }

  function handleSearchInput() {
    clearTimeout(searchTimeout);
    currentPage = 1;
    searchTimeout = setTimeout(() => {
      fetchReports();
    }, 500);
  }

  async function handleDelete(id, title) {
    if (
      !confirm(
        `Apakah Anda yakin ingin menghapus laporan "${title || "Tanpa Judul"}"?`,
      )
    )
      return;
    try {
      await api.deleteReport(id);
      reports = reports.filter((r) => r.id !== id);
      alert("Laporan berhasil dihapus.");
    } catch (err) {
      console.error(err);
      alert("Gagal menghapus laporan: " + err.message);
    }
  }

  function openEditModal(report) {
    selectedReport = report;
    isEditModalOpen = true;
  }

  function toggleExpand(id) {
    expandedReports[id] = !expandedReports[id];
  }

  onMount(async () => {
    await fetchReports();
  });
</script>

<div class="page-container animate-fade-in">
  <div class="flex items-center gap-3 pb-8">
    <h4 class="card-title">Laporan & Jurnal</h4>
    <span class="badge-count">{reports.length} dari {totalItems} Laporan</span>
  </div>

  <!-- BAGIAN TABEL DAFTAR -->
  <div class="card table-card animate-slide-up" style="animation-delay: 0.1s;">
    <div class="card-header-row border-b">
      <div class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2">
        <button
          class="cursor-pointer flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-slate-900 text-white hover:bg-slate-800 transition-all shadow-sm flex items-center justify-center gap-2"
          onclick={() => (isCreateModalOpen = true)}
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
            ><line x1="12" y1="5" x2="12" y2="19"></line><line
              x1="5"
              y1="12"
              x2="19"
              y2="12"
            ></line></svg
          >
          <span>Buat Laporan</span>
        </button>
        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2"
          onclick={fetchReports}
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
      <div
        class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2 {totalPages <=
        1
          ? 'opacity-50 pointer-events-none'
          : ''}"
      >
        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2 {currentPage <=
          1
            ? 'opacity-50 cursor-not-allowed'
            : 'cursor-pointer'}"
          onclick={goToPreviousPage}
          disabled={currentPage <= 1}
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
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2 pagination-pill"
        >
          <span>{currentPage}</span>
          <span class="text-slate-500">of</span>
          <span>{totalPages}</span>
        </div>

        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2 {currentPage >=
          totalPages
            ? 'opacity-50 cursor-not-allowed'
            : 'cursor-pointer'}"
          onclick={goToNextPage}
          disabled={currentPage >= totalPages}
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
            e.key === "Enter" && (clearTimeout(searchTimeout), fetchReports())}
          placeholder="Cari Judul atau Nama Intern..."
          class="search-input"
        />
      </div>

      <select
        bind:value={filterType}
        onchange={fetchReports}
        class="filter-select"
      >
        <option value="">Semua Tipe</option>
        <option value="weekly">Mingguan</option>
        <option value="monthly">Bulanan</option>
        <option value="final">Final</option>
      </select>

      <select
        bind:value={filterStatus}
        onchange={fetchReports}
        class="filter-select"
      >
        <option value="">Semua Status</option>
        <option value="submitted">Dikirim</option>
        <option value="reviewed">Direview</option>
        <option value="approved">Approved</option>
        <option value="rejected">Rejected</option>
        <option value="pending">Pending</option>
        <option value="draft">Draft</option>
      </select>
    </div>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Memuat data...</p>
      </div>
    {:else if reports.length === 0}
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
          ><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"
          ></path><polyline points="13 2 13 9 20 9"></polyline></svg
        >
        <p>Belum ada data laporan.</p>
      </div>
    {:else}
      <div class="table-responsive desktop-only">
        <table class="table">
          <thead>
            <tr>
              <th>Judul & Tipe</th>
              <th>Peserta</th>
              <th>Periode</th>
              <th>Status</th>
              <th class="text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {#each reports as r}
              <tr class="table-row">
                <td style="min-width: 240px;">
                  <div class="report-info">
                    <div class="report-icon-wrapper">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path
                          d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"
                        ></path>
                        <polyline points="13 2 13 9 20 9"></polyline>
                      </svg>
                    </div>
                    <div class="report-details">
                      <span class="report-name">{r.title || "Tanpa Judul"}</span
                      >
                      <span class={`type-badge ${getTypeColor(r.type)}`}
                        >{r.type || "-"}</span
                      >
                    </div>
                  </div>
                </td>
                <td>
                  <div class="user-info">
                    {#if r.intern_avatar && getAvatarUrl(r.intern_avatar)}
                      <img
                        src={getAvatarUrl(r.intern_avatar)}
                        alt={r.intern_name}
                        class="w-8 h-8 rounded-full object-cover"
                      />
                    {:else}
                      <div class="avatar-placeholder">
                        {r.intern_name?.charAt(0) || "U"}
                      </div>
                    {/if}
                    <span>{r.intern_name || "-"}</span>
                  </div>
                </td>
                <td class="text-muted mono">
                  {formatDate(r.period_start)} - {formatDate(r.period_end)}
                </td>
                <td class="text-center">
                  <span
                    class={`status-badge equal-badge ${getStatusColor(r.status || "pending")}`}
                  >
                    {r.status || "Pending"}
                  </span>
                </td>
                <td class="text-right">
                  <button
                    class="btn-icon text-sky-600 hover:text-sky-700 bg-sky-50 hover:bg-sky-100"
                    onclick={() => goto(`/reports/${r.id}`)}
                    title="Lihat Detail"
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
                    class="btn-icon text-slate-600 hover:text-slate-700 bg-slate-50 hover:bg-slate-100"
                    onclick={(e) => {
                      e.stopPropagation();
                      openEditModal(r);
                    }}
                    title="Edit Data"
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
                        d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
                      ></path>
                      <path
                        d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"
                      ></path>
                    </svg>
                  </button>
                  <button
                    class="btn-icon text-slate-600 hover:text-slate-700 bg-slate-50 hover:bg-slate-100"
                    onclick={() => handleDelete(r.id, r.title)}
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
        {#each reports as r}
          <div class="entry-card">
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="entry-head" onclick={() => toggleExpand(r.id)}>
              <div class="report-info">
                <div class="report-icon-wrapper">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="18"
                    height="18"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path
                      d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"
                    ></path>
                    <polyline points="13 2 13 9 20 9"></polyline>
                  </svg>
                </div>
                <div class="report-details">
                  <div class="report-name">{r.title || "Tanpa Judul"}</div>
                  <div class="text-muted small">{r.intern_name || "-"}</div>
                </div>
              </div>
              <button class="expand-btn">
                <span
                  class="material-symbols-outlined transition-transform duration-200 {expandedReports[
                    r.id
                  ]
                    ? 'rotate-180'
                    : ''}">expand_more</span
                >
              </button>
            </div>

            {#if expandedReports[r.id]}
              <div class="entry-details" transition:slide={{ duration: 200 }}>
                <div class="detail-row">
                  <div class="detail-label">TIPE</div>
                  <span class={`type-badge equal-badge ${getTypeColor(r.type)}`}
                    >{r.type || "-"}</span
                  >
                </div>
                <div class="detail-row">
                  <div class="detail-label">PESERTA</div>
                  <div class="user-info">
                    {#if r.intern_avatar && getAvatarUrl(r.intern_avatar)}
                      <img
                        src={getAvatarUrl(r.intern_avatar)}
                        alt={r.intern_name}
                        class="w-8 h-8 rounded-full object-cover"
                      />
                    {:else}
                      <div class="avatar-placeholder">
                        {r.intern_name?.charAt(0) || "U"}
                      </div>
                    {/if}
                    <span>{r.intern_name || "-"}</span>
                  </div>
                </div>
                <div class="detail-row">
                  <div class="detail-label">PERIODE</div>
                  <div class="detail-value mono">
                    {formatDate(r.period_start)} - {formatDate(r.period_end)}
                  </div>
                </div>
                <div class="detail-row">
                  <div class="detail-label">STATUS</div>
                  <span
                    class={`status-badge equal-badge ${getStatusColor(r.status || "pending")}`}
                  >
                    {r.status || "Pending"}
                  </span>
                </div>

                <div class="mobile-actions mt-4 pt-4 border-t border-slate-100">
                  <button
                    class="mini-btn mobile"
                    onclick={(e) => {
                      e.stopPropagation();
                      goto(`/reports/${r.id}`);
                    }}
                  >
                    <span class="material-symbols-outlined">visibility</span>
                    <span class="btn-text">Detail</span>
                  </button>
                  <button
                    class="mini-btn-circle mobile"
                    onclick={(e) => {
                      e.stopPropagation();
                      openEditModal(r);
                    }}
                  >
                    <span class="material-symbols-outlined">edit</span>
                  </button>
                  <button
                    class="mini-btn mobile danger"
                    onclick={(e) => {
                      e.stopPropagation();
                      handleDelete(r.id, r.title);
                    }}
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

<ReportCreateModal
  isOpen={isCreateModalOpen}
  onClose={() => (isCreateModalOpen = false)}
  onSuccess={() => {
    isCreateModalOpen = false;
    fetchReports();
  }}
/>

<ReportEditModal
  isOpen={isEditModalOpen}
  report={selectedReport}
  onClose={() => {
    isEditModalOpen = false;
    selectedReport = null;
  }}
  onSuccess={() => {
    isEditModalOpen = false;
    selectedReport = null;
    fetchReports();
  }}
/>

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

  /* Report Info in Table */
  .report-info {
    display: grid;
    grid-template-columns: 32px 1fr;
    gap: 12px;
    align-items: center;
  }

  .report-icon-wrapper {
    width: 32px;
    height: 32px;
    background: #8b5cf6;
    color: white;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 2px 4px rgba(139, 92, 246, 0.2);
  }

  .report-details {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  .report-name {
    font-weight: 600;
    color: #1f2937;
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

  /* Modal z-index helpers */
  :global(.z-120) {
    z-index: 120;
  }
  :global(.z-110) {
    z-index: 110;
  }
  :global(.z-100) {
    z-index: 100;
  }

  /* Type Badges */
  .type-badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 6px;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    width: fit-content;
    border: 1px solid transparent;
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
    color: #7c3aed;
  }
  .bg-rose-100 {
    background: #fff1f2;
    border-color: #fecdd3;
  }
  .text-rose-700 {
    color: #be123c;
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
  .bg-amber-100 {
    background: #fefce8;
    border-color: #fef08a;
  }
  .text-amber-700 {
    color: #a16207;
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
      /* border-bottom: 1px solid #e2e8f0; */
      background: #ffffff;
      box-shadow: none;
    }
    .entry-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      cursor: pointer;
    }
    .entry-head .report-details {
      display: flex;
      flex-direction: column;
      min-width: 0;
    }
    .entry-head .report-name {
      font-size: 0.95rem;
      font-weight: 600;
      color: #0f172a;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .entry-head .text-muted {
      font-size: 0.8rem;
      color: #64748b;
      margin: 0;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .entry-head .report-icon-wrapper {
      width: 32px;
      height: 32px;
      font-size: 1rem;
      flex-shrink: 0;
    }
    .entry-details {
      margin-top: 16px;
      padding-top: 16px;
      border-top: 1px solid #f1f5f9;
    }

    .entry-head .report-info {
      display: grid;
      grid-template-columns: 32px 1fr;
      gap: 12px;
      align-items: center;
      flex: 1;
      min-width: 0;
    }

    .detail-row {
      margin-bottom: 16px;
    }
    .detail-row:last-child {
      margin-bottom: 0;
    }
    .detail-label {
      font-size: 11px;
      font-weight: 700;
      color: #94a3b8;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      margin-bottom: 4px;
    }
    .detail-value {
      font-weight: 600;
      color: #0f172a;
      font-size: 14px;
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
    }
    .mobile-actions {
      display: flex;
      gap: 10px;
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

    .mini-btn .btn-text {
      display: inline;
    }
    .mini-btn.danger {
      background: #ef4444;
      border-color: #ef4444;
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

  .mt-3 {
    margin-top: 0.75rem;
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
