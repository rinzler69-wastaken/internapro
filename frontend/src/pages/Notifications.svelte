<script>
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import { goto } from "@mateothegreat/svelte5-router";
  import { api } from "../lib/api.js";
  import { portal } from "../lib/portal.js";
  import { auth } from "../lib/auth.svelte.js";

  // State
  let notifications = $state([]);
  let loading = $state(false);
  let searchQuery = $state("");
  let filterRead = $state("");
  let currentPage = $state(1);
  let totalPages = $state(1);
  let totalItems = $state(0);
  let searchTimeout;
  let expandedNotifications = $state({});

  function formatDate(dateStr) {
    if (!dateStr) return "—";
    return new Date(dateStr).toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function getTimeAgo(dateStr) {
    if (!dateStr) return "";
    const now = new Date();
    const date = new Date(dateStr);
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (seconds < 60) return "baru saja";
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes} menit lalu`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours} jam lalu`;
    const days = Math.floor(hours / 24);
    if (days < 7) return `${days} hari lalu`;
    return formatDate(dateStr);
  }

  // --- Fetch Data ---
  async function fetchNotifications() {
    loading = true;
    try {
      const params = { page: currentPage, limit: 50 };
      if (searchQuery) params.search = searchQuery;
      if (filterRead === "read") params.is_read = true;
      if (filterRead === "unread") params.is_read = false;

      const res = await api.getNotifications(params);
      notifications = res.data || [];
      const pagination = res.pagination || {};
      totalPages = Math.max(pagination.total_pages || 1, 1);
      totalItems = pagination.total_items || 0;
      currentPage = pagination.page || currentPage;
      console.log("Fetched notifications:", notifications);
    } catch (err) {
      console.error("Failed to fetch notifications:", err);
      alert("Gagal memuat data notifikasi: " + err.message);
    } finally {
      loading = false;
    }
  }

  function goToPreviousPage() {
    currentPage = Math.max(1, currentPage - 1);
    fetchNotifications();
  }

  function goToNextPage() {
    if (currentPage >= totalPages) return;
    currentPage += 1;
    fetchNotifications();
  }

  function handleSearchInput() {
    clearTimeout(searchTimeout);
    currentPage = 1;
    searchTimeout = setTimeout(() => {
      fetchNotifications();
    }, 500);
  }

  async function markRead(id) {
    try {
      await api.markNotificationRead(id);
      // Update locally
      const index = notifications.findIndex((n) => n.id === id);
      if (index !== -1) {
        notifications[index].is_read = true;
      }
    } catch (err) {
      console.error(err);
      alert("Gagal menandai notifikasi: " + err.message);
    }
  }

  async function markAll() {
    if (!confirm("Tandai semua notifikasi sebagai sudah dibaca?")) return;
    try {
      await api.markAllNotificationsRead();
      await fetchNotifications();
      alert("Semua notifikasi telah ditandai sebagai sudah dibaca.");
    } catch (err) {
      console.error(err);
      alert("Gagal menandai semua notifikasi: " + err.message);
    }
  }

  async function handleDelete(id) {
    if (!confirm("Apakah Anda yakin ingin menghapus notifikasi ini?")) return;
    try {
      await api.deleteNotification(id);
      notifications = notifications.filter((n) => n.id !== id);
    } catch (err) {
      console.error(err);
      alert("Gagal menghapus notifikasi: " + err.message);
    }
  }

  function toggleExpand(id) {
    expandedNotifications[id] = !expandedNotifications[id];
  }

  onMount(async () => {
    await fetchNotifications();
  });
</script>

