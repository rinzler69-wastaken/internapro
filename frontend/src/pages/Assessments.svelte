<script>
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import { api } from "../lib/api.js";
  import { portal } from "../lib/portal.js";
  import { auth } from "../lib/auth.svelte.js";
  import { getAvatarUrl } from "../lib/utils.js";
  import Pagination from "../components/Pagination.svelte";

  import AssessmentCreateModal from "./AssessmentCreateModal.svelte";
  import AssessmentEditModal from "./AssessmentEditModal.svelte";

  // State
  let assessments = $state([]);
  let loading = $state(false);
  let isCreateModalOpen = $state(false);
  let isEditModalOpen = $state(false);
  let selectedAssessment = $state(null);
  let searchQuery = $state("");
  let filterRole = $state("");
  let currentPage = $state(1);
  let totalPages = $state(1);
  let totalItems = $state(0);
  let searchTimeout;
  let expandedAssessments = $state({});

  // Helpers
  function calculateAverage(a) {
    const total =
      a.quality_score +
      a.speed_score +
      a.initiative_score +
      a.teamwork_score +
      a.communication_score;
    return (total / 5).toFixed(1);
  }

  function getGradeInfo(avg) {
    if (avg >= 85) return { grade: "A", class: "grade-a" };
    if (avg >= 75) return { grade: "B", class: "grade-b" };
    if (avg >= 60) return { grade: "C", class: "grade-c" };
    return { grade: "D", class: "grade-d" };
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
  async function fetchAssessments() {
    loading = true;
    try {
      const params = { page: currentPage, limit: 50 };
      if (searchQuery) params.search = searchQuery;
      if (filterRole) params.assessor_role = filterRole;

      const res = await api.getAssessments(params);
      assessments = res.data || [];
      const pagination = res.pagination || {};
      totalPages = Math.max(pagination.total_pages || 1, 1);
      totalItems = pagination.total_items || 0;
      currentPage = pagination.page || currentPage;
    } catch (err) {
      console.error("Failed to fetch assessments:", err);
      alert("Gagal memuat data penilaian: " + err.message);
    } finally {
      loading = false;
    }
  }

  function goToPreviousPage() {
    currentPage = Math.max(1, currentPage - 1);
    fetchAssessments();
  }

  function goToNextPage() {
    if (currentPage >= totalPages) return;
    currentPage += 1;
    fetchAssessments();
  }

  function handleSearchInput() {
    clearTimeout(searchTimeout);
    currentPage = 1;
    searchTimeout = setTimeout(() => {
      fetchAssessments();
    }, 500);
  }

  async function handleDelete(id, name) {
    if (
      !confirm(
        `Apakah Anda yakin ingin menghapus penilaian untuk "${name || "Intern"}"?`,
      )
    )
      return;
    try {
      await api.deleteAssessment(id);
      assessments = assessments.filter((a) => a.id !== id);
      alert("Penilaian berhasil dihapus.");
    } catch (err) {
      console.error(err);
      alert("Gagal menghapus penilaian: " + err.message);
    }
  }

  function openEditModal(assessment) {
    selectedAssessment = assessment;
    isEditModalOpen = true;
  }

  function toggleExpand(id) {
    expandedAssessments[id] = !expandedAssessments[id];
  }

  onMount(async () => {
    await fetchAssessments();
  });
</script>

<div class="page-container animate-fade-in">
  <div class="flex items-center gap-3 pb-8">
    <h4 class="card-title">Evaluasi Kinerja</h4>
    <span class="badge-count"
      >{assessments.length} dari {totalItems} Penilaian</span
    >
  </div>

  <!-- BAGIAN TABEL DAFTAR -->
  <div class="card table-card animate-slide-up" style="animation-delay: 0.1s;">
    <div class="card-header-row border-b">
      <div class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2">
        {#if auth.user?.role !== "intern"}
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
            <span>Buat Penilaian</span>
          </button>
        {/if}
        <button
          class="flex-none md:flex-none w-10 h-10 md:w-auto md:h-auto md:px-5 md:py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2"
          onclick={fetchAssessments}
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
          <span class="hidden md:inline">Refresh</span>
        </button>
      </div>
      <Pagination
        {currentPage}
        {totalPages}
        onPageChange={(page) => {
          currentPage = page;
          fetchAssessments();
        }}
      />
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
            (clearTimeout(searchTimeout), fetchAssessments())}
          placeholder="Cari Nama Intern atau Judul Tugas..."
          class="search-input"
        />
      </div>

      {#if auth.user?.role === "intern"}
        <select
          bind:value={filterRole}
          onchange={fetchAssessments}
          class="filter-select"
        >
          <option value="">Semua Penilai</option>
          <option value="supervisor">Pembimbing</option>
          <option value="admin">Admin</option>
        </select>
      {/if}
    </div>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Memuat data...</p>
      </div>
    {:else if assessments.length === 0}
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
          ><path d="M9 11l3 3L22 4"></path><path
            d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"
          ></path></svg
        >
        <p>Belum ada data penilaian.</p>
      </div>
    {:else}
      <div class="table-responsive desktop-only">
        <table class="table">
          <thead>
            <tr>
              <th>Peserta & Tugas</th>
              <th>Kualitas</th>
              <th>Kecepatan</th>
              <th>Inisiatif</th>
              <th class="text-center">Rata-rata</th>
              <th class="text-center">Grade</th>
              <th class="text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {#each assessments as item}
              {@const avg = Number(calculateAverage(item))}
              {@const gradeInfo = getGradeInfo(avg)}
              <tr class="table-row">
                <td style="min-width: 240px;">
                  <div class="assessment-info">
                    <div class="assessment-icon-wrapper">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
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
                    </div>
                    <div class="assessment-details">
                      <div class="user-info-inline">
                        {#if item.intern_avatar && getAvatarUrl(item.intern_avatar)}
                          <img
                            src={getAvatarUrl(item.intern_avatar)}
                            alt={item.intern_name}
                            class="w-6 h-6 rounded-full object-cover"
                          />
                        {:else}
                          <div class="avatar-small">
                            {item.intern_name?.charAt(0) || "U"}
                          </div>
                        {/if}
                        <span class="assessment-name"
                          >{item.intern_name || "Tidak Diketahui"}</span
                        >
                      </div>
                      <span class="task-subtitle"
                        >{item.task_title || "Penilaian Umum"}</span
                      >
                    </div>
                  </div>
                </td>
                <td>
                  <div class="bar-container">
                    <div class="progress-track-small">
                      <div
                        class="progress-bar-small bg-indigo-500"
                        style="width: {item.quality_score}%"
                      ></div>
                    </div>
                    <span class="score-text">{item.quality_score}</span>
                  </div>
                </td>
                <td>
                  <div class="bar-container">
                    <div class="progress-track-small">
                      <div
                        class="progress-bar-small bg-emerald-500"
                        style="width: {item.speed_score}%"
                      ></div>
                    </div>
                    <span class="score-text">{item.speed_score}</span>
                  </div>
                </td>
                <td>
                  <div class="bar-container">
                    <div class="progress-track-small">
                      <div
                        class="progress-bar-small bg-amber-500"
                        style="width: {item.initiative_score}%"
                      ></div>
                    </div>
                    <span class="score-text">{item.initiative_score}</span>
                  </div>
                </td>
                <td class="text-center">
                  <span class="avg-score">{avg}</span>
                </td>
                <td class="text-center">
                  <span class={`grade-badge ${gradeInfo.class}`}
                    >{gradeInfo.grade}</span
                  >
                </td>
                <td class="text-right">
                  <a
                    href={`/assessments/${item.id}`}
                    class="btn-icon text-sky-600 hover:text-sky-700 bg-sky-50 hover:bg-sky-100"
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
                  </a>
                  {#if auth.user && ["admin", "supervisor"].includes(auth.user.role)}
                    <button
                      class="btn-icon text-slate-600 hover:text-slate-700 bg-slate-50 hover:bg-slate-100"
                      onclick={(e) => {
                        e.stopPropagation();
                        openEditModal(item);
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
                  {/if}
                  {#if auth.user?.role !== "intern"}
                    <button
                      class="btn-icon text-slate-600 hover:text-slate-700 bg-slate-50 hover:bg-slate-100"
                      onclick={() => handleDelete(item.id, item.intern_name)}
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
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="mobile-list">
        {#each assessments as item}
          {@const avg = Number(calculateAverage(item))}
          {@const gradeInfo = getGradeInfo(avg)}
          <div class="entry-card">
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="entry-head" onclick={() => toggleExpand(item.id)}>
              <div class="assessment-info">
                <div class="assessment-icon-wrapper">
                  {#if item.intern_avatar && getAvatarUrl(item.intern_avatar)}
                    <img
                      src={getAvatarUrl(item.intern_avatar)}
                      alt={item.intern_name}
                      class="w-full h-full rounded-lg object-cover"
                    />
                  {:else}
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
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
                  {/if}
                </div>
                <div class="assessment-details">
                  <div class="assessment-name">
                    {item.intern_name || "Intern"}
                  </div>
                  <div class="text-muted small">
                    {item.task_title || "Penilaian Umum"}
                  </div>
                </div>
              </div>
              <button class="expand-btn">
                <span
                  class="material-symbols-outlined transition-transform duration-200 {expandedAssessments[
                    item.id
                  ]
                    ? 'rotate-180'
                    : ''}">expand_more</span
                >
              </button>
            </div>

            {#if expandedAssessments[item.id]}
              <div class="entry-details" transition:slide={{ duration: 200 }}>
                <div class="detail-row">
                  <div class="detail-label">PENILAI</div>
                  <div class="detail-value">
                    {item.assessor_name || "-"}
                    <span class="text-xs text-slate-500"
                      >({item.assessor_role})</span
                    >
                  </div>
                </div>
                <div class="detail-row">
                  <div class="detail-label">TANGGAL</div>
                  <div class="detail-value mono">
                    {formatDate(item.assessment_date || item.created_at)}
                  </div>
                </div>

                <!-- Bars for Mobile -->
                <div
                  class="detail-row"
                  style="display:block; margin-bottom: 8px;"
                >
                  <div
                    style="display:flex; justify-content:space-between; margin-bottom:6px; font-size:13px;"
                  >
                    <span class="text-slate-500 font-semibold">Kualitas</span>
                    <span class="font-bold text-slate-700"
                      >{item.quality_score}</span
                    >
                  </div>
                  <div class="progress-track-small" style="height: 6px;">
                    <div
                      class="progress-bar-small bg-indigo-500"
                      style="width: {item.quality_score}%"
                    ></div>
                  </div>
                </div>
                <div
                  class="detail-row"
                  style="display:block; margin-bottom: 8px;"
                >
                  <div
                    style="display:flex; justify-content:space-between; margin-bottom:6px; font-size:13px;"
                  >
                    <span class="text-slate-500 font-semibold">Kecepatan</span>
                    <span class="font-bold text-slate-700"
                      >{item.speed_score}</span
                    >
                  </div>
                  <div class="progress-track-small" style="height: 6px;">
                    <div
                      class="progress-bar-small bg-emerald-500"
                      style="width: {item.speed_score}%"
                    ></div>
                  </div>
                </div>
                <div
                  class="detail-row"
                  style="display:block; margin-bottom: 8px;"
                >
                  <div
                    style="display:flex; justify-content:space-between; margin-bottom:6px; font-size:13px;"
                  >
                    <span class="text-slate-500 font-semibold">Inisiatif</span>
                    <span class="font-bold text-slate-700"
                      >{item.initiative_score}</span
                    >
                  </div>
                  <div class="progress-track-small" style="height: 6px;">
                    <div
                      class="progress-bar-small bg-amber-500"
                      style="width: {item.initiative_score}%"
                    ></div>
                  </div>
                </div>
                <div
                  class="detail-row"
                  style="display:block; margin-bottom: 8px;"
                >
                  <div
                    style="display:flex; justify-content:space-between; margin-bottom:6px; font-size:13px;"
                  >
                    <span class="text-slate-500 font-semibold">Teamwork</span>
                    <span class="font-bold text-slate-700"
                      >{item.teamwork_score}</span
                    >
                  </div>
                  <div class="progress-track-small" style="height: 6px;">
                    <div
                      class="progress-bar-small bg-blue-500"
                      style="width: {item.teamwork_score}%"
                    ></div>
                  </div>
                </div>
                <div
                  class="detail-row"
                  style="display:block; margin-bottom: 12px;"
                >
                  <div
                    style="display:flex; justify-content:space-between; margin-bottom:6px; font-size:13px;"
                  >
                    <span class="text-slate-500 font-semibold">Komunikasi</span>
                    <span class="font-bold text-slate-700"
                      >{item.communication_score}</span
                    >
                  </div>
                  <div class="progress-track-small" style="height: 6px;">
                    <div
                      class="progress-bar-small bg-purple-500"
                      style="width: {item.communication_score}%"
                    ></div>
                  </div>
                </div>

                <div class="detail-row">
                  <div class="detail-label">RATA-RATA</div>
                  <span class="avg-score">{avg}</span>
                </div>
                <div class="detail-row">
                  <div class="detail-label">GRADE</div>
                  <span class={`grade-badge ${gradeInfo.class}`}
                    >{gradeInfo.grade}</span
                  >
                </div>

                <div class="mobile-actions mt-4 pt-4 border-t border-slate-100">
                  <a
                    href={`/assessments/${item.id}`}
                    class="mini-btn mobile"
                    onclick={(e) => e.stopPropagation()}
                  >
                    <span class="material-symbols-outlined">visibility</span>
                    <span class="btn-text">Detail</span>
                  </a>
                  {#if auth.user && ["admin", "supervisor"].includes(auth.user.role)}
                    <button
                      class="mini-btn-circle mobile"
                      onclick={(e) => {
                        e.stopPropagation();
                        openEditModal(item);
                      }}
                    >
                      <span class="material-symbols-outlined">edit</span>
                    </button>
                  {/if}
                  {#if auth.user?.role !== "intern"}
                    <button
                      class="mini-btn mobile danger"
                      onclick={(e) => {
                        e.stopPropagation();
                        handleDelete(item.id, item.intern_name);
                      }}
                    >
                      <span class="material-symbols-outlined">delete</span>
                      <span class="btn-text">Hapus</span>
                    </button>
                  {/if}
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<AssessmentCreateModal
  isOpen={isCreateModalOpen}
  onClose={() => (isCreateModalOpen = false)}
  onSuccess={() => {
    isCreateModalOpen = false;
    fetchAssessments();
  }}
/>

<AssessmentEditModal
  isOpen={isEditModalOpen}
  assessment={selectedAssessment}
  onClose={() => {
    isEditModalOpen = false;
    selectedAssessment = null;
  }}
  onSuccess={() => {
    isEditModalOpen = false;
    selectedAssessment = null;
    fetchAssessments();
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

  .text-xs {
    font-size: 0.75rem;
  }

  .text-slate-500 {
    color: #64748b;
  }

  .text-slate-700 {
    color: #334155;
  }

  .text-slate-800 {
    color: #1e293b;
  }

  .font-semibold {
    font-weight: 600;
  }

  .font-bold {
    font-weight: 700;
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

  /* Assessment Info in Table */
  .assessment-info {
    display: grid;
    grid-template-columns: 32px 1fr;
    gap: 12px;
    align-items: center;
  }

  .assessment-icon-wrapper {
    width: 32px;
    height: 32px;
    background: #10b981;
    color: white;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 2px 4px rgba(16, 185, 129, 0.2);
  }

  .assessment-details {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  .user-info-inline {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .avatar-small {
    width: 24px;
    height: 24px;
    background: rgb(15 23 42);
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 0.7rem;
    flex-shrink: 0;
  }

  .assessment-name {
    font-weight: 600;
    color: #1f2937;
    font-size: 0.95rem;
  }

  .task-subtitle {
    font-size: 0.75rem;
    color: #6b7280;
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

  /* Progress Bars */
  .bar-container {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .progress-track-small {
    flex: 1;
    height: 6px;
    background: #f1f5f9;
    border-radius: 3px;
    overflow: hidden;
    min-width: 60px;
  }
  .progress-bar-small {
    height: 100%;
    border-radius: 3px;
  }
  .score-text {
    font-size: 12px;
    font-weight: 600;
    color: #64748b;
    width: 24px;
    text-align: right;
  }

  .bg-indigo-500 {
    background: #6366f1;
  }
  .bg-emerald-500 {
    background: #10b981;
  }
  .bg-amber-500 {
    background: #f59e0b;
  }
  .bg-blue-500 {
    background: #3b82f6;
  }
  .bg-purple-500 {
    background: #a855f7;
  }

  /* Score Display */
  .avg-score {
    font-weight: 800;
    font-size: 16px;
    color: #0f172a;
  }

  .grade-badge {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 14px;
    border: 2px solid transparent;
  }
  .grade-a {
    background: #f0fdf4;
    color: #15803d;
    border-color: #bbf7d0;
  }
  .grade-b {
    background: #eff6ff;
    color: #1d4ed8;
    border-color: #bfdbfe;
  }
  .grade-c {
    background: #fffbeb;
    color: #b45309;
    border-color: #fde68a;
  }
  .grade-d {
    background: #fff1f2;
    color: #be123c;
    border-color: #fecdd3;
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
    text-decoration: none;
  }

  /* Ensure color inheritance works for links */
  a.btn-icon {
    color: inherit;
  }

  a.btn-icon.text-sky-600 {
    color: #0284c7;
  }
  a.btn-icon.text-sky-600:hover {
    color: #0369a1;
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
    .entry-head .assessment-details {
      display: flex;
      flex-direction: column;
      min-width: 0;
    }
    .entry-head .assessment-name {
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
    .entry-head .assessment-icon-wrapper {
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

    .entry-head .assessment-info {
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
      text-decoration: none;
    }
    .mini-btn .btn-text {
      display: inline;
    }
    .mini-btn.danger {
      background: #ef4444;
      border-color: #ef4444;
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

    /* Ensure links don't have underlines */
    a.mini-btn {
      text-decoration: none;
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

  .mt-2 {
    margin-top: 0.5rem;
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
