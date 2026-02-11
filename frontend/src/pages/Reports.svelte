<script>
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import { goto } from "@mateothegreat/svelte5-router";
  import { api } from "../lib/api.js";
  import { auth } from "../lib/auth.svelte.js";
  import ReportCreateModal from "./ReportCreateModal.svelte";
  import ReportEditModal from "./ReportEditModal.svelte";

  // State
  let reports = $state([]);
  let loading = $state(true);
  let isModalOpen = $state(false);
  let isEditModalOpen = $state(false);
  let selectedReport = $state(null);
  let expandedReports = $state({});
  let searchQuery = $state("");
  let filterStatus = $state("");
  let filterType = $state("");
  let searchTimeout;

  async function fetchReports() {
    loading = true;
    try {
      const params = { page: 1, limit: 50 };
      if (searchQuery) params.search = searchQuery;
      if (filterStatus) params.status = filterStatus;
      if (filterType) params.type = filterType;

      const res = await api.getReports(params);
      reports = res.data || [];
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  function handleSearchInput() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      fetchReports();
    }, 500);
  }

  function toggleExpand(id) {
    expandedReports[id] = !expandedReports[id];
  }

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
    if (!dateStr) return "-";
    return new Date(dateStr).toLocaleDateString("id-ID", {
      day: "numeric",
      month: "short",
    });
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

  onMount(fetchReports);
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    <!-- Header -->
    <div class="header">
      <div class="header-content">
        <h2 class="title">Laporan & Jurnal</h2>
        <p class="subtitle">
          Dokumentasi aktivitas mingguan dan bulanan peserta magang.
        </p>
      </div>

      {#if auth.user?.role !== "intern"}
        <button class="btn-primary" onclick={() => (isModalOpen = true)}>
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          Buat Laporan
        </button>
      {:else}
        <!-- Interns also need to create reports -->
        <button class="btn-primary" onclick={() => (isModalOpen = true)}>
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          Buat Laporan
        </button>
      {/if}
    </div>

    <!-- TABEL DAFTAR LAPORAN -->
    <div class="card list-card animate-slide-up" style="animation-delay: 0.1s;">
      <div class="card-header border-b">
        <h3>Daftar Laporan Masuk</h3>
        <span class="badge-count">{reports.length} File</span>
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
              (clearTimeout(searchTimeout), fetchReports())}
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
          <option value="submitted">Submitted</option>
          <option value="reviewed">Reviewed</option>
          <option value="draft">Draft</option>
        </select>
      </div>

      {#if loading}
        <div class="loading-state">Memuat data laporan...</div>
      {:else if reports.length === 0}
        <div class="empty-state">
          <div class="empty-icon">📂</div>
          <p>Belum ada laporan yang dikumpulkan.</p>
        </div>
      {:else}
        <!-- DESKTOP TABLE -->
        <div class="table-container desktop-only">
          <table class="table">
            <thead>
              <tr>
                <th>Judul & Tipe</th>
                <th>Peserta</th>
                <th>Periode</th>
                <th>Status</th>
                <th class="text-right">Opsi</th>
              </tr>
            </thead>
            <tbody>
              {#each reports as r}
                <tr class="hover-row">
                  <td style="min-width: 240px;">
                    <div class="report-info">
                      <span class={`type-badge ${getTypeColor(r.type)}`}
                        >{r.type}</span
                      >
                      <span class="report-title"
                        >{r.title || "Tanpa Judul"}</span
                      >
                    </div>
                  </td>
                  <td>
                    <div class="user-info">
                      <div class="avatar">
                        {r.intern_name?.charAt(0) || "U"}
                      </div>
                      <div class="user-text">
                        {r.intern_name}
                      </div>
                    </div>
                  </td>
                  <td class="text-slate-500 text-sm">
                    {formatDate(r.period_start)} - {formatDate(r.period_end)}
                  </td>
                  <td>
                    <span
                      class={`status-badge ${getStatusColor(r.status || "pending")}`}
                    >
                      {r.status || "Pending"}
                    </span>
                  </td>
                  <td class="text-right">
                    <button
                      class="btn-icon text-slate-500 hover:text-blue-600"
                      title="Detail"
                      onclick={() => goto(`/reports/${r.id}`)}
                    >
                      <svg
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        ><path
                          d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                        /><circle cx="12" cy="12" r="3" /></svg
                      >
                    </button>
                    <button
                      class="btn-icon text-emerald-600 hover:text-emerald-700"
                      onclick={() => openEditModal(r)}
                      title="Edit Laporan"
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
                      class="btn-icon danger"
                      onclick={() => handleDelete(r.id, r.title)}
                      title="Hapus Laporan"
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

        <!-- MOBILE LIST -->
        <div class="mobile-list">
          {#each reports as r}
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="entry-card" onclick={() => toggleExpand(r.id)}>
              <div class="entry-head">
                <div class="user-info mobile">
                  <div class="avatar">{r.intern_name?.charAt(0) || "U"}</div>
                  <div class="user-details">
                    <div class="user-name">{r.intern_name || "Intern"}</div>
                    <div class="report-title-mobile">
                      {r.title || "Tanpa Judul"}
                    </div>
                  </div>
                </div>
                <button class="expand-btn">
                  <span
                    class="material-symbols-outlined transition-transform {expandedReports[
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
                    <span class={`type-badge ${getTypeColor(r.type)}`}
                      >{r.type}</span
                    >
                  </div>
                  <div class="detail-row">
                    <div class="detail-label">PERIODE</div>
                    <div class="detail-value">
                      {formatDate(r.period_start)} - {formatDate(r.period_end)}
                    </div>
                  </div>
                  <div class="detail-row">
                    <div class="detail-label">STATUS</div>
                    <span
                      class={`status-badge ${getStatusColor(r.status || "pending")}`}
                    >
                      {r.status || "Pending"}
                    </span>
                  </div>

                  <div
                    class="mobile-actions mt-4 pt-4 border-t border-slate-100"
                  >
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
                    <!-- {#if auth.user && (["admin", "supervisor"].includes(auth.user.role) || r.intern_id == auth.user.id)} -->
                    <button
                      class="mini-btn mobile"
                      onclick={(e) => {
                        e.stopPropagation();
                        openEditModal(r);
                      }}
                    >
                      <span class="material-symbols-outlined">edit</span>
                      <span class="btn-text">Edit</span>
                    </button>
                    <!-- {/if} -->
                    <!-- {#if auth.user?.role !== "intern" || r.intern_id === auth.user?.id} -->
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
                    <!-- {/if} -->
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
    isOpen={isModalOpen}
    onClose={() => (isModalOpen = false)}
    onSuccess={fetchReports}
  />

  <ReportEditModal
    isOpen={isEditModalOpen}
    report={selectedReport}
    onClose={() => (isEditModalOpen = false)}
    onSuccess={fetchReports}
  />
</div>

<style>
  :root {
    --space-xs: 8px;
    --space-sm: 12px;
    --space-md: 16px;
    --space-lg: 24px;

    --avatar-size: 32px;
  }

  :global(body) {
    font-family: "Geist", "Inter", sans-serif;
    color: #0f172a;
  }

  .page-bg {
    min-height: 100vh;
    background-color: #f8fafc;
    /* background-image: radial-gradient(
        at 0% 0%,
        rgba(16, 185, 129, 0.03) 0%,
        transparent 50%
      ),
      radial-gradient(
        at 100% 100%,
        rgba(14, 165, 233, 0.03) 0%,
        transparent 50%
      ); */
    padding: 0;
    overflow-x: hidden; /* Check for overflow */
  }

  .container {
    max-width: 1200px;
    margin: 0 auto;
  }

  /* --- HEADER --- */
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    flex-wrap: wrap;
    gap: 16px;
  }
  .title {
    font-size: 20px;
    font-weight: 600;
    color: #0f172a;
    margin: 0 0 0 0;
    letter-spacing: -0.02em;
  }
  .subtitle {
    color: #64748b;
    font-size: 16px;
    margin: 0;
  }

  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669);
    color: white;
    padding: 10px 20px;
    border-radius: 999px;
    font-weight: 600;
    font-size: 14px;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.2);
    transition: all 0.2s;
  }
  .btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(16, 185, 129, 0.3);
  }

  /* --- CARD --- */
  .card {
    background: white;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.02),
      0 2px 4px -1px rgba(0, 0, 0, 0.02);
    overflow: hidden;
    margin-bottom: 32px;
  }
  .card-header {
    padding: 20px 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .card-header h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: #1e293b;
  }
  .border-b {
    border-bottom: 1px solid #f1f5f9;
  }

  /* --- TOOLBAR --- */
  .toolbar {
    padding: 16px 24px;
    background: #fcfcfc;
    border-bottom: 1px solid #f1f5f9;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    align-items: center;
  }
  .search-wrapper {
    position: relative;
    flex: 1;
    min-width: 280px;
  }
  .search-icon {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    color: #94a3b8;
    font-size: 20px;
  }
  .search-input {
    width: 100%;
    padding: 10px 12px 10px 40px;
    border: 1px solid #e2e8f0;
    border-radius: 999px;
    font-size: 14px;
    transition: all 0.2s;
  }
  .search-input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }
  .filter-select {
    padding: 10px 16px;
    border: 1px solid #e2e8f0;
    border-radius: 999px;
    font-size: 14px;
    color: #475569;
    background: white;
    cursor: pointer;
    min-width: 140px;
  }

  /* --- TABLE --- */
  .badge-count {
    background: #f1f5f9;
    color: #64748b;
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 600;
  }

  .table-container {
    overflow-x: auto;
  }
  .table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
  }

  .table th {
    text-align: left;
    padding: 16px 24px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    color: #64748b;
    border-bottom: 1px solid #e2e8f0;
    background: #fcfcfc;
  }
  .table td {
    padding: var(--space-md) var(--space-lg);
    height: 64px;
    border-bottom: 1px solid #f1f5f9;
    vertical-align: middle;
    color: #334155;
  }
  .table tr:last-child td {
    border-bottom: none;
  }
  .hover-row:hover td {
    background-color: #f8fafc;
  }

  /* Table Content */
  .report-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .report-title {
    font-weight: 600;
    color: #0f172a;
    font-size: 14px;
  }

  .user-info
  /* .user-info.mobile { */ {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    min-width: 0;
  }

  .user-info.mobile {
    display: flex;
    flex-direction: row; /* critical */
    align-items: center;
    gap: 12px;
  }

  .avatar {
    width: var(--avatar-size);
    height: var(--avatar-size);
    line-height: var(--avatar-size);
    background: #0f172a;
    color: white;
    border-radius: 999px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 12px;
    flex-shrink: 0;
  }
  .name {
    font-weight: 500;
    font-size: 14px;
  }

  .name,
  .user-name,
  .report-title,
  .report-title-mobile {
    line-height: 1.2;
  }

  .text-right {
    text-align: right;
  }
  .text-slate-500 {
    color: #64748b;
  }
  .text-sm {
    font-size: 13px;
  }

  /* Badges */
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
  .bg-purple-100 {
    background: #faf5ff;
    border-color: #e9d5ff;
  }
  .bg-rose-100 {
    background: #fff1f2;
    border-color: #fecdd3;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    padding: 4px 12px;
    border-radius: 999px;
    font-size: 12px;
    font-weight: 600;
    border: 1px solid transparent;
    text-transform: capitalize;
  }
  .bg-emerald-100 {
    background: #ecfdf5;
    border-color: #a7f3d0;
  }
  .bg-amber-100 {
    background: #fefce8;
    border-color: #fef08a;
  }
  .bg-red-100 {
    background: #fef2f2;
    border-color: #fecaca;
  }
  .bg-slate-100 {
    background: #f1f5f9;
    border-color: #e2e8f0;
  }

  .btn-icon {
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    padding: 6px;
    border-radius: 6px;
    transition: all 0.2s;
  }
  .btn-icon:hover {
    background: #e2e8f0;
    color: #0f172a;
  }

  /* States */
  .empty-state,
  .loading-state {
    text-align: center;
    padding: 60px 20px;
    color: #94a3b8;
    font-style: italic;
  }
  .empty-icon {
    font-size: 32px;
    margin-bottom: 12px;
    opacity: 0.5;
  }

  /* Mobile Styles */
  .mobile-list {
    display: none;
  }
  .desktop-only {
    display: block;
  }

  @media (max-width: 900px) {
    .page-bg {
      padding: 0px;
    }
    .header {
      flex-direction: column;
      align-items: stretch;
      gap: 16px;
    }
    .btn-primary {
      width: 100%;
      justify-content: center;
      margin-top: 8px;
    }

    .desktop-only {
      display: none;
    }
    .mobile-list {
      display: flex;
      flex-direction: column;
      gap: 12px;
      padding: 0;
    }

    .card.list-card {
      background: transparent;
      border: none;
      box-shadow: none;
    }
    .card.list-card .card-header {
      display: none;
    } /* Optional: hide header on mobile if desired */

    .list-card {
      padding: 0px;
    }

    .entry-card {
      padding: var(--space-md);
      border-radius: 16px;
      border: 1px solid #e2e8f0;
      background: #ffffff;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
    }

    .entry-head {
      display: flex;
      justify-content: space-between;
      align-items: center;
      cursor: pointer;
      min-height: 64px;
    }

    .user-details {
      display: flex;
      flex-direction: column;
      justify-content: center;
    }

    *,
    *::before,
    *::after {
      box-sizing: border-box;
    }

    .user-name {
      font-weight: 600;
      font-size: 14px;
      color: #0f172a;
    }
    .report-title-mobile {
      font-size: 13px;
      color: #64748b;
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

    .entry-details {
      margin-top: 16px;
      padding-top: 16px;
      border-top: 1px solid #f1f5f9;
    }
    .detail-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
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
    }
    .detail-value {
      font-weight: 600;
      color: #0f172a;
      font-size: 14px;
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
    .mini-btn .btn-text {
      display: inline;
    }

    .rotate-180 {
      transform: rotate(180deg);
    }
  }

  /* Animation */
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
</style>
