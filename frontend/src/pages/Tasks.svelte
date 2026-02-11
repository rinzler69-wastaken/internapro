<script>
  import { onMount } from "svelte";
  import { goto, route } from "@mateothegreat/svelte5-router";
  import { api } from "../lib/api.js";
  import { auth } from "../lib/auth.svelte.js";

  import TaskCreateModal from "./TaskCreateModal.svelte";
  import TaskEditModal from "./TaskEditModal.svelte";

  // State
  let tasks = $state([]);
  let pagination = $state({ page: 1, total_pages: 1 });
  let search = $state("");
  let status = $state("");
  let priority = $state("");
  let internId = $state("");
  let interns = $state([]);
  let loading = $state(true);
  let exporting = $state(false);
  let isModalOpen = $state(false);
  let isEditModalOpen = $state(false);
  let editingTaskId = $state(null);

  const statusLabels = {
    pending: "Pending",
    scheduled: "Terjadwal",
    in_progress: "Dalam Proses",
    submitted: "Menunggu Review",
    revision: "Revisi",
    completed: "Selesai",
  };

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

  function formatSubmitted(task) {
    return formatDate(task?.submitted_at || task?.completed_at);
  }

  function getPriorityColor(p) {
    switch (p) {
      case "high":
        return "tone-rose"; // red
      case "medium":
        return "tone-amber"; // yellow
      default:
        return "tone-emerald"; // green
    }
  }

  function rowTint(status) {
    switch (status) {
      case "completed":
        return "bg-emerald-50";
      case "in_progress":
        return "bg-amber-50";
      case "revision":
        return "bg-rose-50";
      case "submitted":
        return "bg-blue-50";
      default:
        return "";
    }
  }

  const statusBadge = (status) => {
    switch (status) {
      case "completed":
        return {
          text: "Selesai",
          cls: "bg-emerald-100 text-emerald-700 border-emerald-200",
        };
      case "submitted":
        return {
          text: "Menunggu Review",
          cls: "bg-blue-100 text-blue-700 border-blue-200",
        };
      case "in_progress":
        return {
          text: "Dalam Proses",
          cls: "bg-amber-100 text-amber-700 border-amber-200",
        };
      case "revision":
        return {
          text: "Revisi",
          cls: "bg-rose-100 text-rose-700 border-rose-200",
        };
      case "scheduled":
        return {
          text: "Terjadwal",
          cls: "bg-indigo-100 text-indigo-700 border-indigo-200",
        };
      default:
        return {
          text: statusLabels[status] || status || "Pending",
          cls: "bg-slate-100 text-slate-600 border-slate-200",
        };
    }
  };

  // Helper function to get status color class for pills
  function getStatusColor(s) {
    switch (s) {
      case "completed":
        return "tone-emerald";
      case "submitted":
        return "tone-blue";
      case "in_progress":
        return "tone-amber";
      case "revision":
        return "tone-rose";
      default:
        return "tone-slate";
    }
  }

  async function fetchTasks() {
    loading = true;
    try {
      const params = { page: pagination.page || 1, limit: 15 };
      if (search) params.search = search;
      if (status) params.status = status;
      if (priority) params.priority = priority;
      if (internId) params.intern_id = internId;

      const res = await api.getTasks(params);
      tasks = res.data || [];
      pagination = res.pagination || { page: 1, total_pages: 1 };
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  async function fetchInterns() {
    if (auth.user?.role === "intern") return;
    try {
      const res = await api.getInterns({ status: "active", limit: 100 });
      interns = res.data || [];
    } catch (err) {
      console.error(err);
    }
  }

  async function exportTasks() {
    exporting = true;
    try {
      const params = new URLSearchParams();
      if (status) params.append("status", status);
      if (internId) params.append("intern_id", internId);
      const res = await fetch(
        `/api/export/tasks${params.toString() ? `?${params.toString()}` : ""}`,
        {
          headers: auth.token ? { Authorization: `Bearer ${auth.token}` } : {},
        },
      );
      if (!res.ok) throw new Error("Gagal mengekspor tugas");
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `Data_Tugas_${new Date().toISOString().slice(0, 10)}.xlsx`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } finally {
      exporting = false;
    }
  }

  async function handleDelete(id, title) {
    if (!confirm(`Apakah Anda yakin ingin menghapus tugas "${title}"?`)) return;
    try {
      await api.deleteTask(id);
      tasks = tasks.filter((t) => t.id !== id);
      alert("Tugas berhasil dihapus.");
    } catch (err) {
      console.error(err);
      alert("Gagal menghapus tugas: " + err.message);
    }
  }

  function setPage(p) {
    if (!pagination.total_pages) return;
    const target = Math.min(Math.max(1, p), pagination.total_pages);
    if (target !== pagination.page) {
      pagination = { ...pagination, page: target };
      fetchTasks();
    }
  }

  onMount(() => {
    fetchTasks();
    fetchInterns();
  });
</script>

<div class="page-shell">
  <div class="page-header">
    <div class="page-title">
      <h1>Daftar Penugasan</h1>
      <!-- <p class="muted">
        pantau, review, dan kelola tugas magang dengan ringkas.
      </p> -->
    </div>
    {#if auth.user?.role !== "intern"}
      <div class="page-actions">
        <button class="ghost" onclick={exportTasks} disabled={exporting}>
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            ><path d="M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-2"
            ></path><polyline points="7 11 12 16 17 11"></polyline><line
              x1="12"
              y1="4"
              x2="12"
              y2="16"
            ></line></svg
          >
          {exporting ? "Menyiapkan..." : "Export"}
        </button>
        <button class="primary" onclick={() => (isModalOpen = true)}>
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            ><line x1="12" y1="5" x2="12" y2="19"></line><line
              x1="5"
              y1="12"
              x2="19"
              y2="12"
            ></line></svg
          >
          Buat Tugas
        </button>
      </div>
    {/if}
  </div>

  <section class="card mb-6">
    <div class="card-body">
      <div class="filters">
        <div class="field stretch">
          <label for="searchTasks">Cari</label>
          <div class="input-icon">
            <svg
              class="search-icon"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <circle cx="11" cy="11" r="8" /><line
                x1="21"
                y1="21"
                x2="16.65"
                y2="16.65"
              />
            </svg>
            <input
              id="searchTasks"
              placeholder="Judul tugas..."
              bind:value={search}
              onkeydown={(e) => e.key === "Enter" && fetchTasks()}
            />
          </div>
        </div>
        <div class="field">
          <label for="statusFilter">Status</label>
          <select id="statusFilter" bind:value={status}>
            <option value="">Semua</option>
            <option value="pending">Pending</option>
            <option value="in_progress">Dalam Proses</option>
            <option value="submitted">Menunggu Review</option>
            <option value="revision">Revisi</option>
            <option value="completed">Selesai</option>
          </select>
        </div>
        <div class="field">
          <label for="priorityFilter">Prioritas</label>
          <select id="priorityFilter" bind:value={priority}>
            <option value="">Semua</option>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
          </select>
        </div>
        {#if auth.user?.role !== "intern"}
          <div class="field">
            <label for="internFilter">Intern</label>
            <select id="internFilter" bind:value={internId}>
              <option value="">Semua</option>
              {#each interns as intern}
                <option value={intern.id}
                  >{intern.full_name || intern.name}</option
                >
              {/each}
            </select>
          </div>
        {/if}
        <div class="field actions">
          <button
            class="circle-btn primary-circle"
            onclick={fetchTasks}
            title="Terapkan Filter"
          >
            <span class="material-symbols-outlined">filter_alt</span>
            <span class="btn-text-mobile">Terapkan</span>
          </button>
          <button
            class="circle-btn ghost-circle"
            title="Reset Filter"
            onclick={() => {
              search = "";
              status = "";
              priority = "";
              internId = "";
              fetchTasks();
            }}
          >
            <span class="material-symbols-outlined">restart_alt</span>
            <span class="btn-text-mobile">Reset</span>
          </button>
        </div>
      </div>
    </div>
  </section>

  <div class="card list-card animate-slide-up" style="animation-delay: 0.1s;">
    <div
      class="card-header border-b"
      style="padding: 4px 0px; justify-content: flex-start; gap: 12px; margin-bottom: 0;"
    >
      <h3>Daftar Tugas</h3>
      <span class="badge-count">{tasks.length} Item</span>
    </div>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Memuat daftar tugas...</p>
      </div>
    {:else if tasks.length === 0}
      <div class="empty-state">
        <div class="empty">📋</div>
        <p>Tidak ada tugas ditemukan.</p>
        {#if auth.user?.role !== "intern"}
          <p class="muted">Coba ubah filter atau buat tugas baru.</p>
        {/if}
      </div>
    {:else}
      <div class="table-container desktop-only">
        <table class="table desktop-table">
          <thead>
            <tr>
              <th>Judul</th>
              {#if auth.user?.role !== "intern"}<th>Intern</th>{/if}
              <th>Status</th>
              <th>Prioritas</th>
              <th>Deadline</th>
              <th>Dikumpulkan</th>
              <th>Aksi</th>
            </tr>
          </thead>
          <tbody>
            {#each tasks as task}
              {@const badge = statusBadge(task.status)}
              <tr
                class={`hover-row ${rowTint(task.status)}`}
                onclick={() => goto(`/tasks/${task.id}`)}
              >
                <td class="font-medium">
                  <div class="flex flex-col gap-1">
                    <span class="font-bold text-slate-800">{task.title}</span>
                    <span class="text-xs text-slate-500 truncate max-w-[200px]"
                      >{task.description || "—"}</span
                    >
                  </div>
                </td>
                {#if auth.user?.role !== "intern"}
                  <td>
                    <div class="user-info">
                      <div class="avatar-mini">
                        {task.intern_name?.charAt(0) || "U"}
                      </div>
                      <span>{task.intern_name || "-"}</span>
                    </div>
                  </td>
                {/if}
                <td>
                  <span class={`status-badge equal-badge ${badge.cls}`}>
                    {badge.text}
                  </span>
                  {#if task.is_late && task.status !== "completed"}
                    <span class="chip danger ml-2">Lewat</span>
                  {/if}
                </td>
                <td>
                  <span
                    class={`pill priority ${getPriorityColor(task.priority)}`}
                  >
                    <span class="dot"></span>
                    {task.priority === "medium"
                      ? "Medium"
                      : task.priority || "-"}
                  </span>
                </td>
                <td
                  class={`mono ${task.is_late && task.deadline ? "late" : ""}`}
                >
                  {formatDate(task.deadline)}
                </td>
                <td class="mono">{formatSubmitted(task)}</td>
                <td class="action-cell">
                  {#if auth.user?.role !== "intern"}
                    <button
                      class="btn-icon text-amber-600 hover:text-amber-700 bg-amber-50 hover:bg-amber-100"
                      onclick={(e) => {
                        e.stopPropagation();
                        editingTaskId = task.id;
                        isEditModalOpen = true;
                      }}
                      title="Edit Tugas"
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
                      class="btn-icon text-rose-600 hover:text-rose-700 bg-rose-50 hover:bg-rose-100"
                      onclick={(e) => {
                        e.stopPropagation();
                        handleDelete(task.id, task.title);
                      }}
                      title="Hapus Tugas"
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
                  {/if}
                  <button
                    class="btn-icon text-sky-600 hover:text-sky-700 bg-sky-50 hover:bg-sky-100"
                    title="Lihat"
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
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      <div class="mobile-list">
        {#each tasks as task}
          {@const badge = statusBadge(task.status)}
          <div
            class={`entry-card ${rowTint(task.status)}`}
            role="button"
            tabindex="0"
            onclick={() => goto(`/tasks/${task.id}`)}
            onkeydown={(e) => e.key === "Enter" && goto(`/tasks/${task.id}`)}
          >
            <div class="entry-head">
              <div class="title-row">
                <span class="material-symbols-outlined task-icon"
                  >assignment</span
                >
                <span class="task-title">{task.title}</span>
              </div>
              <span class={`status-badge equal-badge ${badge.cls}`}
                >{badge.text}</span
              >
            </div>

            {#if auth.user?.role !== "intern"}
              <div class="intern-grid">
                <div class="intern-box">
                  <div class="avatar-mini">
                    {task.intern_name?.charAt(0) || "U"}
                  </div>
                  <span class="intern-box-label">{task.intern_name || "-"}</span
                  >
                </div>
              </div>
            {/if}

            <div class="info-grid mt-3">
              <div class="info-box">
                <p class="label">Priority</p>
                <span
                  class={`pill priority small ${getPriorityColor(task.priority)}`}
                >
                  <span class="dot"></span>{task.priority}
                </span>
              </div>
              <div class="info-box">
                <p class="label">Deadline</p>
                <p
                  class={`value ${task.is_late ? "text-red-600 font-bold" : ""}`}
                >
                  {formatDate(task.deadline)}
                </p>
              </div>
            </div>

            <div class="mobile-actions mt-3">
              {#if auth.user?.role !== "intern"}
                <button
                  class="mini-btn mobile"
                  onclick={(e) => {
                    e.stopPropagation();
                    editingTaskId = task.id;
                    isEditModalOpen = true;
                  }}
                >
                  <span class="material-symbols-outlined">edit</span>
                  <span class="btn-text">Edit</span>
                </button>
                <button
                  class="mini-btn mobile danger"
                  onclick={(e) => {
                    e.stopPropagation();
                    handleDelete(task.id, task.title);
                  }}
                >
                  <span class="material-symbols-outlined">delete</span>
                  <span class="btn-text">Hapus</span>
                </button>
              {/if}
              <button class="mini-btn mobile">
                <span class="material-symbols-outlined">visibility</span>
                <span class="btn-text">Detail</span>
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  {#if pagination.total_pages > 1}
    <div class="pager">
      <button
        class="ghost"
        onclick={() => setPage((pagination.page || 1) - 1)}
        disabled={(pagination.page || 1) <= 1}>‹ Prev</button
      >
      <span class="muted"
        >Halaman {pagination.page || 1} dari {pagination.total_pages}</span
      >
      <button
        class="ghost"
        onclick={() => setPage((pagination.page || 1) + 1)}
        disabled={(pagination.page || 1) >= pagination.total_pages}
        >Next ›</button
      >
    </div>
  {/if}
</div>

<TaskCreateModal
  isOpen={isModalOpen}
  onClose={() => (isModalOpen = false)}
  onSuccess={() => {
    isModalOpen = false;
    fetchTasks();
  }}
/>

<TaskEditModal
  isOpen={isEditModalOpen}
  taskId={editingTaskId}
  onClose={() => {
    isEditModalOpen = false;
    editingTaskId = null;
  }}
  onSuccess={() => {
    isEditModalOpen = false;
    editingTaskId = null;
    fetchTasks();
  }}
/>

<style>
  /* html, */
  body {
    width: 100%;
    max-width: 100%;
    overflow-x: clip;
  }

  body {
    overscroll-behavior-x: none;
    -webkit-overflow-scrolling: touch;
  }

  * {
    max-width: 100%;
  }

  :global(body) {
    font-family: "Geist", "Inter", sans-serif;
    color: #0f172a;
    background: #f8fafc;
    overflow-x: hidden; /* Safety */
  }

  /* Match Attendance container style */
  .page-shell {
    margin: 0 auto;
    /* padding: 1rem;  */
    width: 100%;
    max-width: 1200px;
    overflow-x: clip;
    position: relative;
  }

  @media (min-width: 640px) {
    .page-shell {
      /* max-width: 640px; */
      padding: 1px;
    }
  }
  @media (min-width: 768px) {
    .page-shell {
      max-width: 768px;
    }
  }
  @media (min-width: 1024px) {
    .page-shell {
      max-width: 1024px;
    }
  }
  @media (min-width: 1280px) {
    .page-shell {
      max-width: 1200px; /* or whatever your other pages use */
    }
  }

  .page-header {
    display: flex;
    align-items: center; /* Center alignment */
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 24px;
    flex-wrap: wrap;
  }

  .page-title h1 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #1e293b;
    letter-spacing: -0.02em;
  }
  .muted {
    color: #64748b;
    margin: 0;
    font-size: 14px;
  }

  .page-actions {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  /* Mobile: Header stacks, Actions full width side-by-side */
  @media (max-width: 640px) {
    .page-header {
      flex-direction: column;
      align-items: stretch; /* Full width children */
      gap: 16px;
    }
    .page-title {
      text-align: left;
    }
    .page-actions {
      width: 100%;
      margin-top: 0; /* Handled by header gap */
      justify-content: space-between;
    }
    .page-actions button {
      flex: 1; /* Equal width */
      justify-content: center;
      width: 50%; /* Ensure they take space */
    }
  }

  /* Cards */
  .card {
    background: white;
    border-radius: 16px;
    border: 1px solid #e2e8f0;
    box-shadow:
      0 1px 3px 0 rgba(0, 0, 0, 0.1),
      0 1px 2px -1px rgba(0, 0, 0, 0.1);
    /* overflow: hidden; */
  }
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .card-header h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 700;
    color: #1e293b;
  }

  .badge-count {
    background: #f1f5f9;
    color: #64748b;
    font-size: 12px;
    font-weight: 600;
    padding: 4px 10px;
    border-radius: 99px;
  }
  .border-b {
    border-bottom: 1px solid #f1f5f9;
  }

  /* Filters */
  .filters {
    display: grid;
    grid-template-columns: minmax(0, 2fr) repeat(3, minmax(0, 1fr)) auto;
    gap: 16px;
    align-items: end;
  }
  @media (max-width: 900px) {
    .filters {
      grid-template-columns: 1fr;
      gap: 12px;
    }
  }
  .field.stretch {
    flex: 1;
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field.actions {
    justify-content: flex-end;
    flex-direction: row;
    gap: 8px;
    min-width: 90px;
  }

  @media (max-width: 900px) {
    .field.actions {
      width: 100%;
      margin-top: 4px;
    }
    .circle-btn {
      flex: 1;
      border-radius: 999px !important;
      width: auto !important;
      height: 40px !important;
      padding: 0 16px !important;
    }
    .btn-text-mobile {
      display: inline !important;
      margin-left: 8px;
    }
  }

  .field label {
    font-weight: 600;
    color: #475569;
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.02em;
  }
  .input-icon {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
  }
  .search-icon {
    position: absolute;
    left: 12px;
    color: #94a3b8;
  }
  .input-icon input {
    padding-left: 38px;
  }

  input,
  select {
    border: 1px solid #cbd5e1;
    border-radius: 10px;
    padding: 10px 14px;
    font-size: 14px;
    background: #fff;
    width: 100%;
    box-sizing: border-box;
    transition:
      border-color 0.15s,
      box-shadow 0.15s;
    height: 40px;
  }
  input:focus,
  select:focus {
    outline: none;
    border-color: #6366f1;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
  }

  /* Buttons */
  .primary,
  .secondary,
  .ghost {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    border-radius: 999px;
    font-weight: 600;
    cursor: pointer;
    text-decoration: none;
    border: 1px solid transparent;
    padding: 0 12px;
    transition: all 0.15s;
    font-size: 14px;
    height: 40px;
    box-sizing: border-box;
    white-space: nowrap;
  }
  .primary {
    background: linear-gradient(135deg, #10b981, #059669);
    color: white;
    box-shadow: 0 2px 4px rgba(16, 185, 129, 0.1);
    border: none;
  }
  .primary:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 6px rgba(16, 185, 129, 0.2);
  }

  .ghost {
    background: white;
    border-color: #e2e8f0;
    color: #475569;
  }
  .ghost:hover {
    background: #f8fafc;
    border-color: #cbd5e1;
    color: #0f172a;
  }
  .ghost:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  /* Circular Buttons */
  .circle-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: 999px;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.2s;
    padding: 0;
    color: white;
    flex-shrink: 0;
  }
  .primary-circle {
    background: #0f172a;
    color: white;
  }
  .primary-circle:hover {
    background: #1e293b;
    transform: scale(1.05);
  }

  .ghost-circle {
    background: white;
    border-color: #e2e8f0;
    color: #64748b;
  }
  .ghost-circle:hover {
    border-color: #ef4444;
    color: #ef4444;
    background: #fff5f5;
    transform: scale(1.05);
  }

  .btn-text-mobile {
    display: none;
    font-weight: 600;
    font-size: 13px;
  }

  /* Table */
  .table-container {
    width: 100%;
    max-width: 100%;
    overflow-x: auto;
  }
  .desktop-only {
    display: block;
  }
  .mobile-list {
    display: none;
  }
  .table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
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
  }

  .table th {
    text-align: left;
    padding: 14px 24px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    color: #64748b;
    background: #fcfcfc;
    border-bottom: 1px solid #e2e8f0;
  }
  .table td {
    padding: 16px 24px;
    border-bottom: 1px solid #f1f5f9;
    vertical-align: middle;
    font-size: 14px;
    color: #334155;
  }
  .hover-row:hover td {
    background-color: #f8fafc;
    cursor: pointer;
  }

  .action-cell {
    white-space: nowrap;
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .mini-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    border-radius: 999px;
    border: 1px solid #0f172a;
    background: #0f172a;
    color: #fff;
    font-weight: 600;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .mini-btn.icon-only {
    width: 34px;
    height: 34px;
    padding: 0;
    justify-content: center;
  }
  .mini-btn:hover {
    background: #111827;
    border-color: #111827;
  }

  .mini-btn.mobile {
    flex: 1;
    justify-content: center;
    padding: 8px;
    height: 40px;
  }
  .btn-text {
    display: none;
  }
  @media (max-width: 900px) {
    .mini-btn.mobile .btn-text {
      display: inline;
      font-weight: 700;
      font-size: 12px;
    }
  }

  /* User Info & Chips */
  .user-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .avatar-mini {
    width: 28px;
    height: 28px;
    border-radius: 999px;
    background: #0f172a;
    color: white;
    font-weight: 700;
    font-size: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  /* Badges */
  .status-badge {
    display: inline-flex;
    padding: 4px 10px;
    border-radius: 99px;
    font-size: 12px;
    font-weight: 600;
    border: 1px solid transparent;
    white-space: nowrap;
  }
  .chip {
    padding: 2px 6px;
    border-radius: 6px;
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
  }
  .chip.danger {
    background: #fef2f2;
    color: #b91c1c;
  }

  .pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    border-radius: 99px;
    font-weight: 600;
    font-size: 12px;
    border: 1px solid transparent;
  }
  .pill.priority {
    background: white;
    border: 1px dashed #cbd5e1;
    color: #475569;
  }
  .pill.small {
    padding: 2px 8px;
    font-size: 11px;
  }

  .tone-rose {
    color: #be123c;
    border-color: #fecdd3;
    background: #fff1f2;
  }
  .tone-amber {
    color: #b45309;
    border-color: #fde68a;
    background: #fffbeb;
  }
  .tone-emerald {
    color: #047857;
    border-color: #a7f3d0;
    background: #ecfdf3;
  }
  .tone-blue {
    color: #be123c;
    border-color: #bfdbfe;
    background: #eff6ff;
  }
  .tone-slate {
    color: #475569;
    border-color: #e2e8f0;
    background: #f1f5f9;
  }
  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
  }

  .mono {
    font-family: ui-monospace, monospace;
    color: #475569;
  }
  .mono.late {
    color: #b91c1c;
    font-weight: 700;
  }

  /* Mobile Entry Card */
  .entry-card {
    padding: 8px;
    /* border-radius: 16px;
    border: 1px solid #e2e8f0; */
    background: #ffffff;
    border-bottom: 1px solid #e2e8f0;
    padding-bottom: 24px;
    box-shadow: 0 6px 20px -18px rgba(15, 23, 42, 0.3);
    cursor: pointer;
    transition: transform 0.1s;
  }
  .entry-card:active {
    transform: scale(0.98);
  }

  .entry-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 12px;
  }
  .title-row {
    display: flex;
    align-items: center;
    gap: 8px;
    overflow: hidden;
    min-width: 0;
    flex: 1;
  }
  .task-icon {
    color: #6366f1;
    font-size: 20px;
    flex-shrink: 0;
  }
  .task-title {
    font-weight: 700;
    font-size: 15px;
    color: #0f172a;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .intern-grid {
    margin-bottom: 12px;
    display: flex;
  }
  .intern-box {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    background: #f8fafc;
    border-radius: 8px;
    border: 1px solid #f1f5f9;
  }
  .intern-box-label {
    font-size: 13px;
    font-weight: 600;
    color: #334155;
  }

  .info-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }
  .info-box {
    padding: 5px;
    background: #f8fafc;
    border: 1px solid #f1f5f9;
    border-radius: 10px;
    text-align: center;
  }
  .info-box .label {
    font-size: 10px;
    text-transform: uppercase;
    color: #94a3b8;
    font-weight: 700;
    margin-bottom: 4px;
  }
  .info-box .value {
    font-weight: 600;
    color: #334155;
    font-size: 13px;
    font-family: ui-monospace, monospace;
  }

  .mobile-actions {
    display: flex;
    gap: 8px;
  }

  /* Loading & Empty */
  .loading-state,
  .empty-state {
    padding: 40px;
    text-align: center;
    color: #94a3b8;
  }
  .spinner {
    width: 30px;
    height: 30px;
    border: 3px solid #e2e8f0;
    border-top-color: #6366f1;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 10px;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  .empty {
    font-size: 32px;
    margin-bottom: 10px;
  }

  .pager {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
    margin-top: 24px;
  }

  .page-header,
  .page-actions,
  .title-row,
  .user-info,
  .intern-box {
    min-width: 0;
  }

  /* Action Buttons - Standardized */
  .btn-icon {
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    padding: 6px;
    border-radius: 6px;
    transition: all 0.2s;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  .btn-icon:hover {
    background: #e2e8f0;
    color: #0f172a;
  }
</style>