<div class="page-container animate-fade-in">
  <div class="flex items-center gap-3 pb-8">
    <h4 class="card-title">Notifikasi</h4>
    <span class="badge-count"
      >{notifications.length} dari {totalItems} Notifikasi</span
    >
  </div>

  <!-- BAGIAN TABEL DAFTAR -->
  <div class="card table-card animate-slide-up" style="animation-delay: 0.1s;">
    <div class="card-header-row border-b">
      <div class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2">
        <button
          class="cursor-pointer flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-slate-900 text-white hover:bg-slate-800 transition-all shadow-sm flex items-center justify-center gap-2"
          onclick={markAll}
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
            ><polyline points="20 6 9 17 4 12"></polyline></svg
          >
          <span>Tandai Semua</span>
        </button>
        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2"
          onclick={fetchNotifications}
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
            e.key === "Enter" &&
            (clearTimeout(searchTimeout), fetchNotifications())}
          placeholder="Cari Notifikasi..."
          class="search-input"
        />
      </div>

      <select
        bind:value={filterRead}
        onchange={fetchNotifications}
        class="filter-select"
      >
        <option value="">Semua Status</option>
        <option value="unread">Belum Dibaca</option>
        <option value="read">Sudah Dibaca</option>
      </select>
    </div>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Memuat data...</p>
      </div>
    {:else if notifications.length === 0}
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
          ><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path><path
            d="M13.73 21a2 2 0 0 1-3.46 0"
          ></path></svg
        >
        <p>Tidak ada notifikasi.</p>
      </div>
    {:else}
      <div class="table-responsive desktop-only">
        <table class="table">
          <thead>
            <tr>
              <th>Notifikasi</th>
              <th>Pesan</th>
              <th>Waktu</th>
              <th class="text-center">Status</th>
              <th class="text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {#each notifications as n}
              <tr class="table-row {n.is_read ? '' : 'unread-row'}">
                <td style="min-width: 200px;">
                  <div class="notification-info">
                    <div
                      class="notification-icon-wrapper {n.is_read
                        ? 'read'
                        : 'unread'}"
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"
                        ></path>
                        <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
                      </svg>
                      {#if !n.is_read}
                        <span class="unread-dot"></span>
                      {/if}
                    </div>
                    <div class="notification-details">
                      <span class="notification-title"
                        >{n.title || "Notifikasi"}</span
                      >
                    </div>
                  </div>
                </td>
                <td>
                  <span class="notification-message">{n.message || "—"}</span>
                </td>
                <td class="text-muted mono">
                  {getTimeAgo(n.created_at)}
                </td>
                <td class="text-center">
                  <span
                    class={`status-badge equal-badge ${n.is_read ? "bg-slate-100 text-slate-600 border-slate-200" : "bg-blue-100 text-blue-700 border-blue-200"}`}
                  >
                    {n.is_read ? "Dibaca" : "Baru"}
                  </span>
                </td>
                <td class="text-right">
                  <div class="flex justify-end gap-1">
                    <button
                      class="btn-icon {n.link
                        ? 'text-indigo-600 hover:text-indigo-700 bg-indigo-50 hover:bg-indigo-100'
                        : 'text-slate-300 bg-slate-50 cursor-not-allowed opacity-50'}"
                      onclick={() => n.link && goto(n.link)}
                      disabled={!n.link}
                      title={n.link ? "Pergi ke Tautan" : "Tidak ada tautan"}
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
                          d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"
                        ></path>
                        <polyline points="15 3 21 3 21 9"></polyline>
                        <line x1="10" y1="14" x2="21" y2="3"></line>
                      </svg>
                    </button>
                    {#if !n.is_read}
                      <button
                        class="btn-icon text-sky-600 hover:text-sky-700 bg-sky-50 hover:bg-sky-100"
                        onclick={() => markRead(n.id)}
                        title="Tandai Dibaca"
                      >
                        <svg
                          width="18"
                          height="18"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                        >
                          <path d="M9 11l3 3L22 4"></path>
                          <path
                            d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"
                          ></path>
                        </svg>
                      </button>
                    {/if}
                    <button
                      class="btn-icon text-slate-600 hover:text-slate-700 bg-slate-50 hover:bg-slate-100"
                      onclick={() => handleDelete(n.id)}
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
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="mobile-list">
        {#each notifications as n}
          <div class="entry-card {n.is_read ? '' : 'unread-card'}">
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="entry-head" onclick={() => toggleExpand(n.id)}>
              <div class="notification-info">
                <div
                  class="notification-icon-wrapper {n.is_read
                    ? 'read'
                    : 'unread'}"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="18"
                    height="18"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"
                    ></path>
                    <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
                  </svg>
                  {#if !n.is_read}
                    <span class="unread-dot"></span>
                  {/if}
                </div>
                <div class="notification-details">
                  <div class="notification-title">
                    {n.title || "Notifikasi"}
                  </div>
                  <div class="text-muted small">{getTimeAgo(n.created_at)}</div>
                </div>
              </div>
              <button class="expand-btn">
                <span
                  class="material-symbols-outlined transition-transform duration-200 {expandedNotifications[
                    n.id
                  ]
                    ? 'rotate-180'
                    : ''}">expand_more</span
                >
              </button>
            </div>

            {#if expandedNotifications[n.id]}
              <div class="entry-details" transition:slide={{ duration: 200 }}>
                <div class="detail-row">
                  <div class="detail-label">PESAN</div>
                  <div class="detail-value message-value">
                    {n.message || "—"}
                  </div>
                </div>
                <div class="detail-row">
                  <div class="detail-label">WAKTU</div>
                  <div class="detail-value mono">
                    {formatDate(n.created_at)}
                  </div>
                </div>
                <div class="detail-row">
                  <div class="detail-label">STATUS</div>
                  <span
                    class={`status-badge equal-badge ${n.is_read ? "bg-slate-100 text-slate-600 border-slate-200" : "bg-blue-100 text-blue-700 border-blue-200"}`}
                  >
                    {n.is_read ? "Dibaca" : "Baru"}
                  </span>
                </div>

                <div class="mobile-actions mt-4 pt-4 border-t border-slate-100">
                  <button
                    class="mini-btn mobile"
                    onclick={(e) => {
                      e.stopPropagation();
                      if (n.link) goto(n.link);
                    }}
                    disabled={!n.link}
                    style={!n.link
                      ? "opacity: 0.5; cursor: not-allowed; background: #f1f5f9; color: #cbd5e1;"
                      : ""}
                  >
                    <span class="material-symbols-outlined">open_in_new</span>
                    <span class="btn-text">Buka</span>
                  </button>
                  {#if !n.is_read}
                    <button
                      class="mini-btn mobile"
                      onclick={(e) => {
                        e.stopPropagation();
                        markRead(n.id);
                      }}
                    >
                      <span class="material-symbols-outlined">check_box</span>
                      <span class="btn-text">Baca</span>
                    </button>
                  {/if}
                  <button
                    class="mini-btn mobile danger"
                    onclick={(e) => {
                      e.stopPropagation();
                      handleDelete(n.id);
                    }}
                    style="background-color: #ef4444; border-color: #ef4444;"
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

  .table-row.unread-row {
    background-color: #eff6ff;
  }

  .table-row.unread-row:hover {
    background-color: #dbeafe;
  }

  .table-row:last-child td {
    border-bottom: none;
  }

  /* Notification Info in Table */
  .notification-info {
    display: grid;
    grid-template-columns: 32px 1fr;
    gap: 12px;
    align-items: center;
  }

  .notification-icon-wrapper {
    width: 32px;
    height: 32px;
    background: #94a3b8;
    color: white;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 2px 4px rgba(148, 163, 184, 0.2);
    position: relative;
  }

  .notification-icon-wrapper.unread {
    background: #3b82f6;
    box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);
  }

  .notification-icon-wrapper.read {
    background: #94a3b8;
  }

  .unread-dot {
    position: absolute;
    top: -2px;
    right: -2px;
    width: 10px;
    height: 10px;
    background: #ef4444;
    border: 2px solid white;
    border-radius: 50%;
  }

  .notification-details {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .notification-title {
    font-weight: 600;
    color: #1f2937;
  }

  .notification-message {
    color: #475569;
    font-size: 0.9rem;
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
    min-width: 120px;
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

  .bg-blue-100 {
    background: #eff6ff;
    border-color: #bfdbfe;
  }
  .text-blue-700 {
    color: #1d4ed8;
  }
  .bg-slate-100 {
    background: #f1f5f9;
    border-color: #e2e8f0;
  }
  .text-slate-600 {
    color: #475569;
  }
  .equal-badge {
    min-width: 80px;
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
      gap: 12px;
    }
    .entry-card {
      padding: 14px;
      border-radius: 16px;
      border: 1px solid #e2e8f0;
      background: #ffffff;
      box-shadow: 0 6px 20px -18px rgba(15, 23, 42, 0.3);
    }
    .entry-card.unread-card {
      background: #eff6ff;
      border-color: #bfdbfe;
    }
    .entry-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      cursor: pointer;
    }
    .entry-head .notification-details {
      display: flex;
      flex-direction: column;
      min-width: 0;
    }
    .entry-head .notification-title {
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
    .entry-head .notification-icon-wrapper {
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

    .entry-head .notification-info {
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
    .detail-value.message-value {
      font-weight: 400;
      color: #475569;
      line-height: 1.5;
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
    .mini-btn .btn-text {
      display: inline;
    }

    .mini-btn.danger {
      border-color: #f87171;
      background: #f87171;
    }
    .mini-btn.danger .btn-text {
      color: #fff;
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
